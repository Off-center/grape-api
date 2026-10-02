package authusecases

import (
	"context"
	"grape-api/internal/application/services"
	domainauth "grape-api/internal/domain/auth"
)

type LogoutUseCase struct {
	authRepository domainauth.IAuthRepository
	sessionService *services.SessionService
}

func NewLogoutUseCase(authRepository domainauth.IAuthRepository, sessionService *services.SessionService) *LogoutUseCase {
	return &LogoutUseCase{authRepository: authRepository, sessionService: sessionService}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, token string) error {
	tokenHash := uc.sessionService.HashSessionToken(token)
	return uc.authRepository.DeleteSessionByTokenHashAndType(ctx, tokenHash[:], domainauth.SESSION)
}
