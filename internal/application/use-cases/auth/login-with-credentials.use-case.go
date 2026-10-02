package authusecases

import (
	"context"
	"grape-api/internal/apperror"
	"grape-api/internal/application/services"
	domainauth "grape-api/internal/domain/auth"
	"grape-api/internal/domain/user"
	"grape-api/pkg"
)

type LoginWithCredentialsInput struct {
	Username  string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginWithCredentialsUseCase struct {
	authRepository domainauth.IAuthRepository
	userRepository user.IUserRepository
	sessionService *services.SessionService
}

func NewLoginWithCredentialsUseCase(
	authRepository domainauth.IAuthRepository,
	userRepository user.IUserRepository,
	sessionService *services.SessionService,
) *LoginWithCredentialsUseCase {
	return &LoginWithCredentialsUseCase{
		authRepository: authRepository,
		userRepository: userRepository,
		sessionService: sessionService,
	}
}

func (uc *LoginWithCredentialsUseCase) Execute(ctx context.Context, input LoginWithCredentialsInput) (string, error) {
	foundUser, err := uc.userRepository.FindByUsername(ctx, input.Username)

	if err != nil {
		return "", apperror.New(apperror.CodeUnauthorized, "invalid username or password")
	}

	account, err := uc.authRepository.FindAccountByUserIDAndIssuer(ctx, foundUser.ID(), domainauth.CREDENTIALS)
	if err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}

	isValid, err := pkg.VerifyPassword(input.Password, account.PasswordHash)
	if err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}
	if !isValid {
		return "", apperror.New(apperror.CodeUnauthorized, "invalid username or password")
	}

	token := uc.sessionService.GenerateSessionToken()
	tokenHash := uc.sessionService.HashSessionToken(token)
	expiresAt, err := uc.sessionService.GenerateExpirationTime(domainauth.SESSION)
	if err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}

	sessionData, err := domainauth.NewUserSession(
		foundUser.ID(),
		tokenHash[:],
		input.IPAddress,
		input.UserAgent,
		expiresAt,
	)
	if err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}

	if err := uc.authRepository.CreateSession(ctx, &sessionData); err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}

	return token, nil
}
