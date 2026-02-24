package config

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func ConnectDB() *sqlx.DB {

	databaseURL := os.Getenv("DATABASE_URL")

	var db *sqlx.DB
	var err error

	if databaseURL != "" {
		db, err = sqlx.Connect("postgres", databaseURL)
	} else {
		dsn := "host=" + os.Getenv("DB_HOST") +
			" port=" + os.Getenv("DB_PORT") +
			" user=" + os.Getenv("DB_USER") +
			" password=" + os.Getenv("DB_PASSWORD") +
			" dbname=" + os.Getenv("DB_NAME") +
			" sslmode=" + os.Getenv("DB_SSLMODE")

		db, err = sqlx.Connect("postgres", dsn)
	}

	if err != nil {
		log.Fatal("DB connection failed:", err)
	}

	log.Println("✅ Database connected!")
	return db
}
