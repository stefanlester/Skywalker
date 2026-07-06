package skywalker

import (
	"database/sql"

	// pgx/v5's stdlib package registers the "pgx" database/sql driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

func (s *Skywalker) OpenDB(dbType, dsn string) (*sql.DB, error) {
	if dbType == "postgres" || dbType == "postgresql" {
		dbType = "pgx"
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
