package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"grape-api/internal/application/services"
	domainauth "grape-api/internal/domain/auth"
	"grape-api/internal/domain/instance"
	"grape-api/internal/persistence/pgx"
	"grape-api/internal/persistence/pgx/repositories"
	"io"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type initialSetupStore interface {
	IsInstanceInitialized(ctx context.Context) (bool, error)
	ReplaceInitialSetupSession(ctx context.Context, data *domainauth.Session) error
}

type postgresInitialSetupStore struct {
	db             *sql.DB
	authRepository domainauth.IAuthRepository
}

func (s *postgresInitialSetupStore) IsInstanceInitialized(ctx context.Context) (bool, error) {
	var initialized bool

	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM instances
			WHERE setup_status = $1
		)
	`, instance.INITIALIZED).Scan(&initialized)

	if err != nil {
		return false, fmt.Errorf("check instance setup status: %w", err)
	}

	return initialized, nil
}

func (s *postgresInitialSetupStore) ReplaceInitialSetupSession(ctx context.Context, data *domainauth.Session) error {
	return s.authRepository.ReplaceInitialSetupSession(ctx, data)
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("initial-setup-token", flag.ContinueOnError)
	flags.SetOutput(stderr)
	raw := flags.Bool("raw", false, "print only the generated token")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(stderr, "DATABASE_URL is required")
		return 1
	}

	db, err := pgx.NewPostgresDB(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(stderr, "connect to database: %v\n", err)
		return 1
	}
	defer db.Close()

	authRepository := repositories.NewPgxAuthRepository(db)
	store := &postgresInitialSetupStore{
		db:             db,
		authRepository: authRepository,
	}
	sessionService := services.NewSessionService(authRepository)

	if err := generateInitialSetupToken(ctx, store, sessionService, *raw, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	return 0
}

func generateInitialSetupToken(
	ctx context.Context,
	store initialSetupStore,
	sessionService *services.SessionService,
	raw bool,
	stdout io.Writer,
) error {
	initialized, err := store.IsInstanceInitialized(ctx)
	if err != nil {
		return err
	}
	if initialized {
		return fmt.Errorf("the instance is already initialized")
	}

	token := sessionService.GenerateSessionToken()
	tokenHash := sessionService.HashSessionToken(token)
	expiresAt, err := sessionService.GenerateExpirationTime(domainauth.INITIAL_SETUP)
	if err != nil {
		return fmt.Errorf("generate token expiration: %w", err)
	}

	session, err := domainauth.NewInitialSetupSession(tokenHash[:], expiresAt)
	if err != nil {
		return fmt.Errorf("create initial setup session: %w", err)
	}
	if err := store.ReplaceInitialSetupSession(ctx, &session); err != nil {
		return fmt.Errorf("persist initial setup session: %w", err)
	}

	if raw {
		_, err = fmt.Fprintln(stdout, token)
		return err
	}

	_, err = fmt.Fprintf(
		stdout,
		"Initial Setup Token generated successfully (valid for %s):\n%s\n",
		formatDuration(time.Until(expiresAt)),
		token,
	)
	return err
}

func formatDuration(duration time.Duration) string {
	minutes := int(duration.Round(time.Minute) / time.Minute)
	return fmt.Sprintf("%d minutes", minutes)
}
