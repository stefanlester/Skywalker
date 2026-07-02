package skywalker

import (
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// migrationSourceURL builds the file:// source URL for golang-migrate.
// RootPath is converted to forward slashes so the URL parses on Windows.
func (s *Skywalker) migrationSourceURL() string {
	return "file://" + filepath.ToSlash(s.RootPath) + "/migrations"
}

func (s *Skywalker) MigrateUp(dsn string) error {
	m, err := migrate.New(s.migrationSourceURL(), dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("Error running migration:", err)
		return err
	}
	return nil
}

func (s *Skywalker) MigrateDownAll(dsn string) error {
	m, err := migrate.New(s.migrationSourceURL(), dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		return err
	}

	return nil
}

func (s *Skywalker) Steps(n int, dsn string) error {
	m, err := migrate.New(s.migrationSourceURL(), dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}

	return nil
}

func (s *Skywalker) MigrateForce(dsn string) error {
	m, err := migrate.New(s.migrationSourceURL(), dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}