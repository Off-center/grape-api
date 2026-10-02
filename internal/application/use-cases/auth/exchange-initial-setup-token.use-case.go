package authusecases

import (
	"context"
	"errors"
	"grape-api/internal/apperror"
	"grape-api/internal/application/services"
	domainauth "grape-api/internal/domain/auth"
)

type ExchangeInitialSetupTokenInput struct {
	InitialSetupToken string
	IPAddress         string
	UserAgent         string
}

type ExchangeInitialSetupTokenUseCase struct {
	authRepository domainauth.IAuthRepository
	sessionService *services.SessionService
}

func NewExchangeInitialSetupTokenUseCase(
	authRepository domainauth.IAuthRepository,
	sessionService *services.SessionService,
) *ExchangeInitialSetupTokenUseCase {
	return &ExchangeInitialSetupTokenUseCase{
		authRepository: authRepository,
		sessionService: sessionService,
	}
}

func (uc *ExchangeInitialSetupTokenUseCase) Execute(ctx context.Context, input ExchangeInitialSetupTokenInput) (string, error) {
	if input.InitialSetupToken == "" {
		return "", apperror.New(apperror.CodeUnauthorized, "invalid initial setup token")
	}

	setupToken := uc.sessionService.GenerateSessionToken()
	setupTokenHash := uc.sessionService.HashSessionToken(setupToken)
	expiresAt, err := uc.sessionService.GenerateExpirationTime(domainauth.SETUP)
	if err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}

	setupSession, err := domainauth.NewSetupSession(
		setupTokenHash[:],
		input.IPAddress,
		input.UserAgent,
		expiresAt,
	)
	if err != nil {
		return "", apperror.New(apperror.CodeInternalError, "an internal error occurred")
	}

	initialTokenHash := uc.sessionService.HashSessionToken(input.InitialSetupToken)
	if err := uc.authRepository.ExchangeInitialSetupSession(ctx, initialTokenHash[:], &setupSession); err != nil {
		var appErr *apperror.Error
		if errors.As(err, &appErr) && appErr.Code == apperror.CodeUnauthorized {
			return "", apperror.New(apperror.CodeUnauthorized, "invalid initial setup token")
		}

		return "", apperror.Wrap(apperror.CodeInternalError, "an internal error occurred", err)
	}

	return setupToken, nil
}
