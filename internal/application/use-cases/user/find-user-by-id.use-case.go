package userusecases

import (
	"context"
	"grape-api/internal/domain/user"
	"uuid"
)

type FindUserByIDUseCase struct {
	userRepository user.IUserRepository
}

func NewFindUserByIDUseCase(userRepository user.IUserRepository) *FindUserByIDUseCase {
	return &FindUserByIDUseCase{userRepository: userRepository}
}

func (uc *FindUserByIDUseCase) Execute(ctx context.Context, id string) (*user.User, error) {
	idUUID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	return uc.userRepository.FindByID(ctx, idUUID)
}
