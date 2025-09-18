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
func (h *Handler) GetServices(ctx *gin.Context) {
	var services []repository.Service
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		services, err = h.Repository.GetServices()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		services, err = h.Repository.GetServicesByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	cart := h.Repository.GetCart()

	ctx.HTML(http.StatusOK, "op_services.html", gin.H{
		"services": services,
		"cart":     cart,
		"query":    searchQuery,
	})
}

// Страница услуги
func (h *Handler) GetService(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	service, err := h.Repository.GetService(id)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/")
		return
	}

	ctx.HTML(http.StatusOK, "op_service.html", gin.H{
		"service": service,
	})
}

// Страница заявки
func (h *Handler) GetCart(ctx *gin.Context) {
	cart := h.Repository.GetCart()

	// Получаем полную информацию об услугах в заявке
	var cartServices []repository.Service
	for _, item := range cart.Items {
		service, err := h.Repository.GetService(item.ServiceID)
		if err == nil {
			cartServices = append(cartServices, service)
		}
	}

	ctx.HTML(http.StatusOK, "op_calc.html", gin.H{
		"cart":     cart,
		"services": cartServices,
	})
}
