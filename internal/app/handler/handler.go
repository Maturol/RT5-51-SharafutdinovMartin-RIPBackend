package handler

import (
	"blood_loss_calc/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorResponse структура для ошибок API
// @Description Стандартный ответ с ошибкой
type ErrorResponse struct {
	Status      string `json:"status" example:"error"`
	Description string `json:"description" example:"Error description"`
}

// MessageResponse структура для простых сообщений
// @Description Стандартный ответ с сообщением
type MessageResponse struct {
	Message string `json:"message" example:"Success message"`
}

// SuccessResponse структура для успешных операций
// @Description Стандартный успешный ответ
type SuccessResponse struct {
	Status string      `json:"status" example:"success"`
	Data   interface{} `json:"data"`
}

// BloodlosscalcResponse структура для ответа с заявками
// @Description Ответ со списком заявок
type BloodlosscalcResponse struct {
	ID             int      `json:"id"`
	Status         string   `json:"status"`
	CreatedAt      string   `json:"created_at"`
	FormedAt       *string  `json:"formed_at"`
	CompletedAt    *string  `json:"completed_at"`
	PatientHeight  *float64 `json:"patient_height"`
	PatientWeight  *int     `json:"patient_weight"`
	CreatorLogin   string   `json:"creator_login"`
	ModeratorLogin *string  `json:"moderator_login"`
	ServiceCount   *int     `json:"service_count,omitempty"`
}

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	// CORS middleware
	router.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		ctx.Next()
	})

	// Публичные routes (доступны без авторизации)
	public := router.Group("/api")
	{
		public.POST("/auth", h.AuthenticateUser)
		public.POST("/register", h.RegisterUser)
		public.GET("/operations", h.GetOperations)
		public.GET("/operations/:id", h.GetOperation)
	}

	// Защищенные routes (требуют JWT)
	auth := router.Group("api")
	auth.Use(h.AuthMiddleware())
	{
		// User routes
		auth.GET("/user", h.GetUserProfile)
		auth.PUT("/user", h.UpdateUserProfile)
		auth.POST("/logout", h.LogoutUser)

		// Cart & user bloodlosscalcs (доступны всем авторизованным)
		auth.GET("/operationcart", h.GetOperationCartInfo)
		auth.POST("/operations/:id/add_to_bloodlosscalc", h.AddOperationToBloodlosscalc)
		auth.GET("/bloodlosscalcs", h.GetBloodlosscalcs)
		auth.GET("/bloodlosscalcs/:id", h.GetBloodlosscalcByID)
		auth.PUT("/bloodlosscalcs/:id", h.UpdateBloodlosscalc)
		auth.DELETE("/bloodlosscalcs/:id", h.DeleteBloodlosscalc)
		auth.DELETE("/bloodlosscalc_operations", h.RemoveOperationFromBloodlosscalc)
		auth.PUT("/bloodlosscalc_operations", h.UpdateBloodlosscalcOperation)
		auth.PUT("/bloodlosscalcs/:id/form", h.FormBloodlosscalc)
	}

	// Moderator only routes
	moderator := auth.Group("")
	moderator.Use(h.RequireModerator())
	{
		moderator.PUT("/bloodlosscalcs/:id/complete", h.CompleteBloodlosscalc)
		moderator.POST("/operations", h.CreateOperation)
		moderator.PUT("/operations/:id", h.UpdateOperation)
		moderator.DELETE("/operations/:id", h.DeleteOperation)
		moderator.POST("/operations/:id/image", h.UploadOperationImage)
	}
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

// Вспомогательный метод для получения текущего user_id
func (h *Handler) getCurrentUserID(ctx *gin.Context) int {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0
	}
	return userID.(int)
}
