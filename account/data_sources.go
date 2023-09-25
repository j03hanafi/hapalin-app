package main

import (
	"fmt"
	"github.com/j03hanafi/hapalin-app/account/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"os"
)

type dataSources struct {
	DB *sqlx.DB
}

// InitDS establishes connections to fields in dataSources
func initDS() (*dataSources, error) {
	l := logger.Get()
	l.Info("Initializing data sources")

	// load env variables - we could pass these in,
	// but this is sort of just a top-level (main package)
	// helper function, so I'll just read them in here

	var (
		pgHost = os.Getenv("PG_HOST")
		pgPort = os.Getenv("PG_PORT")
		pgUser = os.Getenv("PG_USER")
		pgPass = os.Getenv("PG_PASS")
		pgDB   = os.Getenv("PG_DB")
		pgSSL  = os.Getenv("PG_SSL")
	)

	pgConnString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", pgHost, pgPort, pgUser, pgPass, pgDB, pgSSL)

	l.Info("Connecting to Postgres...")
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		l.Error("Error connecting to Postgres",
			zap.Error(err),
		)
		return nil, err
	}

	// Verify database connection is working
	if err = db.Ping(); err != nil {
		l.Error("Error pinging Postgres",
			zap.Error(err),
		)
		return nil, err
	}

	return &dataSources{
		DB: db,
	}, nil
}

// close to be used in graceful server shutdown
func (ds *dataSources) close() error {
	l := logger.Get()
	if err := ds.DB.Close(); err != nil {
		l.Error("Error closing Postgres connection",
			zap.Error(err),
		)
		return err
	}
	return nil
}
