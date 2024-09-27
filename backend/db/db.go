package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/nathanbizkit/article-management/env"
)

func New(e env.ENVer) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		e.DBUser(), e.DBPass(), e.DBHost(), e.DBPort(), e.DBName())

	var d *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		d, err = sql.Open("postgres", psqlInfo)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	if err = d.Ping(); err != nil {
		return nil, err
	}

	return d, nil
}
