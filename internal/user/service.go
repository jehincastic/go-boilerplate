package user

import (
	"context"
	"errors"
	"time"

	"github.com/jehincastic/go-boilerplate/internal/auth"
	"github.com/jehincastic/go-boilerplate/internal/platform/errs"
	"github.com/redis/go-redis/v9"
)

// RepositoryAPI is the user persistence surface used by the service.
type RepositoryAPI interface {
	Insert(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// Service registers users and checks passwords.
type Service struct {
	users     RepositoryAPI
	redis     *redis.Client
	jwtSecret string
	tokenTTL  time.Duration
}

// NewService returns a user service.
func NewService(users RepositoryAPI, redisClient *redis.Client, jwtSecret string, tokenTTL time.Duration) *Service {
	return &Service{users: users, redis: redisClient, jwtSecret: jwtSecret, tokenTTL: tokenTTL}
}

// Register stores a new user and caches the profile.
func (s *Service) Register(ctx context.Context, name, email, password string) (*User, error) {
	user := &User{Name: name, Email: email}
	if err := user.SetPassword(password); err != nil {
		return nil, err
	}
	if err := s.users.Insert(ctx, user); err != nil {
		return nil, err
	}
	if err := s.cache(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// Login returns an access token when the email and password match.
func (s *Service) Login(ctx context.Context, email, password string) (auth.Token, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return auth.Token{}, errs.ErrInvalidCredentials
		}
		return auth.Token{}, err
	}

	match, err := user.PasswordMatches(password)
	if err != nil {
		return auth.Token{}, err
	}
	if !match {
		return auth.Token{}, errs.ErrInvalidCredentials
	}

	if err := s.cache(ctx, user); err != nil {
		return auth.Token{}, err
	}
	return auth.Sign(user.ID, user.Email, s.jwtSecret, s.tokenTTL)
}

func (s *Service) cache(ctx context.Context, user *User) error {
	return auth.SaveSession(ctx, s.redis, auth.Session{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, s.tokenTTL)
}
