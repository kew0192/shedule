package main

import (
	"context"
	"log"
	"task_logic/db"
	"task_logic/handlers"
	"task_logic/middleware"
	"task_logic/objects"

	"github.com/gin-gonic/gin"
)

func main() {
	repo, err := db.NewPostgresRepository("postgres://postgres:postgres@localhost:5432/task_logic?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	if err := repo.InitTables(ctx); err != nil {
		log.Fatalf("Failed to init tables: %v", err)
	}
	log.Println("Database connected successfully")

	service := objects.NewService(repo)
	logicHandler := handlers.NewLogicHandler(service)

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.TokenMiddleware())

	// Классы
	r.POST("/class", logicHandler.CreateClassHandler)
	r.GET("/class/:name", logicHandler.GetClassHandler)
	r.GET("/classes", logicHandler.GetAllClassesHandler)
	r.DELETE("/class", logicHandler.DeleteClassHandler)

	// Учителя
	r.POST("/teacher", logicHandler.CreateTeacherHandler)
	r.GET("/teacher/:name", logicHandler.GetTeacherHandler)
	r.DELETE("/teacher", logicHandler.DeleteTeacherHandler)

	// Расписание
	r.GET("/schedule/:class", logicHandler.GetScheduleHandler)

	// Защищенные (уже под TokenMiddleware)
	r.POST("/lessons/add", logicHandler.AddLessonsHandler)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	log.Println("Server starting on :8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
