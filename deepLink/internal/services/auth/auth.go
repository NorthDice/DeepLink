package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/NorthDice/DeepLink/internal/domain/models"
	"github.com/NorthDice/DeepLink/internal/lib/jwt"
	"github.com/NorthDice/DeepLink/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}
type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}
type UserProvider interface {
	User(ctx context.Context, email string) (user models.User, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type AppProvider interface {
	App(
		ctx context.Context,
		appId int32,
	) (app models.App, err error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId       = errors.New("invalid app id")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appId int32,
) (token string, err error) {
	const op = "auth.Login"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("logging in")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found")

			return "", fmt.Errorf("%w: %s", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user")

		return "", fmt.Errorf("%w: %s", err, op)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials")

		return "", fmt.Errorf("%w: %s", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found")

			return "", fmt.Errorf("%w: %s", op, ErrInvalidAppId)
		}
		return "", fmt.Errorf("%w: %s", err, op)
	}

	log.Info("user logged in")

	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token")

		return "", fmt.Errorf("%w: %s", err, op)
	}

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash")

		return 0, fmt.Errorf("%w: %s", err, op)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)

	log.Info("User id ", id)

	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists")

			return 0, fmt.Errorf("%w: %s", ErrUserAlreadyExists, op)
		}
		log.Error("failed to save user")

		return 0, fmt.Errorf("%w: %s", err, op)
	}

	log.Info("user registered")

	return id, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (isAdmin bool, err error) {
	const op = "auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err = a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found")

			return false, fmt.Errorf("%w: %s", op, ErrInvalidCredentials)
		}

		return false, fmt.Errorf("%w: %s", err, op)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
