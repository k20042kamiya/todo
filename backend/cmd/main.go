package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"todo/auth"
	"todo/infrastructure/database"
	"todo/infrastructure/firebase"
	todopkg "todo/todo"
	userpkg "todo/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	db, err := database.NewDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying DB: %v", err)
	}
	defer sqlDB.Close()

	authClient, err := firebase.NewFirebaseAuth(context.Background())
	if err != nil {
		log.Fatalf("Failed to initialize Firebase Auth: %v", err)
	}

	txManager := database.NewTransactionManager(db)
	todoRepo := todopkg.NewRepository(db)
	userRepo := userpkg.NewRepository(db)
	todoUC := todopkg.NewUsecase(txManager, todoRepo)
	userUC := userpkg.NewUsecase(userRepo)
	todoHandler := todopkg.NewHandler(todoUC)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")
	api.Use(auth.Auth(authClient, userUC))
	api.GET("/todos", todoHandler.GetTodos)
	api.POST("/todos", todoHandler.CreateTodo)
	api.PUT("/todos/:id", todoHandler.UpdateTodo)
	api.DELETE("/todos/:id", todoHandler.DeleteTodo)

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}
