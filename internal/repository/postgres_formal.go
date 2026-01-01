package repository

import (
	"database/sql"
	"log"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
)

type PostgresRepo struct {
	db *dbpg.DB
}

func NewPostgresRepo(dbconn *dbpg.DB) CommentRepository {
	return &PostgresRepo{db: dbconn}
}

func ConnectWithRetries(appConfig *config.Config, retryCount int, idleTime time.Duration) *dbpg.DB {
	dbOptions := dbpg.Options{
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 10 * time.Minute,
	}
	dsnLink := appConfig.GetString("POSTGRES_DSN")
	var dbConn *dbpg.DB
	var err error

	for range retryCount {
		dbConn, err = dbpg.New(dsnLink, nil, &dbOptions)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to PGDB: %s\nWaiting %v before next retry...", err, idleTime)
		time.Sleep(idleTime)
	}

	if err != nil {
		log.Fatal("Failed to connect to DB. Exiting the app...")
	}

	return dbConn
}

func MigrateWithRetries(db *sql.DB, migrationsPath string, retries int, idle time.Duration) {
	for i := range retries {
		log.Printf("Migration try #%d...", i)
		err := Migrate(db, migrationsPath)
		if err == nil {
			break
		}
		switch i {
		case retries:
			log.Fatalln("Out of retries. Exiting...")
		default:
			log.Printf("Migration try #%d was unsuccessful. Waiting %v before next try...", i, idle)
			time.Sleep(idle)
		}
	}
}

func Migrate(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return err
	}

	sourceURL := "file://" + absPath
	log.Println("Running migrations from:", sourceURL)

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Database migrations applied successfully")
	return nil
}
