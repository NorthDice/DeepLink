package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/NorthDice/DeepLink/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"strconv"
)

var (
	IsEmptyString = ""
)
var (
	migrationsPath string
)

func main() {

	flag.StringVar(&migrationsPath, "migrationsDir", "", "path to the migrations directory")

	cfg := config.MustLoad()

	flag.Parse()

	if migrationsPath == IsEmptyString {
		panic("migrations path is required")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		strconv.Itoa(cfg.Database.Postgres.Port),
		cfg.Database.Postgres.Name,
	)

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to run")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied")

}
