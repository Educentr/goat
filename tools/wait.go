package tools

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type MigrationWaitConfig interface {
	Check(rows *sql.Rows) error
	GetCycles() int
	GetInterval() time.Duration
	GetQuery() string
	SetCycles(c int)
	SetInterval(d time.Duration)
	SetQuery(s string)
}

type WaitConfigBase struct {
	Query    string
	Interval time.Duration
	Cycles   int
}

type PostgresqlWaitConfig struct {
	WaitConfigBase
	ExpectedVersion int
}

type MysqlWaitConfig struct {
	ExpectedVersion string
	WaitConfigBase
}

const (
	PsqlMigrationVersionQuery  = `SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`
	MysqlMigrationVersionQuery = `SELECT version FROM migrations ORDER BY version DESC LIMIT 1`
	defaultCycles              = 20
)

func (b *WaitConfigBase) GetQuery() string {
	return b.Query
}

func (b *WaitConfigBase) GetInterval() time.Duration {
	return b.Interval
}

func (b *WaitConfigBase) GetCycles() int {
	return b.Cycles
}

func (b *WaitConfigBase) SetQuery(s string) {
	b.Query = s
}

func (b *WaitConfigBase) SetInterval(d time.Duration) {
	b.Interval = d
}

func (b *WaitConfigBase) SetCycles(c int) {
	b.Cycles = c
}

func (p *PostgresqlWaitConfig) Check(rows *sql.Rows) error {
	var version int
	if err := rows.Scan(&version); err != nil {
		return err
	}

	if version < p.ExpectedVersion {
		return fmt.Errorf("migration step failed, version is wrong")
	}

	return nil
}

func PostgresqlDefault(expectedVersion int) MigrationWaitConfig {
	return &PostgresqlWaitConfig{
		WaitConfigBase: WaitConfigBase{
			Query:    PsqlMigrationVersionQuery,
			Interval: time.Second,
			Cycles:   defaultCycles,
		},
		ExpectedVersion: expectedVersion,
	}
}

func (p *MysqlWaitConfig) Check(rows *sql.Rows) error {
	var version string
	if err := rows.Scan(&version); err != nil {
		return err
	}

	if version != p.ExpectedVersion {
		return fmt.Errorf("migration step failed, version is wrong")
	}

	return nil
}

func MysqlDefault(expectedVersion string) MigrationWaitConfig {
	return &MysqlWaitConfig{
		WaitConfigBase: WaitConfigBase{
			Query:    MysqlMigrationVersionQuery,
			Interval: time.Second,
			Cycles:   defaultCycles,
		},
		ExpectedVersion: expectedVersion,
	}
}

// WaitMigration checks the success of migrations. Sql driver must be passed as the first argument. Default sql driver value is "postgres".
func WaitMigration(expectedVersion int, db *sql.DB, args ...func(cfg MigrationWaitConfig)) error {
	cfg := PostgresqlDefault(expectedVersion)
	for _, e := range args {
		e(cfg)
	}

	return WaitAppMigration(db, cfg)
}

func WaitAppMigration(db *sql.DB, cfg MigrationWaitConfig) error {
	for range cfg.GetCycles() {
		if err := func() error {
			rows, err := db.Query(cfg.GetQuery())
			if err != nil {
				return err
			}
			defer rows.Close()

			if !rows.Next() {
				return errors.New("empty migrations table") //nolint:err113
			}

			if rows.Err() != nil {
				return rows.Err()
			}

			checkErr := cfg.Check(rows)
			if checkErr != nil {
				return checkErr
			}

			return nil
		}(); err == nil {
			return nil
		}

		time.Sleep(cfg.GetInterval())
	}

	return errors.New("migration step failed, too long") //nolint:err113
}
