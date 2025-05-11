package db

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func Open() (*sql.DB, error) {
	dsn := os.Getenv("MARIADB_DSN")
	return sql.Open("mysql", dsn)
}
