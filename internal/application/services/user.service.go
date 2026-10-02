package services

import (
	"context"
	"fmt"
	"grape-api/internal/apperror"
	"grape-api/internal/domain/auth"
	"grape-api/internal/domain/user"
	"grape-api/pkg"
	"uuid"
)

type UserService struct {
	transactionManager pkg.TransactionManager
	userRepository     user.IUserRepository
	authRepository     auth.IAuthRepository
}

func NewUserService(
	transactionManager pkg.TransactionManager,
	userRepository user.IUserRepository,
	authRepository auth.IAuthRepository,
) *UserService {
	return &UserService{
		transactionManager: transactionManager,
		userRepository:     userRepository,
		authRepository:     authRepository,
	}
}

func (s *UserService) CreateUser(name, username, password string, role user.USER_ROLE) (*user.User, error) {
	var createdUser *user.User

	err := s.transactionManager.WithinTransaction(context.Background(), func(txCtx context.Context) error {
		existingUser, err := s.userRepository.FindByUsername(txCtx, username)
		if err == nil && existingUser != nil && existingUser.ID() != uuid.Nil() {
			return apperror.New(apperror.CodeConflict, "user already exists")
		}

		hashedPassword, err := pkg.HashPassword(password)
		if err != nil {
			return fmt.Errorf("hash password: %w", err)
		}

		userData := user.NewUser(name, username, "", role, user.PENDING)
		createdUser, err = s.userRepository.Create(txCtx, userData)
		if err != nil {
			return err
		}

		account := auth.NewAccount(createdUser.ID(), auth.CREDENTIALS, hashedPassword)
		if err := s.authRepository.CreateAccount(txCtx, &account); err != nil {
			return fmt.Errorf("create credentials account: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
