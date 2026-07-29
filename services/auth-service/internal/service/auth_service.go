package service

import (
	"context"
	"errors"
	"time"

	types "github.com/global-news/shared-types"

	"github.com/global-news/auth-service/internal/config"
	"github.com/global-news/auth-service/internal/domain"
	"github.com/global-news/auth-service/pkg/hash"
	"github.com/global-news/auth-service/pkg/jwt"
	"github.com/global-news/auth-service/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type authService struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	auditRepo   domain.AuditRepository
	cfg         *config.Config
}

func NewAuthService(ur domain.UserRepository, sr domain.SessionRepository, ar domain.AuditRepository, cfg *config.Config) domain.AuthUseCase {
	return &authService{
		userRepo:    ur,
		sessionRepo: sr,
		auditRepo:   ar,
		cfg:         cfg,
	}
}

func (s *authService) Register(ctx context.Context, email, password, firstName, lastName string) (*types.User, error) {
	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &types.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         types.RoleUser,
		IsVerified:   false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		logger.Log.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	s.auditRepo.LogAction(ctx, &domain.AuditLog{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Action:    "REGISTER",
		CreatedAt: time.Now(),
	})

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return "", "", errors.New("invalid credentials")
	}

	if !hash.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", errors.New("invalid credentials")
	}

	sessionID := uuid.New().String()
	duration := time.Duration(s.cfg.TokenExpiration) * time.Minute

	token, err := jwt.GenerateToken(user.ID, user.Role, sessionID, s.cfg.JWTSecret, duration)
	if err != nil {
		return "", "", err
	}

	session := &domain.Session{
		SessionID: sessionID,
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
	}

	err = s.sessionRepo.CreateSession(ctx, session, duration)
	if err != nil {
		return "", "", err
	}

	s.auditRepo.LogAction(ctx, &domain.AuditLog{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Action:    "LOGIN",
		CreatedAt: time.Now(),
	})

	return token, sessionID, nil
}

func (s *authService) Logout(ctx context.Context, sessionID string) error {
	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err == nil && session != nil {
		s.auditRepo.LogAction(ctx, &domain.AuditLog{
			ID:        uuid.New().String(),
			UserID:    session.UserID,
			Action:    "LOGOUT",
			CreatedAt: time.Now(),
		})
	}

	return s.sessionRepo.DeleteSession(ctx, sessionID)
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	return nil
}

func (s *authService) ResetPassword(ctx context.Context, email, newPassword string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	hashedPassword, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(ctx, user)
	if err == nil {
		s.sessionRepo.DeleteUserSessions(ctx, user.ID)
		s.auditRepo.LogAction(ctx, &domain.AuditLog{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			Action:    "PASSWORD_RESET",
			CreatedAt: time.Now(),
		})
	}
	return err
}

func (s *authService) GetProfile(ctx context.Context, userID string) (*types.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
