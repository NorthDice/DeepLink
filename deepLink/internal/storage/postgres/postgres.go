package mypostgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/NorthDice/DeepLink/internal/domain/models"
	"github.com/NorthDice/DeepLink/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConstraintUnique = "23505"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(dsn string) (*Storage, error) {
	const op = "new.postgres"

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pool, err := pgxpool.
		NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		db: pool,
	}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

// SaveUser returns id of last registered user
func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	const op = "saveUser.postgres"

	query := `INSERT INTO users (email, pass_hash) VALUES ($1, $2) RETURNING id`

	var userId int64

	err := s.db.QueryRow(ctx, query, email, passHash).Scan(&userId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}

// User returns user by email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "user.postgres"

	query := `SELECT id, email, pass_hash FROM users WHERE email = $1`

	var user models.User
	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.PassHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

// IsAdmin checks if a user with the given userID has administrative privileges in the system.
func (s *Storage) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "isAdmin.postgres"

	query := "SELECT user_id FROM admins WHERE user_id = $1"

	var isAdmin bool

	err := s.db.QueryRow(ctx, query, userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appId int32) (models.App, error) {
	const op = "app.postgres"

	query := `SELECT id, name, secret FROM apps WHERE id = $1`

	var app models.App

	err := s.db.QueryRow(ctx, query, appId).Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.App{}, storage.ErrAppNotFound
		}
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
