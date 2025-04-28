package database

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/raikh/calc_micro_final/internal/config"
)

var db *sqlx.DB

func InitDB(ctx context.Context, appCfg *config.Config) *sqlx.DB {
	// mysql & postgres not tested and written only for example
	switch dbType := appCfg.GetKey("APP_DB_TYPE"); dbType {
	case "mysql":
		db = initMysql(appCfg)
	case "postgres":
		db = initPostgres(appCfg)
	case "sqlite":
		dbName := fmt.Sprintf("file:%s/%s?mode=rwc", appCfg.GetKey("ROOT_DIR"), appCfg.GetKey("APP_DB_NAME"))
		db = openConn("sqlite3", dbName)
	default:
		log.Fatalf("Unsupported database type: %s\n", dbType)
	}

	err := db.PingContext(ctx)
	if err != nil {
		panic(err.Error())
	}

	createTables(ctx, db)

	return db
}

func initMysql(appCfg *config.Config) *sqlx.DB {
	cfg := mysql.Config{
		User:                 appCfg.GetKey("APP_DB_USER"),
		Passwd:               appCfg.GetKey("APP_DB_PASSWORD"),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", appCfg.GetKey("APP_DB_HOST"), appCfg.GetKey("APP_DB_PORT")),
		DBName:               appCfg.GetKey("APP_DB_NAME"),
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	return openConn("mysql", cfg.FormatDSN())
}

func initPostgres(appCfg *config.Config) *sqlx.DB {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		appCfg.GetKey("APP_DB_USER"),
		appCfg.GetKey("APP_DB_PASSWORD"),
		appCfg.GetKey("APP_DB_HOST"),
		appCfg.GetKey("APP_DB_PORT"),
		appCfg.GetKey("APP_DB_NAME"),
	)

	return openConn("postgres", connString)
	// pgx.Connect(context.Background(), connString)
}

func openConn(driverName string, dataSourceName string) *sqlx.DB {
	db, err := sqlx.Open(driverName, dataSourceName)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Connection to the database is successful")

	return db
}

func GetDB() *sqlx.DB {
	return db
}

func createTables(ctx context.Context, db *sqlx.DB) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
        password TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT NULL,
        deleted_at TIMESTAMP DEFAULT NULL
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL,
		expression TEXT NOT NULL,
        result double precision,
        status text NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT NULL,
        deleted_at TIMESTAMP DEFAULT NULL
	);`
		tasksTable = `
	CREATE TABLE IF NOT EXISTS tasks(
		id TEXT PRIMARY KEY,
		expression_id TEXT NOT NULL,
        arg1 double precision NULL,
        arg2 double precision NULL,
        operation text NOT NULL,
        operation_time integer,
        dependencies TEXT,
        result double precision,
        completed boolean,
        is_processing boolean,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT NULL,
        deleted_at TIMESTAMP DEFAULT NULL
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		log.Println("Error creating users table")
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		log.Printf("Error creating expressions table: %v", err)
		return err
	}
	if _, err := db.ExecContext(ctx, tasksTable); err != nil {
		log.Println("Error creating tasks table")
		return err
	}

	return nil
}
