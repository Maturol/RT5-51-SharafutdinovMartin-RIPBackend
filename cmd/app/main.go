package main

import (
	"fmt"
	"os"

	"blood_loss_calc/internal/app/config"
	"blood_loss_calc/internal/app/dsn"
	"blood_loss_calc/internal/app/handler"
	"blood_loss_calc/internal/app/repository"
	"blood_loss_calc/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "blood_loss_calc/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Blood Loss Calculator API
// @version 1.0
// @description API для расчета кровопотери при операциях

// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@bloodloss.local

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT токен в формате: "Bearer {token}"

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	// Получаем Redis хост и порт из env
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6380")

	rep, errRep := repository.New(postgresString, redisHost, redisPort)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	hand := handler.NewHandler(rep)

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
