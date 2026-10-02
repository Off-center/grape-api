package middlewares

import (
	"grape-api/internal/application/services"
	domainauth "grape-api/internal/domain/auth"
	"grape-api/internal/interfaces/http/authcontext"
	"grape-api/pkg"
	"net/http"
)

func Authentication(sessionService *services.SessionService, acceptedSessionType domainauth.SessionType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionToken := pkg.GetBearerToken(r)

			if sessionToken == "" {
				http.Error(w, "missing session token", http.StatusUnauthorized)
				return
			}

			result, err := sessionService.VerifySession(r.Context(), sessionToken, acceptedSessionType)

			if err != nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			needRefresh, err := sessionService.VerifyNeedRefresh(result.Session)

			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if needRefresh {
				err := sessionService.RefreshSession(r.Context(), result.Session.ID(), acceptedSessionType)

				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			}

			ctx := authcontext.WithUser(r.Context(), authcontext.User{
				ID:       result.UserInfo.ID,
				Username: result.UserInfo.Username,
				Role:     result.UserInfo.Role,
			})

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}
