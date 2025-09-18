package api

import (
	"blood_loss_calc/internal/app/handler"
	"blood_loss_calc/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Server start up")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/", handler.GetServices)
	r.GET("/service/:id", handler.GetService)
	r.GET("/cart", handler.GetCart)

	r.Run()
	log.Println("Server down")
}
