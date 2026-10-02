package instanceusecases

import (
	"context"
	"time"

	"grape-api/internal/apperror"
	"grape-api/internal/domain/instance"
	"grape-api/internal/domain/user"
)

type UpdateInstanceInput struct {
	Name               string
	Description        string
	RegistrationMethod instance.REGISTRATION_METHOD
}

type UpdateInstanceUseCase struct {
	userRepository     user.IUserRepository
	instanceRepository instance.IInstanceRepository
}

func NewUpdateInstanceUseCase(instanceRepository instance.IInstanceRepository, userRepository user.IUserRepository) *UpdateInstanceUseCase {
	return &UpdateInstanceUseCase{
		instanceRepository: instanceRepository,
		userRepository:     userRepository,
	}
}

func (uc *UpdateInstanceUseCase) Execute(ctx context.Context, data UpdateInstanceInput) error {
	masterUsers, err := uc.userRepository.ListByRole(ctx, user.MASTER)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternalError, "list master users", err)
	}

	if len(masterUsers) == 0 {
		return apperror.New(apperror.CodeNotFound, "master user not found")
	}

	if len(masterUsers) > 1 {
		return apperror.New(apperror.CodeConflict, "multiple master users found")
	}

	masterUser := masterUsers[0]
	if masterUser == nil {
		return apperror.New(apperror.CodeInternalError, "invalid master user")
	}

	instances, err := uc.instanceRepository.List(ctx)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternalError, "list instances", err)
	}

	if len(instances) == 0 {
		return apperror.New(apperror.CodeNotFound, "instance not found")
	}

	if len(instances) > 1 {
		return apperror.New(apperror.CodeConflict, "multiple instances found")
	}

	currentInstance := instances[0]
	if currentInstance == nil {
		return apperror.New(apperror.CodeInternalError, "invalid instance")
	}

	if currentInstance.SetupStatus != instance.UNINITIALIZED {
		return apperror.New(apperror.CodeConflict, "instance is already initialized")
	}

	switch data.RegistrationMethod {
	case instance.PUBLIC, instance.APPROVAL, instance.CLOSED:
	default:
		return apperror.New(apperror.CodeValidation, "invalid registration method")
	}

	currentInstance.Name = data.Name
	currentInstance.Description = data.Description
	currentInstance.RegistrationMethod = data.RegistrationMethod
	currentInstance.MasterUserID = masterUser.ID()
	currentInstance.SetupStatus = instance.INITIALIZED
	currentInstance.UpdatedAt = time.Now()

	if err := uc.instanceRepository.Update(ctx, currentInstance); err != nil {
		return apperror.Wrap(apperror.CodeInternalError, "update instance", err)
	}

	return nil
}
