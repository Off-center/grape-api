package userusecases

import (
	"context"
	"grape-api/internal/application/services"
	"grape-api/internal/domain/user"
)

type CreateUserInput struct {
	Name     string
	Username string
	Password string
}

type CreateUserUseCase struct {
	userService *services.UserService
}

func NewCreateUserUseCase(userService *services.UserService) *CreateUserUseCase {
	return &CreateUserUseCase{
		userService: userService,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, input CreateUserInput) (*user.User, error) {
	return uc.userService.CreateUser(input.Name, input.Username, input.Password, user.MEMBER)
}
