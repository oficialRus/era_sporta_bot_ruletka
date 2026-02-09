package service

import (
	"context"

	"era_sporta_bot_ruletka/internal/domain"
	"era_sporta_bot_ruletka/internal/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
	spinRepo *repository.SpinRepository
}

func NewUserService(userRepo *repository.UserRepository, spinRepo *repository.SpinRepository) *UserService {
	return &UserService{userRepo: userRepo, spinRepo: spinRepo}
}

func (s *UserService) GetByTelegramID(ctx context.Context, telegramUserID int64) (*domain.User, error) {
	return s.userRepo.GetByTelegramID(ctx, telegramUserID)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *UserService) Upsert(ctx context.Context, u *domain.User) error {
	return s.userRepo.Upsert(ctx, u)
}

func (s *UserService) GetUserState(ctx context.Context, user *domain.User, spinLimit int) (*UserState, error) {
	spinCount, err := s.spinRepo.CountByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return &UserState{
		User:          user,
		SpinAvailable: spinCount < spinLimit,
		SpinsUsed:     spinCount,
		SpinLimit:     spinLimit,
	}, nil
}

type UserState struct {
	User          *domain.User
	SpinAvailable bool
	SpinsUsed     int
	SpinLimit     int
}
