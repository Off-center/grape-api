package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	authusecases "grape-api/internal/application/use-cases/auth"
	authhttpdtos "grape-api/internal/interfaces/http/handlers/dtos/auth"
	customvalidations "grape-api/internal/interfaces/http/handlers/dtos/custom-validations"
	"grape-api/pkg"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	exchangeInitialSetupToken *authusecases.ExchangeInitialSetupTokenUseCase
	loginWithCredentials      *authusecases.LoginWithCredentialsUseCase
	logout                    *authusecases.LogoutUseCase
}

func NewAuthHandler(
	exchangeInitialSetupToken *authusecases.ExchangeInitialSetupTokenUseCase,
	loginWithCredentials *authusecases.LoginWithCredentialsUseCase,
	logout *authusecases.LogoutUseCase,
) *AuthHandler {
	return &AuthHandler{
		exchangeInitialSetupToken: exchangeInitialSetupToken,
		loginWithCredentials:      loginWithCredentials,
		logout:                    logout,
	}
}

func (h *AuthHandler) ExchangeInitialSetupToken(w http.ResponseWriter, r *http.Request) {
	token := pkg.GetBearerToken(r)
	ipAddress := pkg.RemoteIPAddress(r.RemoteAddr)
	userAgent := r.UserAgent()

	if token == "" {
		pkg.WriteError(w, http.StatusUnauthorized, "missing initial setup token")
		return
	}

	input := authusecases.ExchangeInitialSetupTokenInput{
		InitialSetupToken: token,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
	}

	setupToken, err := h.exchangeInitialSetupToken.Execute(r.Context(), input)

	if err != nil {
		pkg.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	pkg.WriteWithStatus(w, http.StatusCreated, authhttpdtos.ExchangeInitialSetupTokenResponse{Token: setupToken})
}

func (h *AuthHandler) LoginWithCredentials(w http.ResponseWriter, r *http.Request) {
	var reqBody authhttpdtos.LoginWithCredentialsRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&reqBody); err != nil {
		pkg.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var extraValue any
	if err := decoder.Decode(&extraValue); !errors.Is(err, io.EOF) {
		pkg.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reqBody.Email = strings.TrimSpace(reqBody.Email)

	validate := validator.New(
		validator.WithRequiredStructEnabled(),
	)

	if err := validate.RegisterValidation("password", customvalidations.Password); err != nil {
		panic(err)
	}

	if err := validate.Struct(reqBody); err != nil {
		fmt.Println("Validation error:", err)
		pkg.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := authusecases.LoginWithCredentialsInput{
		Username:  reqBody.Email,
		Password:  reqBody.Password,
		IPAddress: pkg.RemoteIPAddress(r.RemoteAddr),
		UserAgent: r.UserAgent(),
	}

	token, err := h.loginWithCredentials.Execute(r.Context(), input)

	if err != nil {
		pkg.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	pkg.WriteWithStatus(w, http.StatusCreated, authhttpdtos.LoginResponse{Token: token})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := pkg.GetBearerToken(r)

	if token == "" {
		pkg.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.logout.Execute(r.Context(), token); err != nil {
		pkg.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
