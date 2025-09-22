package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// схема таблицы + индекс
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    date    CHAR(8)        NOT NULL,
    title   VARCHAR(256)   NOT NULL,
    comment TEXT           NOT NULL DEFAULT '',
    repeat  VARCHAR(128)   NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS scheduler_date_idx ON scheduler(date);
`

const defaultPath = "scheduler.db"

// Init открывает/создаёт файл базы, если файла нет — создаёт таблицу
func Init(dbFile string) error {
	if dbFile == "" {
		dbFile = defaultPath
	}

	_, statErr := os.Stat(dbFile)
	install := false
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			install = true
		} else {
			return fmt.Errorf("stat db file: %w", statErr)
		}
	}

	sqlDB, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	if err = sqlDB.Ping(); err != nil {
		return err
	}

	if install {
		if _, err = sqlDB.ExecContext(context.Background(), schema); err != nil {
			return err
		}
	}

	DB = sqlDB
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
