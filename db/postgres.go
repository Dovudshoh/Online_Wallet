package db

import (
	"database/sql"
	"fmt"
	"online_bank/config"

	_ "github.com/lib/pq"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии соединения: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}

	fmt.Println("Подключено к PostgreSQL!")
	return db, nil
}
