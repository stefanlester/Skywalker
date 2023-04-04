package skywalker

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
)

func (s *Skywalker) MigrateUp(dsn string) error {
	m, err := migrate.New("file://"+s.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	} 

	defer m.Close()

	if err = m.Up(); err != nil {
		log.Println("Error running migration:", err)
		return err
	}

	return nil
}

func (s *Skywalker) MigrateDownAll(dsn string) error {
	m, err := migrate.New("file://" + s.RootPath + "/migrations", dsn)
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
	m, err := migrate.New("file://" + s.RootPath + "/migrations", dsn)
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
	m, err := migrate.New("file://" + s.RootPath + "/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil {
		return err
	}

	return nil
}