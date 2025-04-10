package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var username, password, host, port, db string
	var migrationsPath, migrationsTable string

	flag.StringVar(&migrationsPath, "migration_path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migration_table", "migrations", "name of migrations table")
	flag.StringVar(&username, "username", "", "username")
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&host, "host", "127.0.0.1", "host")
	flag.StringVar(&port, "port", "5432", "port")
	flag.StringVar(&db, "db", "", "db name")
	flag.Parse()

	validate := validator.New(validator.WithRequiredStructEnabled())
	validateMigrationsPath(validate, migrationsPath)
	validateHost(validate, host)
	validateRequired(validate, migrationsTable)
	validateRequired(validate, username)
	validateRequired(validate, password)
	validateRequired(validate, port)
	validateRequired(validate, db)

	databaseURL := &url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(username, password),
		Host:     net.JoinHostPort(host, port),
		Path:     db,
		RawQuery: "sslmode=disable",
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL.String(),
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied successfully")
}

func validateMigrationsPath(validate *validator.Validate, migrationsPath string) {
	err := validate.Var(migrationsPath, "required,dirpath")
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, e := range validationErrs {
				if e.Tag() == "required" {
					panic("migration path is required")
				}
				if e.Tag() == "dirpath" {
					panic("invalid migration path")
				}
			}
		}
		panic(err)
	}

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		panic(err)
	}
}

func validateHost(validate *validator.Validate, host string) {
	err := validate.Var(host, "required,hostname")
	if err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, e := range validationErrs {
				if e.Tag() == "required" {
					panic("host is required")
				}
				if e.Tag() == "hostname" {
					panic("invalid host")
				}
			}
		}
		panic(err)
	}
}

func validateRequired(validate *validator.Validate, value string) {
	err := validate.Var(value, "required")
	if err != nil {
		panic(err)
	}
}
