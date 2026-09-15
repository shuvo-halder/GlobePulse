# GlobePulse AI — Production Deployment Guide
**Target Operating System:** Ubuntu Server 24.04 LTS (Noble Numbat)  
**Orchestration Engine:** Docker Engine & Docker Compose (v2)  
**Architecture:** Single-Host Production VPS with Nginx Reverse Proxy  
**Document Version:** 1.0.0 (Production Release)  
**Classification:** DevOps / Systems Engineering Playbook  

---

## Table of Contents

1. [System Overview & Architectural Blueprint](#1-system-overview--architectural-blueprint)
2. [Hardware Sizing & VPS Specifications](#2-hardware-sizing--vps-specifications)
3. [Domain Design & DNS Architecture](#3-domain-design--dns-architecture)
4. [VPS Base Installation & OS Hardening](#4-vps-base-installation--os-hardening)
5. [Required Software Installation (Ubuntu 24.04)](#5-required-software-installation-ubuntu-2404)
6. [Directory Structure & Repository Initialization](#6-directory-structure--repository-initialization)
7. [Production Environment Configuration](#7-production-environment-configuration)
8. [Database Architecture & Migration Management](#8-database-architecture--migration-management)
9. [Redis Session & Cache Management](#9-redis-session--cache-management)
10. [RabbitMQ Message Broker Configuration](#10-rabbitmq-message-broker-configuration)
11. [Docker Compose Production Deployment](#11-docker-compose-production-deployment)
12. [Startup Order & Service Health Verification](#12-startup-order--service-health-verification)
13. [Nginx Reverse Proxy & Edge Routing](#13-nginx-reverse-proxy--edge-routing)
14. [SSL/TLS Security via Let's Encrypt (Certbot)](#14-ssltls-security-via-lets-encrypt-certbot)
15. [CORS Policy & Cross-Origin Security](#15-cors-policy--cross-origin-security)
16. [Authentication, JWT & Secret Management](#16-authentication-jwt--secret-management)
17. [External Intelligence Source Connectors](#17-external-intelligence-source-connectors)
18. [Ingestion Scheduler & Pipeline Operations](#18-ingestion-scheduler--pipeline-operations)
19. [Production Logging & Log Rotation](#19-production-logging--log-rotation)
20. [Host & Application Health Monitoring](#20-host--application-health-monitoring)
21. [Automated Backup Strategy](#21-automated-backup-strategy)
22. [Disaster Recovery Procedure](#22-disaster-recovery-procedure)
23. [Zero-Downtime Reality & Application Update Procedure](#23-zero-downtime-reality--application-update-procedure)
24. [Network Firewall (UFW) & Port Exposure Matrix](#24-network-firewall-ufw--port-exposure-matrix)
25. [Troubleshooting & Diagnostic Playbook](#25-troubleshooting--diagnostic-playbook)
26. [Post-Deployment Verification Checklist](#26-post-deployment-verification-checklist)
27. [Production Command Quick Reference](#27-production-command-quick-reference)
28. [Production Security Audit Checklist](#28-production-security-audit-checklist)
29. [Remaining Manual Verification & Operational Notes](#29-remaining-manual-verification--operational-notes)

---

## 1. System Overview & Architectural Blueprint

GlobePulse AI is a real-time global threat intelligence platform. The system ingests event streams from global humanitarian, seismic, and news intelligence feeds, normalizes them into canonical threat events, enriches them with geographic and provenance metadata, extracts intelligence entities, and visualizes them on a 3D interface.

### Production Network Topology

```text
               Public Internet / End Users
                          │
                   (Ports 80, 443)
                          ▼
            ┌───────────────────────────┐
            │   Nginx (Host Layer)      │
            │   SSL/TLS Termination     │
            │   Security Headers & CORS │
            └─────────────┬─────────────┘
                          │
       ┌──────────────────┴──────────────────┐
       │ (127.0.0.1:3100)                    │ (127.0.0.1:8080 - 8084)
       ▼                                     ▼
┌──────────────┐                     ┌───────────────────────────────┐
│   frontend   │                     │      Backend Microservices    │
│  React/Nginx │                     │  (Go Gin / HTTP / Python AI)  │
└──────────────┘                     └───────────────┬───────────────┘
                                                     │
                       ┌─────────────────────────────┼─────────────────────────────┐
                       │ (Docker Internal: 5432)     │ (Docker Internal: 6379)     │ (Docker Internal: 5672)
                       ▼                             ▼                             ▼
              ┌─────────────────┐           ┌─────────────────┐           ┌─────────────────┐
              │    postgres     │           │      redis      │           │    rabbitmq     │
              │  PostgreSQL 15  │           │     Redis 7     │           │   RabbitMQ 3    │
              └─────────────────┘           └─────────────────┘           └─────────────────┘
                       ▲                                                           ▲
                       │                                                           │
          ┌────────────┴────────────┐                                 ┌────────────┴────────────┐
          │  news-service Pipeline  │                                 │     ai-worker Celery    │
          │  (Ingestion/Extraction) │                                 │     (Worker Process)    │
          └────────────┬────────────┘                                 └─────────────────────────┘
                       │ (Outbound HTTPS: 443)
                       ▼
          ┌─────────────────────────────────────────────────────────┐
          │               External Intelligence Feeds               │
          │    • GDELT 2.0 API    • USGS Earthquakes    • ReliefWeb │
          └─────────────────────────────────────────────────────────┘
```

### Component Breakdown

| Component | Technology | Role in Production | Internal Port | Host Exposure Policy |
| :--- | :--- | :--- | :--- | :--- |
| **Edge Proxy** | Nginx (Ubuntu Host) | SSL termination, reverse proxy, CORS, static cache | 80, 443 | **Public (0.0.0.0)** |
| **frontend** | React 19 + Nginx Alpine | Web UI & 3D Globe visualization | 80 | `127.0.0.1:3100` (Private loopback only; port 3000 reserved for existing host workloads) |
| **auth-service** | Go 1.21 (Gin) | User registration, login, JWT issuance, Redis sessions | 8081 | `127.0.0.1:8081` (Private loopback) |
| **news-service** | Go 1.21 (HTTP) | Scheduled ingestion, deduplication, enrichment, entity extraction | 8080 | `127.0.0.1:8080` (Private loopback) |
| **country-service** | Go 1.21 (HTTP) | Country risk metrics & geospatial metadata | 8082 | `127.0.0.1:8082` (Private loopback) |
| **analytics-service**| Go 1.21 (HTTP) | Aggregated threat telemetry & analytics | 8084 | `127.0.0.1:8084` (Private loopback) |
| **ai-service** | Python 3.11 (FastAPI) | AI threat scoring & analysis dispatch | 8083 | `127.0.0.1:8083` (Private loopback) |
| **ai-worker** | Python 3.11 (Celery) | Asynchronous task worker for `ai_analysis_queue` | None | None (Internal Worker) |
| **postgres** | PostgreSQL 15 Alpine | Primary relational datastore (Threats, Sources, Users, Entities) | 5432 | **None** (Internal Docker network only; no host port published) |
| **redis** | Redis 7 Alpine | Auth session cache & Celery task result backend | 6379 | **None** (Internal Docker network only; no host port published) |
| **rabbitmq** | RabbitMQ 3 Management | AMQP Message broker for Celery & async queues | 5672, 15672 | AMQP: **None** (Internal only); Web UI: `127.0.0.1:15672` (SSH Tunnel only) |

### 1.1 Multi-Tenant VPS Coexistence & Isolation Architecture

The target Ubuntu 24.04 LTS host already hosts running production containers that **must remain completely untouched and unperturbed**:
* `n8n`: Workflow automation engine
* `n8n-postgres`: PostgreSQL database dedicated to n8n (listens internally on 5432 without host publication)
* `go-whatsapp-web-multidevice-whatsapp_go-1`: WhatsApp integration service listening on host port `0.0.0.0:3000`

To ensure 100% collision-free coexistence and zero security bleed:
1. **Host Port Conflict Prevention:** The GlobePulse frontend is bound to host loopback port `127.0.0.1:3100:80` instead of `3000`, leaving host port 3000 completely free for `whatsapp_go`.
2. **Database Isolation:** GlobePulse runs its own dedicated `postgres:15-alpine` container with its own named Docker volume `pgdata`. It does **not** publish port 5432 to the host, preventing conflicts with `n8n-postgres` or any future PostgreSQL workloads.
3. **Redis & RabbitMQ AMQP Isolation:** GlobePulse `redis` (port 6379) and `rabbitmq` (port 5672) publish zero host ports. They are reachable solely across the private Docker network by GlobePulse containers.
4. **Dedicated Compose Network (`globepulse-net`):** All GlobePulse containers attach to an isolated bridge network named `globepulse-net`, completely decoupling DNS resolution and network namespaces from `n8n` or default Docker bridges.
5. **Scoped Operational Boundary:** All Docker operations must be executed strictly within `/opt/globepulse/app/` using project-scoped commands (`docker compose`). Global destructive commands (e.g. `docker compose down` in other directories, `docker system prune -a --volumes`) are strictly prohibited.

---

## 2. Hardware Sizing & VPS Specifications

GlobePulse AI runs 10 discrete containers simultaneously: 5 compiled Go services, 2 Python processes (FastAPI + Celery), PostgreSQL, Redis, and RabbitMQ. Memory consumption is dominated by the JVM-like Erlang VM (RabbitMQ), Python Celery runtime, and PostgreSQL buffer pools.

### Recommended Specifications

| Resource | Minimum (Staging / Low Ingestion) | Recommended Production Starting Point |
| :--- | :--- | :--- |
| **Operating System** | Ubuntu Server 24.04 LTS (x86_64) | Ubuntu Server 24.04 LTS (x86_64) |
| **vCPU Cores** | 2 vCPU | 4 vCPU |
| **RAM** | 4 GB | 8 GB |
| **Disk Storage** | 40 GB NVMe / SSD | 80 GB+ NVMe SSD |
| **Swap Space** | 2 GB Swap file | 4 GB Swap file |
| **Bandwidth** | 100 Mbps (1 TB / month egress) | 1 Gbps (3 TB+ / month egress) |

> **Swap Requirement:** Even on an 8 GB VPS, configuring a 4 GB swap space is **mandatory**. Transient spikes during image builds (`npm run build` with Vite or Celery memory spikes) can trigger the Linux Out-Of-Memory (OOM) killer without swap.

---

## 3. Domain Design & DNS Architecture

The platform architecture splits the client-side user interface from the API microservices to maintain clean separation of concerns and allow granular edge caching.

### Recommended Production DNS Records

Configure the following DNS records with your registrar or DNS provider (e.g., Cloudflare, Route53, Namecheap):

| Hostname | Type | Target IP | Description |
| :--- | :--- | :--- | :--- |
| `app.example.com` | `A` | `YOUR_SERVER_IP` | Resolves to Frontend Web Interface |
| `api.example.com` | `A` | `YOUR_SERVER_IP` | Resolves to Microservices Gateway (Nginx) |
| `app.example.com` | `AAAA` | `YOUR_SERVER_IPV6` *(Optional)* | Native IPv6 support |
| `api.example.com` | `AAAA` | `YOUR_SERVER_IPV6` *(Optional)* | Native IPv6 support |

### Cloudflare Proxy Considerations
If using Cloudflare:
- During initial Certbot verification, set proxy status to **DNS Only (Grey Cloud)** to allow the ACME challenge to complete over port 80.
- Once SSL certificates are installed, you may enable Cloudflare **Proxied (Orange Cloud)**. Ensure the Cloudflare SSL/TLS encryption mode is set to **Full (Strict)** to enforce end-to-end TLS between Cloudflare and the VPS.

---

## 4. VPS Base Installation & OS Hardening

Perform these steps on a newly provisioned Ubuntu 24.04 LTS server.

### 4.1 System Update & Base Timezone
Log in via SSH as `root`:
```bash
ssh root@YOUR_SERVER_IP
```

Set the timezone to UTC and update package repositories:
```bash
timedatectl set-timezone UTC
hostnamectl set-hostname globepulse-prod

apt-get update && apt-get upgrade -y
apt-get install -y curl wget git ufw fail2ban unattended-upgrades ca-certificates apt-transport-https htop net-tools
```

### 4.2 Configure Swap Space (4 GB)
```bash
fallocate -l 4G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile

# Persist swap across reboots
echo '/swapfile none swap sw 0 0' >> /etc/fstab

# Tune swappiness for database performance
sysctl vm.swappiness=10
sysctl vm.vfs_cache_pressure=50
echo 'vm.swappiness=10' >> /etc/sysctl.conf
echo 'vm.vfs_cache_pressure=50' >> /etc/sysctl.conf
```

### 4.3 Create Dedicated Non-Root Deployment User
Do **not** run production applications as `root`. Create a dedicated user named `deploy`:
```bash
adduser --gecos "" deploy
usermod -aG sudo deploy

# Authorize your local SSH public key for deploy user
mkdir -p /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
cp /root/.ssh/authorized_keys /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
```

### 4.4 Safe SSH Hardening Sequence
> **CRITICAL:** Do NOT close your existing root session until you have verified SSH access in a new terminal window!

Edit `/etc/ssh/sshd_config.d/50-cloud-init.conf` or `/etc/ssh/sshd_config`:
```bash
cat << 'EOF' > /etc/ssh/sshd_config.d/99-globepulse-hardening.conf
PermitRootLogin prohibit-password
PasswordAuthentication no
PubkeyAuthentication yes
X11Forwarding no
MaxAuthTries 4
ClientAliveInterval 300
ClientAliveCountMax 2
EOF

systemctl restart ssh
```

*Verification test from your local workstation:*
```bash
ssh -i ~/.ssh/id_rsa deploy@YOUR_SERVER_IP
```
Ensure you obtain a shell without being prompted for a password before disconnecting the root session.

### 4.5 Configure UFW (Uncomplicated Firewall)
Default deny all inbound, allow outbound, and only permit SSH, HTTP, and HTTPS:
```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp comment "SSH Management"
ufw allow 80/tcp comment "HTTP (Certbot / Redirect)"
ufw allow 443/tcp comment "HTTPS Production Traffic"

# Enable UFW (type 'y' to confirm)
ufw enable
ufw status verbose
```

### 4.6 Configure Fail2ban
Protect SSH against brute-force attacks:
```bash
cat << 'EOF' > /etc/fail2ban/jail.local
[DEFAULT]
bantime = 1h
findtime = 10m
maxretry = 5

[sshd]
enabled = true
port = 22
mode = aggressive
EOF

systemctl restart fail2ban
systemctl enable fail2ban
```

### 4.7 Automatic Security Updates
Enable Ubuntu's unattended upgrades for automated security patches:
```bash
dpkg-reconfigure -plow unattended-upgrades
```

---

## 5. Required Software Installation (Ubuntu 24.04)

Deploy official Docker packages for Ubuntu 24.04 (Codename: `noble`) and host Nginx.

### 5.1 Install Docker Engine & Docker Compose Plugin
```bash
# Remove any conflicting packages
for pkg in docker.io docker-doc docker-compose podman-docker containerd runc; do apt-get remove -y $pkg; done

# Install Docker GPG key
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

# Add official Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  noble stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine and Compose
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Grant deploy user permission to manage Docker
usermod -aG docker deploy
```

### 5.2 Install Nginx & Certbot
```bash
apt-get install -y nginx certbot python3-certbot-nginx
systemctl enable nginx
systemctl start nginx
```

### 5.3 Verify Versions
Switch to the `deploy` user or re-login so group changes take effect:
```bash
su - deploy
docker --version
docker compose version
nginx -v
certbot --version
```
Expected output:
- Docker Engine >= 26.x
- Docker Compose >= v2.27.x
- Nginx >= 1.24.x
- Certbot >= 2.9.x

---

## 6. Directory Structure & Repository Initialization

Follow standard Linux filesystem conventions (`/opt` for self-contained third-party software packages).

### 6.1 Create Production Directories
Execute as `root` or `sudo`:
```bash
sudo mkdir -p /opt/globepulse/app
sudo mkdir -p /opt/globepulse/backups/postgres
sudo mkdir -p /opt/globepulse/logs/nginx
sudo mkdir -p /opt/globepulse/scripts

sudo chown -R deploy:deploy /opt/globepulse
sudo chmod -R 750 /opt/globepulse
```

### 6.2 Clone Repository
As user `deploy`:
```bash
cd /opt/globepulse
git clone https://github.com/shuvo-halder/GlobePulse.git app
cd /opt/globepulse/app
```

---

## 7. Production Environment Configuration

The repository services require runtime configuration for database access, Redis connectivity, message queue routing, and JWT cryptographic operations.

### 7.1 Master Environment Variable Catalog

| Variable | Scope | Purpose | Required | Production Example / Format | Classification |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `APP_ENV` | Global | Runtime environment mode | Yes | `production` | Public |
| `POSTGRES_DB` | Database | PostgreSQL database name | Yes | `globepulse` | Public |
| `POSTGRES_USER` | Database | PostgreSQL master user | Yes | `gp_admin` | Secret |
| `POSTGRES_PASSWORD` | Database | PostgreSQL master password | Yes | 32+ character random string | **Strict Secret** |
| `DB_HOST` | Go / Python | PostgreSQL container hostname | Yes | `postgres` | Public |
| `DB_PORT` | Go / Python | PostgreSQL container port | Yes | `5432` | Public |
| `DB_USER` | Go / Python | PostgreSQL application user | Yes | `gp_admin` | Secret |
| `DB_PASS` | Go / Python | PostgreSQL application password | Yes | Matches `POSTGRES_PASSWORD` | **Strict Secret** |
| `DB_NAME` | Go / Python | PostgreSQL application database | Yes | `globepulse` | Public |
| `REDIS_ADDR` | Go / Python | Redis host (or host:port for Go) | Yes | `redis:6379` (Go), `redis` (Py) | Public |
| `REDIS_PORT` | Python | Redis port parameter | Yes | `6379` | Public |
| `REDIS_PASS` | Go / Python | Redis authentication password | No | `""` or custom password | Secret |
| `JWT_SECRET` | auth-service | HMAC-SHA256 signature key | Yes | 64+ hex string (`openssl rand -hex 32`) | **Strict Secret** |
| `TOKEN_EXPIRATION_MINUTES` | auth-service | JWT validity duration | No | `60` (Default) | Public |
| `RABBITMQ_URL` | analytics, AI | AMQP connection string | Yes | `amqp://gp_rabbit:SECRET@rabbitmq:5672/` | **Strict Secret** |
| `RABBITMQ_DEFAULT_USER`| rabbitmq | RabbitMQ admin user | Yes | `gp_rabbit` | Secret |
| `RABBITMQ_DEFAULT_PASS`| rabbitmq | RabbitMQ admin password | Yes | 32+ character random string | **Strict Secret** |
| `INGESTION_INTERVAL`| news-service | Polling interval for feeds | No | `10m` (Default 10 minutes) | Public |

### 7.2 Secure Secret Generation
Generate high-entropy credentials directly on the VPS:
```bash
# Generate unique cryptographic tokens
DB_PASSWORD=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-20)
RABBIT_PASSWORD=$(openssl rand -base64 24 | tr -d "=+/" | cut -c1-20)
JWT_SECRET_KEY=$(openssl rand -hex 32)
```

### 7.3 Production Docker Compose Override File
The standard `docker-compose.yml` provides baseline development defaults and loopback bindings. In production, we deploy with a `docker-compose.override.yml` file to apply production environment variables, restart policies, and persistent broker storage without exposing internal datastores to the host.

Notice that:
* `postgres`, `redis`, and RabbitMQ AMQP publish **no host ports**; they communicate strictly over the internal `globepulse-net` network, ensuring complete isolation from existing workloads such as `n8n-postgres`.
* `frontend` binds strictly to loopback `127.0.0.1:3100:80`, preserving host port 3000 for the existing `go-whatsapp-web-multidevice-whatsapp_go-1` container.
* All microservice API ports bind strictly to `127.0.0.1` for exclusive reverse proxying by the host Nginx.

Create `/opt/globepulse/app/docker-compose.override.yml`:

```yaml
version: '3.8'

services:
  postgres:
    environment:
      POSTGRES_USER: gp_admin
      POSTGRES_PASSWORD: ${DB_PASS}
      POSTGRES_DB: globepulse
    restart: unless-stopped

  redis:
    restart: unless-stopped

  rabbitmq:
    environment:
      RABBITMQ_DEFAULT_USER: gp_rabbit
      RABBITMQ_DEFAULT_PASS: ${RABBIT_PASS}
    ports:
      - "127.0.0.1:15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    restart: unless-stopped

  auth-service:
    environment:
      APP_ENV: production
      PORT: 8081
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: gp_admin
      DB_PASS: ${DB_PASS}
      DB_NAME: globepulse
      REDIS_ADDR: redis:6379
      REDIS_PASS: ""
      JWT_SECRET: ${JWT_SECRET}
      TOKEN_EXPIRATION_MINUTES: 60
    ports:
      - "127.0.0.1:8081:8081"
    restart: unless-stopped

  news-service:
    environment:
      APP_ENV: production
      PORT: 8080
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: gp_admin
      DB_PASS: ${DB_PASS}
      DB_NAME: globepulse
      REDIS_ADDR: redis:6379
      REDIS_PASS: ""
      INGESTION_INTERVAL: 10m
    ports:
      - "127.0.0.1:8080:8080"
    restart: unless-stopped

  country-service:
    environment:
      APP_ENV: production
      PORT: 8082
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: gp_admin
      DB_PASS: ${DB_PASS}
      DB_NAME: globepulse
      REDIS_ADDR: redis:6379
      REDIS_PASS: ""
    ports:
      - "127.0.0.1:8082:8082"
    restart: unless-stopped

  analytics-service:
    environment:
      APP_ENV: production
      PORT: 8084
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: gp_admin
      DB_PASS: ${DB_PASS}
      DB_NAME: globepulse
      REDIS_ADDR: redis:6379
      REDIS_PASS: ""
      RABBITMQ_URL: amqp://gp_rabbit:${RABBIT_PASS}@rabbitmq:5672/
    ports:
      - "127.0.0.1:8084:8084"
    restart: unless-stopped

  ai-service:
    environment:
      APP_ENV: production
      PORT: 8083
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: gp_admin
      DB_PASS: ${DB_PASS}
      DB_NAME: globepulse
      REDIS_ADDR: redis
      REDIS_PORT: 6379
      REDIS_PASS: ""
      REDIS_URL: redis://redis:6379/0
      RABBITMQ_URL: amqp://gp_rabbit:${RABBIT_PASS}@rabbitmq:5672/
    ports:
      - "127.0.0.1:8083:8083"
    restart: unless-stopped

  ai-worker:
    environment:
      APP_ENV: production
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: gp_admin
      DB_PASS: ${DB_PASS}
      DB_NAME: globepulse
      REDIS_ADDR: redis
      REDIS_PORT: 6379
      REDIS_PASS: ""
      REDIS_URL: redis://redis:6379/0
      RABBITMQ_URL: amqp://gp_rabbit:${RABBIT_PASS}@rabbitmq:5672/
    restart: unless-stopped

  frontend:
    environment:
      APP_ENV: production
    ports:
      - "127.0.0.1:3100:80"
    restart: unless-stopped

volumes:
  rabbitmq_data:
```

### 7.4 Create Master Production `.env`
Create `/opt/globepulse/app/.env` (permissions strictly `600`):
```bash
cat << EOF > /opt/globepulse/app/.env
# GlobePulse AI Production Environment Secrets
DB_PASS=${DB_PASSWORD}
RABBIT_PASS=${RABBIT_PASSWORD}
JWT_SECRET=${JWT_SECRET_KEY}
EOF

chmod 600 /opt/globepulse/app/.env
chmod 600 /opt/globepulse/app/docker-compose.override.yml
```

---

## 8. Database Architecture & Migration Management

PostgreSQL 15 serves as the relational datastore for:
- User accounts and authentication audit trails (`auth-service`)
- Sources, raw items, threat events, provenance, and intelligence entities (`news-service`)

### 8.1 Critical Migration Sequence
The repository uses standard SQL migration scripts. The migrations **MUST** be executed sequentially against the PostgreSQL instance:

```text
Migration Step 1: services/auth-service/migrations/000001_create_users_table.up.sql
                  ├── Creates 'users' table (UUID, email, password_hash, role)
                  └── Creates 'audit_logs' table (action, IP, user_agent)

Migration Step 2: services/news-service/migrations/000001_create_threat_events_schema.up.sql
                  ├── Creates 'sources' table (id, name, source_type, base_url)
                  ├── Creates 'source_items' table (unique constraint: uq_source_external_id)
                  ├── Creates 'threat_events' table (Step 6 & 7 canonical schema, indices)
                  └── Creates 'threat_event_source_items' (many-to-many relationship)

Migration Step 3: services/news-service/migrations/000002_create_intelligence_entities_schema.up.sql
                  ├── Creates 'intelligence_entities' (unique constraint: uq_intelligence_entities_type_norm)
                  └── Creates 'threat_event_entities' (Step 8 global intelligence mapping)
```

> **WARNING — DESTRUCTIVE COMMANDS FORBIDDEN:**
> Never execute `dropdb`, `DROP TABLE`, or run `.down.sql` scripts in production. The deduplication foundation relies on foreign key constraints and `uq_source_external_id`. Dropping tables causes irreversible loss of intelligence provenance and entity relationships.

### 8.2 Applying Migrations Step-by-Step
Once the PostgreSQL container is running and healthy:

```bash
cd /opt/globepulse/app

# Wait for PostgreSQL ready state
docker compose exec -T postgres pg_isready -U gp_admin -d globepulse

# 1. Apply Auth Service Migration
docker compose exec -T postgres psql -U gp_admin -d globepulse < services/auth-service/migrations/000001_create_users_table.up.sql

# 2. Apply Threat Events Foundation Migration (Step 6 & Step 7)
docker compose exec -T postgres psql -U gp_admin -d globepulse < services/news-service/migrations/000001_create_threat_events_schema.up.sql

# 3. Apply Intelligence Entities Schema Migration (Step 8)
docker compose exec -T postgres psql -U gp_admin -d globepulse < services/news-service/migrations/000002_create_intelligence_entities_schema.up.sql
```

### 8.3 Verifying Schema Application
```bash
docker compose exec -T postgres psql -U gp_admin -d globepulse -c "\dt"
```
Expected table catalog:
```text
               List of relations
 Schema |           Name              | Type  |  Owner   
--------+-----------------------------+-------+----------
 public | audit_logs                  | table | gp_admin
 public | intelligence_entities       | table | gp_admin
 public | source_items                | table | gp_admin
 public | sources                     | table | gp_admin
 public | threat_event_entities       | table | gp_admin
 public | threat_event_source_items   | table | gp_admin
 public | threat_events               | table | gp_admin
 public | users                       | table | gp_admin
(8 rows)
```

---

## 9. Redis Session & Cache Management

### 9.1 Functional Requirements
Redis 7 is utilized for two distinct workloads:
1. **User Session Store (`auth-service`):** Stateful session management storing session IDs mapped to user credentials with TTL expiration.
2. **Celery Result Backend (`ai-service`):** Stores asynchronous task statuses and execution results for intelligence analysis jobs.

### 9.2 Network Isolation & Health Verification
Redis publishes zero host ports and communicates strictly across the private `globepulse-net` Docker bridge network. It is completely isolated from the host and **never** exposed to the public internet.

Verify Redis operational health:
```bash
docker compose exec redis redis-cli ping
# Expected output: PONG

docker compose exec redis redis-cli info memory
```

---

## 10. RabbitMQ Message Broker Configuration

### 10.1 Functional Requirements
RabbitMQ 3 handles asynchronous messaging between:
- Event producers (e.g., `analytics-service`, API requests)
- Event consumers (e.g., `ai-worker` processing `ai_analysis_queue`)

AMQP broker traffic (`5672`) publishes zero host ports and operates entirely within `globepulse-net`.

### 10.2 Management Dashboard Secure Access
RabbitMQ includes a web management console on port `15672`. In production, this port is bound strictly to `127.0.0.1:15672` (loopback only) and blocked by UFW from external traffic.

To inspect queues securely without exposing the dashboard to the public internet, create an SSH local port-forwarding tunnel from your workstation:
```bash
# Execute from your local workstation:
ssh -L 15672:localhost:15672 deploy@YOUR_SERVER_IP
```
Now open your local web browser to `http://localhost:15672` and authenticate using `gp_rabbit` and the generated `${RABBIT_PASS}`.

Verify RabbitMQ health via CLI:
```bash
docker compose exec rabbitmq rabbitmq-diagnostics ping
# Expected output: Ping succeeded
```

---

## 11. Docker Compose Production Deployment

### 11.1 Build and Pre-flight Inspection
Ensure that Docker build contexts compile cleanly. Note that `ai-worker` reuses the `globepulse-ai-service:latest` image built by `ai-service`, eliminating redundant builds and preventing parallel build collisions:

```bash
cd /opt/globepulse/app

# Pull external base images
docker compose pull postgres redis rabbitmq

# Build microservices and frontend
# (Use --parallel for fast multi-core builds, or omit --parallel on memory-constrained 1GB-2GB VPS hosts)
docker compose build --parallel
```

Verify that all built application images exist locally:
```bash
docker images | grep -E "(globepulse|app-)"
```

### 11.2 Launch Core Infrastructure First
To prevent connection race conditions, start databases and message brokers before launching application consumers:

```bash
docker compose up -d postgres redis rabbitmq
docker compose ps
```

Wait 10 seconds for PostgreSQL to initialize the database cluster, then run the database migrations detailed in Section 8.2.

### 11.3 Launch Microservices and Frontend
```bash
docker compose up -d
```

Verify that all 10 containers are running:
```bash
docker compose ps
```
Expected output:
```text
NAME                     IMAGE                               COMMAND                  SERVICE             STATUS              PORTS
app-ai-service-1         globepulse-ai-service:latest        "uvicorn app.main:ap…"   ai-service          Up (healthy)        127.0.0.1:8083->8083/tcp
app-ai-worker-1          globepulse-ai-service:latest        "celery -A app.core.…"   ai-worker           Up                  
app-analytics-service-1  app-analytics-service               "./main"                 analytics-service   Up                  127.0.0.1:8084->8084/tcp
app-auth-service-1       app-auth-service                    "./main"                 auth-service        Up                  127.0.0.1:8081->8081/tcp
app-country-service-1    app-country-service                 "./main"                 country-service     Up                  127.0.0.1:8082->8082/tcp
app-frontend-1           app-frontend                        "/docker-entrypoint.…"   frontend            Up                  127.0.0.1:3100->80/tcp
app-news-service-1       app-news-service                    "./main"                 news-service        Up                  127.0.0.1:8080->8080/tcp
app-postgres-1           postgres:15-alpine                  "docker-entrypoint.s…"   postgres            Up (healthy)        5432/tcp
app-rabbitmq-1           rabbitmq:3-management-alpine        "docker-entrypoint.s…"   rabbitmq            Up (healthy)        5672/tcp, 127.0.0.1:15672->15672/tcp
app-redis-1              redis:7-alpine                      "docker-entrypoint.s…"   redis               Up (healthy)        6379/tcp
```

---

## 12. Startup Order & Service Health Verification

### 12.1 Dependency Flow Matrix

```text
Layer 1: Data Stores (Must be healthy first)
         postgres (5432) ───► redis (6379) ───► rabbitmq (5672)
                   │                    │                │
Layer 2: Backend Core Services          │                │
         ├── auth-service (8081) ◄──────┘                │
         ├── news-service (8080) ◄───────────────────────┤
         ├── country-service (8082)                      │
         ├── analytics-service (8084) ◄──────────────────┤
         └── ai-service (8083) ◄─────────────────────────┤
                   │                                     │
Layer 3: Background Workers                              │
         └── ai-worker (Celery consumer) ◄───────────────┘
                   │
Layer 4: Edge Presentation
         └── frontend (React / Nginx 3100)
```

### 12.2 Container Health Audit
Verify endpoints using `curl` from the VPS terminal:

```bash
# 1. News Service General Health
curl -s http://127.0.0.1:8080/health
# Output: {"status": "ok"}

# 2. News Service Ingestion Pipeline Health (Detailed Connector Telemetry)
curl -s http://127.0.0.1:8080/health/ingestion
# Output: {"connectors":{...},"status":"healthy"}

# 3. Country Service Health
curl -s http://127.0.0.1:8082/health
# Output: {"status": "ok"}

# 4. Analytics Service Health
curl -s http://127.0.0.1:8084/health
# Output: {"status": "ok"}

# 5. Auth Service Health
curl -s http://127.0.0.1:8081/health
# Output: {"status":"ok"}

# 6. AI Service Health
curl -s http://127.0.0.1:8083/health
# Output: {"status":"ok"}

# 7. Frontend Container HTTP Response
curl -sI http://127.0.0.1:3100 | grep "HTTP/1.1"
# Output: HTTP/1.1 200 OK
```

---

## 13. Nginx Reverse Proxy & Edge Routing

The host Nginx acts as the public-facing gateway, terminating TLS, enforcing HTTP security headers, and proxying traffic to the private Docker loopback bindings.

### 13.1 Production Nginx Configuration

Create `/etc/nginx/sites-available/globepulse.conf`:

```nginx
# ==============================================================================
# GlobePulse AI — Production Reverse Proxy
# Domain Placeholders: app.example.com (UI) & api.example.com (API)
# ==============================================================================

# Rate limiting zones to protect auth and ingestion from DDoS
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=10r/m;
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=60r/m;

# ------------------------------------------------------------------------------
# 1. Frontend Web Interface (app.example.com)
# ------------------------------------------------------------------------------
server {
    listen 80;
    listen [::]:80;
    server_name app.example.com;

    # Allow ACME Challenge for Let's Encrypt
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name app.example.com;

    # SSL Certificates (managed by Certbot)
    ssl_certificate /etc/letsencrypt/live/app.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app.example.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Security Headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' https: data: blob: 'unsafe-inline' 'unsafe-eval';" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Proxy to Docker Frontend Container (Vite / Nginx)
    location / {
        proxy_pass http://127.0.0.1:3100;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support (for Vite HMR or future live feeds)
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    access_log /var/log/nginx/globepulse_app_access.log;
    error_log /var/log/nginx/globepulse_app_error.log;
}

# ------------------------------------------------------------------------------
# 2. Microservices API Gateway (api.example.com)
# ------------------------------------------------------------------------------
server {
    listen 80;
    listen [::]:80;
    server_name api.example.com;

    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name api.example.com;

    ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Global Security & Buffer Configurations
    client_max_body_size 10M;
    proxy_connect_timeout 15s;
    proxy_send_timeout 30s;
    proxy_read_timeout 30s;
    proxy_buffer_size 16k;
    proxy_buffers 4 32k;

    # CORS Headers at Gateway Layer
    add_header 'Access-Control-Allow-Origin' 'https://app.example.com' always;
    add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
    add_header 'Access-Control-Allow-Headers' 'DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Authorization' always;
    add_header 'Access-Control-Allow-Credentials' 'true' always;

    # Handle Preflight OPTIONS
    if ($request_method = 'OPTIONS') {
        return 204;
    }

    # Route: Auth Service (/api/v1/auth/*)
    location /api/v1/auth/ {
        limit_req zone=auth_limit burst=5 nodelay;
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Route: News & Ingestion Feeds
    location /api/v1/news/ {
        limit_req zone=api_limit burst=20 nodelay;
        proxy_pass http://127.0.0.1:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Route: Ingestion Monitoring Health Endpoint
    location /health/ingestion {
        proxy_pass http://127.0.0.1:8080/health/ingestion;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Route: Country Service
    location /api/v1/countries/ {
        limit_req zone=api_limit burst=20 nodelay;
        proxy_pass http://127.0.0.1:8082/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Route: AI Analysis Service
    location /api/v1/ai/ {
        limit_req zone=api_limit burst=10 nodelay;
        proxy_pass http://127.0.0.1:8083/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Route: Analytics Telemetry Service
    location /api/v1/analytics/ {
        limit_req zone=api_limit burst=20 nodelay;
        proxy_pass http://127.0.0.1:8084/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    access_log /var/log/nginx/globepulse_api_access.log;
    error_log /var/log/nginx/globepulse_api_error.log;
}
```

Enable the configuration:
```bash
sudo ln -s /etc/nginx/sites-available/globepulse.conf /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
```

---

## 14. SSL/TLS Security via Let's Encrypt (Certbot)

### 14.1 Initial Certificate Acquisition (Webroot Method)
Ensure DNS records for `app.example.com` and `api.example.com` resolve to the server IP before proceeding.

Create the webroot challenge directory:
```bash
sudo mkdir -p /var/www/certbot
sudo chown -R www-data:www-data /var/www/certbot
```

Temporarily issue certificates using Certbot:
```bash
sudo certbot certonly --webroot -w /var/www/certbot \
  -d app.example.com \
  -d api.example.com \
  --email admin@example.com \
  --agree-tos \
  --no-eff-email
```

Generate strong Diffie-Hellman parameters:
```bash
sudo openssl dhparam -out /etc/letsencrypt/ssl-dhparams.pem 2048
```

### 14.2 Verify Nginx Configuration & Reload
```bash
sudo nginx -t
# Output must be: syntax is ok / test is successful

sudo systemctl reload nginx
```

### 14.3 Automated Renewal Verification
Certbot automatically installs a systemd timer on Ubuntu 24.04. Check its status:
```bash
sudo systemctl status certbot.timer
```

Perform a renewal dry run to ensure future renewals succeed:
```bash
sudo certbot renew --dry-run
```

---

## 15. CORS Policy & Cross-Origin Security

The frontend application hosted on `https://app.example.com` makes asynchronous requests to `https://api.example.com`.

### Critical Production Rules:
1. **Forbidden Wildcard with Credentials:** Never configure `Access-Control-Allow-Origin: *` when credentials (`cookies`, `Authorization` headers) are included. Modern browsers will reject the response with a CORS violation.
2. **Strict Origin Binding:** The Nginx reverse proxy explicitly specifies:
   ```nginx
   add_header 'Access-Control-Allow-Origin' 'https://app.example.com' always;
   add_header 'Access-Control-Allow-Credentials' 'true' always;
   ```
3. **Preflight Handling:** Preflight `OPTIONS` requests are intercepted at the Nginx gateway and returned immediately with `204 No Content`, preventing unauthenticated preflight requests from generating unnecessary backend load.

---

## 16. Authentication, JWT & Secret Management

### 16.1 Architecture
The `auth-service` implements hybrid stateless/stateful authentication:
- **Stateless Verification:** Issues signed JWT tokens using HMAC-SHA256 (`JWT_SECRET`).
- **Stateful Revocation:** Stores active user `session_id` entries in Redis. When a user logs out (`/api/v1/auth/logout`), the session key is revoked in Redis, instantly invalidating the session regardless of JWT expiration.
- **Password Security:** Passwords are salted and hashed using `bcrypt` (cost factor 10).

### 16.2 Secret Generation & Rotation
Generate production keys using cryptographic randomness:
```bash
openssl rand -hex 32
```
Never commit `.env` files to git. Store backup copies of production secrets in an encrypted vault (such as HashiCorp Vault, 1Password, or Bitwarden).

---

## 17. External Intelligence Source Connectors

The `news-service` continuously pulls data from external humanitarian and threat observation feeds.

| Connector | Target Feed URL | Frequency / Timeout | Protocol / Format | Auth Required |
| :--- | :--- | :--- | :--- | :--- |
| **GDELT 2.0 API** | `https://api.gdeltproject.org/api/v2/doc/doc` | 15s Timeout, 3 Retries | REST / JSON | None (Public) |
| **USGS Earthquakes**| `https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson` | 10s Timeout, 3 Retries | GeoJSON / REST | None (Public) |
| **ReliefWeb API** | `https://api.reliefweb.int/v1/reports?appname=globepulse` | 10s Timeout, 3 Retries | REST / JSON | None (Public) |

### Failure Isolation Guarantees
- **Non-blocking Retries:** Each connector runs via an isolated HTTP client with exponential backoff (100ms, 200ms, 400ms).
- **Goroutine Panic Recovery:** The ingestion scheduler (`scheduler.go`) wraps each connector execution in a deferred panic recovery block. If the GDELT API returns malformed JSON or times out, it logs an error and records a consecutive failure metric, but **never crashes the Go process or interrupts other feeds**.

---

## 18. Ingestion Scheduler & Pipeline Operations

The ingestion engine follows this strict sequence:

```text
External Source (GDELT / USGS / ReliefWeb)
         │
         ▼
[Connector Fetch] (10–15s Timeout, Exponential Backoff)
         │
         ▼
[Record Normalization] (Transform to Canonical ThreatEvent Model)
         │
         ▼
[Deduplication Check] (PostgreSQL: source_id + external_id uniqueness)
         │
         ▼
[Enrichment Pipeline] (Provenance, Geo-ISO, Coordinate Normalization)
         │
         ▼
[Entity Extraction] (Deterministic URLs, Domains, Locations, ISO Countries)
         │
         ▼
[PostgreSQL Atomic Transaction] (Save Sources, Items, Events, Entities)
```

### Ingestion Interval Configuration
The scheduler polling interval is governed by `INGESTION_INTERVAL` in `.env`:
- Default: `10m` (10 minutes)
- Allowed values: Standard Go duration strings (`5m`, `15m`, `30m`, `1h`)
- To adjust without downtime: Update `.env` and execute `docker compose up -d news-service`.

---

## 19. Production Logging & Log Rotation

Unmanaged Docker logs can fill an SSD within weeks, causing disk write failures and database crashes.

### 19.1 Configure Global Docker Daemon Log Rotation
Create `/etc/docker/daemon.json` on the Ubuntu host:

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "20m",
    "max-file": "5"
  }
}
```

Reload Docker daemon to enforce log caps (20 MB per file, max 5 files = 100 MB per container):
```bash
sudo systemctl reload docker
```

### 19.2 Real-time Log Inspection Commands
```bash
cd /opt/globepulse/app

# Follow all logs in real time
docker compose logs -f

# Inspect specific microservices with timestamps
docker compose logs -f --tail=100 news-service
docker compose logs -f --tail=100 auth-service
docker compose logs -f --tail=100 ai-worker

# Inspect host Nginx access and error logs
sudo tail -f /var/log/nginx/globepulse_api_error.log
sudo tail -f /var/log/nginx/globepulse_app_access.log
```

---

## 20. Host & Application Health Monitoring

For a single VPS deployment, run native diagnostics without adding heavy monitoring infrastructure.

### 20.1 System Resource Telemetry
```bash
# Host CPU, Memory, Load Average
htop

# Disk Utilization
df -h /

# Docker Container Memory & CPU Consumption
docker stats --no-stream
```

### 20.2 Automated Health Check Shell Script
Create `/opt/globepulse/scripts/healthcheck.sh`:

```bash
#!/bin/bash
# ==============================================================================
# GlobePulse AI — Production Health Check Utility
# ==============================================================================
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

echo "=== GlobePulse AI Health Check: $(date -u) ==="

check_service() {
    NAME=$1
    URL=$2
    EXPECTED=$3

    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$URL" || echo "000")
    if [ "$HTTP_CODE" = "$EXPECTED" ]; then
        echo -e "[${GREEN}OK${NC}] $NAME (HTTP $HTTP_CODE)"
    else
        echo -e "[${RED}FAIL${NC}] $NAME (Expected $EXPECTED, got $HTTP_CODE)"
    fi
}

# Microservices Health
check_service "Auth Service" "http://127.0.0.1:8081/health" "200"
check_service "News Service Health" "http://127.0.0.1:8080/health" "200"
check_service "Ingestion Telemetry" "http://127.0.0.1:8080/health/ingestion" "200"
check_service "Country Service" "http://127.0.0.1:8082/health" "200"
check_service "AI Service" "http://127.0.0.1:8083/health" "200"
check_service "Analytics Service" "http://127.0.0.1:8084/health" "200"
check_service "Frontend Container" "http://127.0.0.1:3100" "200"

# Data Stores Health
if docker compose -f /opt/globepulse/app/docker-compose.yml exec -T postgres pg_isready -U gp_admin -d globepulse > /dev/null 2>&1; then
    echo -e "[${GREEN}OK${NC}] PostgreSQL Database"
else
    echo -e "[${RED}FAIL${NC}] PostgreSQL Database is down"
fi

if docker compose -f /opt/globepulse/app/docker-compose.yml exec -T redis redis-cli ping > /dev/null 2>&1; then
    echo -e "[${GREEN}OK${NC}] Redis Session Store"
else
    echo -e "[${RED}FAIL${NC}] Redis Session Store is down"
fi

if docker compose -f /opt/globepulse/app/docker-compose.yml exec -T rabbitmq rabbitmq-diagnostics ping > /dev/null 2>&1; then
    echo -e "[${GREEN}OK${NC}] RabbitMQ Message Broker"
else
    echo -e "[${RED}FAIL${NC}] RabbitMQ Message Broker is down"
fi

echo "============================================="
```

Make executable:
```bash
chmod +x /opt/globepulse/scripts/healthcheck.sh
/opt/globepulse/scripts/healthcheck.sh
```

---

## 21. Automated Backup Strategy

Data safety requires automated logical backups of PostgreSQL and persistent Docker volumes.

### 21.1 PostgreSQL Backup Script
Create `/opt/globepulse/scripts/backup_postgres.sh`:

```bash
#!/bin/bash
# ==============================================================================
# Automated PostgreSQL Backup Script with Retention Pruning
# ==============================================================================
set -e

BACKUP_DIR="/opt/globepulse/backups/postgres"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILENAME="globepulse_db_${TIMESTAMP}.sql.gz"
RETENTION_DAYS=14

mkdir -p "$BACKUP_DIR"

echo "[$(date)] Starting PostgreSQL backup..."
docker compose -f /opt/globepulse/app/docker-compose.yml exec -T postgres \
  pg_dump -U gp_admin -d globepulse --clean --if-exists | gzip > "${BACKUP_DIR}/${FILENAME}"

chmod 600 "${BACKUP_DIR}/${FILENAME}"
echo "[$(date)] Backup completed successfully: ${BACKUP_DIR}/${FILENAME}"

# Prune backups older than 14 days
find "$BACKUP_DIR" -type f -name "globepulse_db_*.sql.gz" -mtime +$RETENTION_DAYS -delete
echo "[$(date)] Pruned backups older than $RETENTION_DAYS days."
```

Make executable:
```bash
chmod +x /opt/globepulse/scripts/backup_postgres.sh
```

### 21.2 Configure Daily Automated Backup Cron Job
Add to user `deploy`'s crontab (`crontab -e`):
```cron
# Execute database backup daily at 02:30 UTC
30 2 * * * /opt/globepulse/scripts/backup_postgres.sh >> /opt/globepulse/logs/backup.log 2>&1
```

### 21.3 Docker Volume Backup
To backup RabbitMQ state or persistent volumes:
```bash
sudo tar -czvf /opt/globepulse/backups/rabbitmq_volume_$(date +%F).tar.gz -C /var/lib/docker/volumes/app_rabbitmq_data .
```

---

## 22. Disaster Recovery Procedure

In the event of total server loss, follow this checklist to restore operations on a fresh Ubuntu 24.04 VPS.

### Disaster Recovery Checklist
```text
[ ] Step 1: Provision fresh Ubuntu Server 24.04 LTS VPS.
[ ] Step 2: Perform Section 4 base hardening (User deploy, SSH keys, UFW, Swap).
[ ] Step 3: Install Docker, Nginx, and Certbot (Section 5).
[ ] Step 4: Clone repository into /opt/globepulse/app (Section 6).
[ ] Step 5: Restore /opt/globepulse/app/.env and docker-compose.override.yml from secure vault.
[ ] Step 6: Start databases: docker compose up -d postgres redis rabbitmq.
[ ] Step 7: Restore PostgreSQL database from latest backup archive.
[ ] Step 8: Build and launch all microservices: docker compose up -d.
[ ] Step 9: Restore Nginx configuration to /etc/nginx/sites-available/globepulse.conf.
[ ] Step 10: Obtain SSL certificates via Certbot.
[ ] Step 11: Execute /opt/globepulse/scripts/healthcheck.sh to verify.
```

### Restoring Database Dump
```bash
# Uncompress and pipe backup into fresh PostgreSQL container
gunzip < /opt/globepulse/backups/postgres/globepulse_db_YYYYMMDD_HHMMSS.sql.gz | \
  docker compose -f /opt/globepulse/app/docker-compose.yml exec -T postgres psql -U gp_admin -d globepulse
```

---

## 23. Zero-Downtime Reality & Application Update Procedure

### Realistic Uptime Expectations
On a single-host VPS using Docker Compose without an active load balancer or Kubernetes rolling ingress:
- **Updates cause brief downtime:** Restarting or rebuilding a container causes an interruption of **2 to 10 seconds** while the process terminates and the new binary binds to the port.
- **Do not claim zero-downtime:** Inform stakeholders that updates are deployed within a minor maintenance window or scheduled during off-peak hours.

### Standard Git Update Flow
To deploy code updates safely:

```bash
cd /opt/globepulse/app

# 1. Create immediate pre-deployment database backup
/opt/globepulse/scripts/backup_postgres.sh

# 2. Pull new git commits
git pull origin main

# 3. Build only modified images
docker compose build

# 4. Check for and apply new database migrations (if any)
# Example: docker compose exec -T postgres psql -U gp_admin -d globepulse < services/news-service/migrations/000003_example.up.sql

# 5. Recreate containers with zero unnecessary downtime
docker compose up -d --remove-orphans

# 6. Verify health
/opt/globepulse/scripts/healthcheck.sh

# 7. Check container logs for errors
docker compose logs --tail=50 -f news-service
```

### Rollback Procedure
If the new deployment fails:
```bash
# Rollback git commit
git checkout HEAD~1

# Rebuild and restart previous stable version
docker compose build
docker compose up -d
```

---

## 24. Network Firewall (UFW) & Port Exposure Matrix

### Port Exposure Matrix

| Port | Protocol | Service | Public Access | Binding Policy | Rationale |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **22** | TCP | OpenSSH | **Yes** | `0.0.0.0:22` | Administrative server management |
| **80** | TCP | Nginx HTTP | **Yes** | `0.0.0.0:80` | ACME challenges & HTTPS redirects |
| **443**| TCP | Nginx HTTPS| **Yes** | `0.0.0.0:443`| Secure user traffic |
| **3100**| TCP | Frontend | **No** | `127.0.0.1:3100` | Proxied exclusively through Nginx (host port 3000 preserved for existing WhatsApp container) |
| **5432**| TCP | PostgreSQL | **No** | **None** (Docker network only) | Internal datastore traffic across `globepulse-net`; completely isolated from `n8n-postgres` |
| **5672**| TCP | RabbitMQ AMQP | **No** | **None** (Docker network only) | Internal Celery broker traffic across `globepulse-net` |
| **6379**| TCP | Redis | **No** | **None** (Docker network only) | Internal session/caching traffic across `globepulse-net` |
| **8080**| TCP | news-service | **No** | `127.0.0.1:8080` | Proxied through Nginx `/api/v1/news` |
| **8081**| TCP | auth-service | **No** | `127.0.0.1:8081` | Proxied through Nginx `/api/v1/auth` |
| **8082**| TCP | country-service | **No**| `127.0.0.1:8082` | Proxied through Nginx `/api/v1/countries` |
| **8083**| TCP | ai-service | **No** | `127.0.0.1:8083` | Proxied through Nginx `/api/v1/ai` |
| **8084**| TCP | analytics-service | **No** | `127.0.0.1:8084` | Proxied through Nginx `/api/v1/analytics` |
| **15672**| TCP| RabbitMQ Web UI | **No** | `127.0.0.1:15672`| Accessible via SSH tunnel only |

> **IMPORTANT DOCKER NETWORKING NOTE:**  
> By default, Docker manipulates `iptables` rules directly, creating NAT table entries that can bypass UFW rules when ports are mapped to `0.0.0.0`. By omitting host port publications for `postgres`, `redis`, and RabbitMQ AMQP (keeping them entirely within `globepulse-net`), and binding Nginx-proxied microservices explicitly to `127.0.0.1:<PORT>:<PORT>`, the attack surface is completely minimized and host-port collisions with existing workloads are entirely avoided.

---

## 25. Troubleshooting & Diagnostic Playbook

### Problem 1: Microservice Container Won't Start or Exits Immediately
**Symptom:** `docker compose ps` shows status `Exited (1)`.  
**Diagnostics:**
```bash
docker compose logs <service-name>
```
**Common Causes:**
- Database not ready when Go service started (`connection refused`). Verify PostgreSQL is running: `docker compose exec postgres pg_isready`.
- Incorrect environment variable in `.env`. Verify variables match Section 7.1.

---

### Problem 2: 502 Bad Gateway Reported by Nginx
**Symptom:** Browser returns `502 Bad Gateway` when loading `https://app.example.com` or `https://api.example.com`.  
**Diagnostics:**
```bash
sudo tail -n 50 /var/log/nginx/globepulse_api_error.log
sudo tail -n 50 /var/log/nginx/globepulse_app_error.log
```
**Common Causes:**
- The upstream Docker container is stopped. Verify: `docker compose ps`.
- Upstream port mismatch. Confirm the container is listening on `127.0.0.1` at the port specified in `proxy_pass`.
- Docker bridge network failure: Restart container with `docker compose restart <service>`.

---

### Problem 3: PostgreSQL Database Connection Failure
**Symptom:** Backend logs report `Failed to connect to PostgreSQL: dial tcp 172.x.x.x:5432: connect: connection refused`.  
**Diagnostics:**
```bash
docker compose exec postgres psql -U gp_admin -d globepulse -c "SELECT 1;"
```
**Common Causes:**
- PostgreSQL container is still initializing.
- Mismatched `POSTGRES_PASSWORD` and `DB_PASS` in `.env`.
- Database name mismatch: ensure both `POSTGRES_DB` and `DB_NAME` are set to `globepulse`.

---

### Problem 4: RabbitMQ Connection Refused (AI Worker or Analytics)
**Symptom:** Celery worker reports `kombu.exceptions.OperationalError: [Errno 111] Connection refused`.  
**Diagnostics:**
```bash
docker compose exec rabbitmq rabbitmq-diagnostics check_running
```
**Fix:**
Ensure `RABBITMQ_URL` in `.env` uses the internal Docker service name:
`amqp://gp_rabbit:PASSWORD@rabbitmq:5672/`  
(Do NOT use `localhost` or `127.0.0.1` inside container environment variables).

---

### Problem 5: Ingestion Feeds Failing or Stalled
**Symptom:** Database table `threat_events` is not receiving new rows.  
**Diagnostics:**
```bash
curl -s http://127.0.0.1:8080/health/ingestion | jq .
```
Check `consecutive_failures` counter for each connector.  
**Common Causes:**
- Outbound DNS resolution failure inside Docker container. Test DNS:
  ```bash
  docker compose exec news-service ping -c 2 api.gdeltproject.org
  ```
- External rate-limiting or feed downtime.

---

### Problem 6: Let's Encrypt Certificate Issuance / Renewal Fails
**Symptom:** `certbot renew` fails with `Timeout during connect (likely firewall problem)`.  
**Diagnostics:**
- Verify port 80 is open: `sudo ufw status | grep 80`.
- Verify DNS resolves to the VPS IP: `dig +short app.example.com`.
- If using Cloudflare, temporarily disable proxy mode (switch from Orange to Grey Cloud).

---

## 26. Post-Deployment Verification Checklist

Execute this checklist sequentially after deploying on the VPS:

- [ ] **VPS Hardening:** Non-root `deploy` user configured; password authentication disabled in SSH; root login restricted.
- [ ] **Firewall Active:** `sudo ufw status` confirms only ports `22`, `80`, and `443` are allowed publicly.
- [ ] **Host Ports Protected & Isolated:** `ss -lntp` or `netstat -tuln` confirms:
  - Ports `5432` (PostgreSQL), `6379` (Redis), and `5672` (RabbitMQ AMQP) are **not** published to any host interface.
  - Frontend (`3100`), RabbitMQ Management (`15672`), and microservice ports (`8080-8084`) are bound strictly to `127.0.0.1`.
  - Port `3000` remains claimed exclusively by the existing `go-whatsapp-web-multidevice-whatsapp_go-1` container without collision.
- [ ] **Existing Workloads Untouched:** `docker ps` verifies that existing containers (`n8n`, `n8n-postgres`, and `go-whatsapp-web-multidevice-whatsapp_go-1`) are intact, healthy, and undisturbed.
- [ ] **GlobePulse Containers Running:** `docker compose ps` shows all 10 containers in an `Up` status.
- [ ] **Database Migrations Complete:** `docker compose exec postgres psql -U gp_admin -d globepulse -c "\dt"` outputs 8 tables.
- [ ] **Ingestion Active:** Querying `SELECT count(*) FROM threat_events;` shows row count increasing across scheduler cycles.
- [ ] **Frontend Operational:** Navigating to `https://app.example.com` loads the 3D globe interface over valid HTTPS via reverse proxy to `127.0.0.1:3100`.
- [ ] **API Gateway Functional:** `curl -sI https://api.example.com/health/ingestion` returns `HTTP/2 200`.
- [ ] **SSL Auto-Renewal Confirmed:** `sudo certbot renew --dry-run` reports success.
- [ ] **Automated Backups Scheduled:** `crontab -l` displays active daily backup job for `/opt/globepulse/scripts/backup_postgres.sh`.
- [ ] **Docker Log Rotation Enforced:** `/etc/docker/daemon.json` contains `max-size: 20m`.

---

## 27. Production Command Quick Reference

| Operational Task | Shell Command |
| :--- | :--- |
| **Start Entire Stack** | `cd /opt/globepulse/app && docker compose up -d` |
| **Stop Entire Stack (GlobePulse ONLY)** | `cd /opt/globepulse/app && docker compose down` *(Never run globally!)* |
| **Restart Single Service** | `docker compose restart <service-name>` |
| **Inspect Container Status** | `docker compose ps` |
| **Verify Existing Host Containers** | `docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"` |
| **Live Stream All Logs** | `docker compose logs -f --tail=100` |
| **Live Stream Specific Service**| `docker compose logs -f --tail=100 news-service` |
| **Trigger Manual DB Backup**| `/opt/globepulse/scripts/backup_postgres.sh` |
| **PostgreSQL Interactive Shell**| `docker compose exec postgres psql -U gp_admin -d globepulse` |
| **Redis Interactive Shell**| `docker compose exec redis redis-cli` |
| **Check Ingestion Health** | `curl -s http://127.0.0.1:8080/health/ingestion \| jq .` |
| **Nginx Syntax Check** | `sudo nginx -t` |
| **Nginx Reload** | `sudo systemctl reload nginx` |
| **Certbot Dry Run** | `sudo certbot renew --dry-run` |
| **Clean Unused Docker Images (Scoped)**| `docker image prune -f` *(Do NOT prune volumes or run prune -a globally)* |
| **Inspect Disk Space** | `df -h /` |

---

## 28. Production Security Audit Checklist

Before releasing the platform to production traffic, verify each security control:

- [ ] **No Default Credentials:** Default passwords (`devpassword`, `devuser`, `supersecret`, `guest`) are purged and replaced with cryptographically random strings.
- [ ] **No Plaintext Secrets in Version Control:** The `.env` file is listed in `.gitignore` and has permissions `600` on the VPS.
- [ ] **Database Isolation:** PostgreSQL does not publish any host port; it communicates exclusively across the internal `globepulse-net` Docker network with zero exposure to the host or existing `n8n-postgres`.
- [ ] **Redis Isolation:** Redis does not publish any host port; it communicates exclusively across the internal `globepulse-net` network.
- [ ] **RabbitMQ Management Dashboard Secured:** Port `15672` is bound strictly to `127.0.0.1:15672` and blocked by UFW; access is restricted to SSH tunneling.
- [ ] **Frontend Port Decoupled:** Frontend is bound to `127.0.0.1:3100`, eliminating any collision with the existing `whatsapp_go` container on port 3000.
- [ ] **Microservices Loopback Bound:** All 5 API services (`8080-8084`) are bound strictly to `127.0.0.1`, reachable exclusively by the host Nginx reverse proxy.
- [ ] **CORS Restricted:** Nginx CORS configuration explicitly reflects `https://app.example.com` and disallows wildcard origins.
- [ ] **HTTP Security Headers Present:** Responses include `X-Frame-Options`, `X-Content-Type-Options`, `X-XSS-Protection`, and `Strict-Transport-Security`.
- [ ] **Unattended Upgrades Active:** Ubuntu OS is configured to automatically download and apply critical security patches.

---

## 29. Remaining Manual Verification & Operational Notes

### 29.1 Resolved Production Blockers Audit

The deployment blocker audit resolved the genuine blockers identified in the repository:

1. **`country-service` & `analytics-service` Missing Migration Directories (RESOLVED):**
   * **Resolution:** Removed the redundant and non-existent `COPY --from=builder /app/migrations ./migrations` directives from `services/country-service/Dockerfile` and `services/analytics-service/Dockerfile`. Neither service executes runtime schema migrations (all database tables are managed by `auth-service` and `news-service`). Both Docker images now build cleanly.

2. **Docker Compose Healthcheck Directives (RESOLVED):**
   * **Resolution:** Added native `healthcheck` specifications to `docker-compose.yml` for infrastructure containers:
     * `postgres`: `pg_isready -U devuser -d globepulse` (5s interval, 5s timeout, 5 retries)
     * `redis`: `redis-cli ping` (5s interval, 5s timeout, 5 retries)
     * `rabbitmq`: `rabbitmq-diagnostics -q ping` (10s interval, 10s timeout, 5 retries)
     * `ai-service`: HTTP probe against `http://127.0.0.1:8083/health` via standard Python `urllib` (no external curl/wget required)

3. **`auth-service` Health Endpoint (RESOLVED):**
   * **Resolution:** Implemented a public, unauthenticated, lightweight `GET /health` endpoint in `services/auth-service/internal/handler/http/router.go` returning `{"status": "ok"}` with HTTP 200. This unifies health verification across all backend services (`auth-service:8081`, `news-service:8080`, `country-service:8082`, `ai-service:8083`, and `analytics-service:8084`).

4. **Multi-Tenant VPS Coexistence & Port Isolation (RESOLVED):**
   * **Resolution:** Reconfigured `docker-compose.yml` and `docker-compose.override.yml`:
     * Removed host port publication for `postgres` (5432) and `redis` (6379) — both now reside purely on the internal Docker network.
     * Removed AMQP host publication (5672); bound RabbitMQ management (15672) to loopback `127.0.0.1:15672:15672` for local SSH tunnels.
     * Remapped the GlobePulse frontend host port from `3000` to `127.0.0.1:3100:80` to prevent collision with the existing `go-whatsapp-web-multidevice-whatsapp_go-1` container bound to host port 3000.
     * Bound all 5 backend microservices (`8080-8084`) strictly to `127.0.0.1`.
     * Placed all containers on a dedicated, isolated custom bridge network `globepulse-net`, ensuring zero DNS or IP collisions with `n8n` or other workloads.

### 29.2 Remaining Manual Verification Required on Target VPS

The following operational tasks require the live Ubuntu 24.04 LTS host environment and cannot be executed inside the container build sandbox:

* **Verify Existing Workloads Intact:** Run `docker ps` to verify that `n8n`, `n8n-postgres`, and `go-whatsapp-web-multidevice-whatsapp_go-1` remain in the `Up` state after GlobePulse deployment.
* **VPS DNS Propagation:** Verify that A/AAAA records for `app.example.com` and `api.example.com` resolve to the public IP address of the target VPS (`dig +short A app.example.com`).
* **Let's Encrypt TLS Certificate Issuance:** Run `sudo certbot --nginx -d app.example.com -d api.example.com` against the live internet-facing Nginx instance.
* **UFW Firewall State:** Verify that UFW allows only ports 22, 80, and 443 (`sudo ufw status verbose`).
* **Hardware Sizing & Swap File:** Verify that the 4 GB swap file is mounted and active (`swapon --show`, `free -m`).
* **Automated Daily Backup Cron:** Confirm the backup cron job executes and writes encrypted archives to `/opt/globepulse/backups/`.
* **Frontend Microservices Integration:** The React frontend currently renders threat markers and alerts using static client-side fixtures (`src/lib/data.ts`); connecting dynamic fetch queries to `/api/v1/news` or `/api/v1/auth` is deferred to subsequent application development phases.
