package main

import (
	"log"
	"saudemais-api/database"
	"saudemais-api/handlers"
	authMiddleware "saudemais-api/middleware"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Aviso: arquivo .env não encontrado.")
	}

	e := echo.New()

	database.Connect()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.POST("/register", handlers.RegisterPatient)
	e.POST("/login", handlers.Login)

	appointmentsGroup := e.Group("/appointments")
	appointmentsGroup.Use(authMiddleware.JWTMiddleware())

	appointmentsGroup.POST("", handlers.CreateAppointment)
	appointmentsGroup.GET("", handlers.ListAppointments)
	appointmentsGroup.DELETE("/:id", handlers.CancelAppointment)

	log.Println("Servidor iniciando na porta 3000...")

	if err := e.Start(":3000"); err != nil {
		log.Fatalf("Não foi possível iniciar o servidor: %v", err)
	}
}
