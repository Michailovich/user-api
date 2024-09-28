package userPack

import (
	"context"
	"time"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
	user.Created = time.Now()
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id int, user *User) error {
	return s.repo.UpdateUser(ctx, id, user)
}
