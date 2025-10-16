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
	api := router.Group("/api")

	// Операции
	api.GET("/operations", h.GetOperations)
	api.GET("/operations/:id", h.GetOperation)
	api.POST("/operations", h.CreateOperation)
	api.PUT("/operations/:id", h.UpdateOperation)
	api.DELETE("/operations/:id", h.DeleteOperation)
	api.POST("/operations/:id/image", h.UploadOperationImage)
	api.POST("/operations/:id/add_to_bloodlosscalc", h.AddOperationToBloodlosscalc)

	// Заявки
	api.GET("/bloodlosscalcs", h.GetBloodlosscalcs)
	api.GET("/bloodlosscalcs/:id", h.GetBloodlosscalcByID)
	api.PUT("/bloodlosscalcs/:id", h.UpdateBloodlosscalc)
	api.PUT("/bloodlosscalcs/:id/form", h.FormBloodlosscalc)
	api.PUT("/bloodlosscalcs/:id/complete", h.CompleteBloodlosscalc)
	api.DELETE("/bloodlosscalcs/:id", h.DeleteBloodlosscalc)

	// Корзина
	api.GET("/operationcart", h.GetOperationCartInfo)

	// Связь м-м
	api.DELETE("/bloodlosscalc_operations", h.RemoveOperationFromBloodlosscalc)
	api.PUT("/bloodlosscalc_operations", h.UpdateBloodlosscalcOperation)

	// Пользователи
	api.POST("/register", h.RegisterUser)
	api.GET("/user", h.GetUserProfile)
	api.PUT("/user", h.UpdateUserProfile)
	api.POST("/auth", h.AuthenticateUser)
	api.POST("/logout", h.LogoutUser)
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
