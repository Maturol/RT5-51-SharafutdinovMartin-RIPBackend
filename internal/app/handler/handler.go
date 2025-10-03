package handler

import (
	"blood_loss_calc/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.GetOperationsWithRequestInfo)
	router.GET("/operation/:id", h.GetOperation)
	router.GET("/bloodlosscalc/:id", h.GetBloodlosscalcByID)
	router.POST("/bloodlosscalc/add_operation", h.AddOperationToBloodlosscalc)
	router.POST("/bloodlosscalc/:id/delete", h.DeleteBloodlosscalc)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
