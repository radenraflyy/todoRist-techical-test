package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"todorist/config"
	"todorist/env"
	"todorist/server/router"

	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadEnv()
	env.GetEnv()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gin.SetMode(env.GinMode)
	port := env.Port
	app := gin.Default()

	// connect to DB
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	env.PgHost, env.PgPort, env.PgUser, env.PgPassword, env.PgDatabase)

	db := config.NewDB(ctx, psqlconn)
	defer db.Close()

	router.SetupRoutes(router.SetupRoutesConfig{
		Router: app,
		DB:     db,
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: app,
	}

	log.Printf("Server running on PORT %d", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-ctx.Done()
	log.Println("Shutting down server...")
	server.Shutdown(context.Background())
}
