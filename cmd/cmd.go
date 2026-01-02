// Package cmd is an entry-point to application
package cmd

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"commentTree/internal/api"
	"commentTree/internal/repository"
	"commentTree/internal/service"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"
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

	// Creating Repository
	repo := repository.NewPostgresRepo(dbConn)

	// Creating Service
	svc := service.NewCommentService(repo)

	// Running DB migration
	repository.MigrateWithRetries(dbConn.Master, "./migrations", 5, 10*time.Second)

	// Creating Handlers
	handlers := api.NewCommentHandlers(svc)

	// Configuring server
	mode := appConfig.GetString("GIN_MODE")
	engine := ginext.New(mode)

	engine.GET("/ping", handlers.SimplePinger)
	engine.POST("/comments", handlers.Create)                    // создание (с указанием родительского)
	engine.GET("/comments", handlers.GetAllRootComments)         // получение всех корневых комментариев с поддержкой квери ?page=1&limit=20&sort=created_at&order=ascending
	engine.GET("/comments/:id", handlers.GetCommentWithChildren) // получение коммента с id и всех его детей
	engine.DELETE("/comments/:id", handlers.DeleteComment)       // удаление комментария и всех вложенных под ним
	engine.GET("/comments/search", handlers.RunSearch)           // поиск
	engine.Static("/web", "./internal/web")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: engine,
	}

	// Server launch
	go func() {
		log.Printf("Server running on http://localhost%s\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				log.Println("Server gracefully stopping...")
			default:
				log.Fatalf("Server stopped: %v", err)
			}
		}
	}()

	// Waiting for interruption to start Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	shutdown(srv, dbConn)
	log.Println("Exiting application...")
}

func shutdown(srv *http.Server, dbConn *dbpg.DB) {
	log.Println("Interrupt received!!! Starting shutdown sequence...")

	// 5 seconds to stop HTTP-server:
	ctx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped")
	}

	// Closing DB connection
	if err := dbConn.Master.Close(); err != nil {
		log.Println("Failed to close DB-conn correctly:", err)
		return
	}
	log.Println("DBconn closed")
}
