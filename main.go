package main

import (
	"task_auth/DB"
	"task_auth/auth"
	"task_auth/handlers"
	"task_auth/middleware"
	"task_auth/users"

	"github.com/gin-gonic/gin"
)

func main() {
	// Подключение к БД
	db, err := DB.OpenPOSTGRESQL("postgres://postgres:postgres@localhost:5432/task_auth?sslmode=disable")
	if err != nil {
		panic(err)
	}

	// Инициализация сервисов
	userService := users.NewService(db)
	authService := auth.NewAuthService(userService)
	authHandler := handlers.NewAuthHandler(authService)

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.POST("/login", authHandler.LoginHandler)
	r.POST("/refresh", authHandler.RefreshHandler)

	admin := r.Group("/admin")
	admin.Use(middleware.AdminMiddleware(db))
	{
		admin.POST("/teacher", authHandler.CreateTeacherHandler)
		admin.GET("/users", authHandler.GetAllUsersHandler)
	}

	r.Run(":8080")
}
