package handler

import (
	"blood_loss_calc/internal/app/repository"
	"net/http"
	"strconv"

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

// Главная страница со списком услуг
func (h *Handler) GetOperations(ctx *gin.Context) {
	var operations []repository.Operation
	var err error

	searchOperationSearch := ctx.Query("operationsearch")
	if searchOperationSearch == "" {
		operations, err = h.Repository.GetOperations()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		operations, err = h.Repository.GetOperationsByTitle(searchOperationSearch)
		if err != nil {
			logrus.Error(err)
		}
	}

	bloodlosscalc := h.Repository.GetBloodlosscalc()

	ctx.HTML(http.StatusOK, "operations.html", gin.H{
		"operations":      operations,
		"bloodlosscalc":   bloodlosscalc,
		"operationsearch": searchOperationSearch,
	})
}

// Страница услуги
func (h *Handler) GetOperation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	operation, err := h.Repository.GetOperation(id)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	ctx.HTML(http.StatusOK, "operation.html", gin.H{
		"operation": operation,
	})
}

// Страница заявки
func (h *Handler) GetBloodlosscalc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("invalid bloodlosscalc id:", err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	bloodlosscalc := h.Repository.GetBloodlosscalcByID(id)

	// Получаем полную информацию об услугах в заявке
	var bloodlosscalcOperations []repository.Operation
	for _, item := range bloodlosscalc.Items {
		operation, err := h.Repository.GetOperation(item.OperationID)
		if err == nil {
			bloodlosscalcOperations = append(bloodlosscalcOperations, operation)
		}
	}

	ctx.HTML(http.StatusOK, "blood_loss_calc.html", gin.H{
		"bloodlosscalc": bloodlosscalc,
		"operations":    bloodlosscalcOperations,
	})
}
