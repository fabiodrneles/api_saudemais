package database

import (
	"fmt"
	"log"
	"os"
	"saudemais-api/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=America/Sao_Paulo",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	var counts int64
	for {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Falha ao conectar ao banco de dados: %v. Tentando novamente em 5 segundos...", err)
		counts++
		if counts > 5 {
			log.Fatalf("Não foi possível conectar ao banco de dados após várias tentativas: %v", err)
		}
		time.Sleep(5 * time.Second)
	}

	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	err = DB.AutoMigrate(&models.Patient{}, &models.Appointment{})
	if err != nil {
		log.Fatalf("Falha ao migrar o schema do banco de dados: %v", err)
	}
	log.Println("Schema do banco de dados migrado com sucesso.")
}
