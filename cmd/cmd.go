// Package cmd is an entry-point to application
package cmd

import (
	"log"
	"time"

	"commentTree/internal/api"
	"commentTree/internal/repository"
	"commentTree/internal/service"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
)

func StartApp() {
	// Reading configs
	appConfig := config.New()
	appConfig.EnableEnv("")
	if err := appConfig.LoadEnvFiles("./.env"); err != nil {
		log.Fatalf("Failed to load envs: %s\nExiting app...", err)
	}

	// Connecting to database
	dbConn := repository.ConnectWithRetries(appConfig, 5, 10*time.Second)
	defer func() {
		if err := dbConn.Master.Close(); err != nil {
			log.Println("Failed to close DB-conn correctly:", err)
		}
	}()

	// Creating Repository
	repo := repository.NewPostgresRepo(dbConn)

	// Creating Service
	svc := service.NewCommentService(repo)

	// Running DB migration
	if err := repository.Migrate(dbConn.Master, "./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %s", err)
	}

	// Creating Handlers
	handlers := api.NewCommentHandlers(svc)

	// Configuring server
	server := ginext.New("") // empty - debug mode, release - prod mode
	server.GET("/ping", handlers.SimplePinger)
	server.POST("/comments", handlers.Create)                    // создание (с указанием родительского)
	server.GET("/comments", handlers.GetAllRootComments)         // получение всех корневых комментариев с ?page=1&limit=20&sort=created_at
	server.GET("/comments/:id", handlers.GetCommentWithChildren) // получение коммента с id и всех его детей
	server.DELETE("/comments/:id", handlers.DeleteComment)       // удаление комментария и всех вложенных под ним
	server.GET("/comments/search", handlers.RunSearch)           // поиск
	server.Static("/web", "./internal/web")

	// Server launch
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
