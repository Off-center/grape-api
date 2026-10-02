package userusecases

import (
	"grape-api/internal/application/services"
	"grape-api/internal/domain/user"
)

type CreateMasterUserInput struct {
	Name     string
	Username string
	Password string
}

type CreateMasterUserUseCase struct {
	userService *services.UserService
}

func NewCreateMasterUserUseCase(userService *services.UserService) *CreateMasterUserUseCase {
	return &CreateMasterUserUseCase{
		userService: userService,
	}
}

func (uc *CreateMasterUserUseCase) Execute(input CreateMasterUserInput) error {
	_, err := uc.userService.CreateUser(input.Name, input.Username, input.Password, user.MASTER)
	return err
}
