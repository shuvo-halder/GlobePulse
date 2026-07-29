import yaml
with open('docker-compose.yml', 'r') as f:
    data = yaml.safe_load(f)

data['services']['frontend'] = {
    'build': {
        'context': '.',
        'dockerfile': 'Dockerfile'
    },
    'ports': ['3000:80'],
    'environment': {'APP_ENV': 'development'},
    'depends_on': ['auth-service', 'news-service', 'country-service', 'analytics-service']
}

with open('docker-compose.yml', 'w') as f:
    yaml.dump(data, f, sort_keys=False)
