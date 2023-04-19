package db

import (
	"embed"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations
var fs embed.FS

type Migrator struct {
	dbUrl string
}

func NewMigrator(dbUrl string) *Migrator {
	return &Migrator{
		dbUrl: dbUrl,
	}
}

func (m *Migrator) Migrate() error {
	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	migrator, err := migrate.NewWithSourceInstance(
		"iofs",
		d,
		"mysql://"+m.dbUrl,
	)

	if err != nil {
		return err
	}

	err = migrator.Up()
	if err != nil {
		return err
	}

	return nil
}
