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

const (
	TypeUP   = "up"
	TypeDOWN = "down"
)

type Flags struct {
	MigrationType   string
	MigrationsPath  string
	MigrationsTable string
	Username        string
	Password        string
	Host            string
	Port            string
	DB              string
}

func main() {
	flags := parseFlags()
	validate := validator.New(validator.WithRequiredStructEnabled())
	validateFlags(validate, flags)

	databaseURL := buildDatabaseURL(flags)

	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", flags.MigrationsPath),
		databaseURL,
	)
	if err != nil {
		panic(err)
	}

	switch flags.MigrationType {
	case TypeUP:
		err = migrator.Up()
	case TypeDOWN:
		err = migrator.Down()
	default:
		panic("unknown migration type")
	}

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied successfully")
}

func parseFlags() *Flags {
	flags := &Flags{}
	flag.StringVar(&flags.MigrationType, "type", TypeUP, "migration type: down/up")
	flag.StringVar(&flags.MigrationsPath, "migration_path", "", "path to migrations")
	flag.StringVar(&flags.MigrationsTable, "migration_table", "migrations", "name of migrations table")
	flag.StringVar(&flags.Username, "username", "", "username")
	flag.StringVar(&flags.Password, "password", "", "password")
	flag.StringVar(&flags.Host, "host", "127.0.0.1", "host")
	flag.StringVar(&flags.Port, "port", "5432", "port")
	flag.StringVar(&flags.DB, "db", "", "db name")
	flag.Parse()
	return flags
}

func validateFlags(validate *validator.Validate, f *Flags) {
	Path(validate, f.MigrationsPath)
	Host(validate, f.Host)
	Required(validate, f.MigrationsTable, "migration_table")
	Required(validate, f.Username, "username")
	Required(validate, f.Password, "password")
	Required(validate, f.Port, "port")
	Required(validate, f.DB, "db")
}

func Type(validate *validator.Validate, migrationType string) {
	err := validate.Var(migrationType, "required,oneof=up down")
	if err != nil {
		panic("migration_type must be 'up' or 'down'")
	}
}

func Path(validate *validator.Validate, path string) {
	if err := validate.Var(path, "required"); err != nil {
		panic("migration_path is required")
	}
	if stat, err := os.Stat(path); os.IsNotExist(err) || !stat.IsDir() {
		panic("invalid migration_path: not a directory or does not exist")
	}
}

func Host(validate *validator.Validate, host string) {
	err := validate.Var(host, "required,hostname|ip")
	if err != nil {
		panic("invalid host")
	}
}

func Required(validate *validator.Validate, value string, name string) {
	if err := validate.Var(value, "required"); err != nil {
		panic(fmt.Sprintf("%s is required", name))
	}
}

func buildDatabaseURL(f *Flags) string {
	return (&url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(f.Username, f.Password),
		Host:     net.JoinHostPort(f.Host, f.Port),
		Path:     f.DB,
		RawQuery: "sslmode=disable",
	}).String()
}
