package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OperationCard struct {
	ID             int
	Title          string
	ImageURL       *string
	BloodLossCoeff float64
}

type BLItem struct {
	OperationTitle  string
	OperationImage  *string
	BloodLossCoeff  float64
	AvgBloodLoss    int
	HbBefore        int
	HbAfter         int
	SurgeryDuration float64
	TotalBloodLoss  int
}

type BloodLossCalcVM struct {
	ID            int
	PatientHeight float64
	PatientWeight int
	Items         []BLItem
}

type OperationCardWithCart struct {
	ID             int
	Title          string
	ImageURL       *string
	BloodLossCoeff float64
	HasActiveCart  bool
}

func (h Handler) GetOperation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	op, err := h.Repository.GetOperation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "operation.html", gin.H{
		"operation": op,
	})
}

func (h Handler) GetBloodlosscalcByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	bl, err := h.Repository.GetBloodlosscalcByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Загрузим связанные записи bloodlosscalc_operations
	itemsDB, err := h.Repository.GetBloodlosscalcOperations(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Для каждого элемента загрузим операцию, сформируем BLItem
	var items []BLItem
	for _, itemDB := range itemsDB {
		op, errOp := h.Repository.GetOperation(itemDB.OperationID)
		if errOp != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, errOp)
			return
		}

		items = append(items, BLItem{
			OperationTitle:  op.Title,
			OperationImage:  op.ImageURL,
			BloodLossCoeff:  op.BloodLossCoeff,
			AvgBloodLoss:    op.AvgBloodLoss,
			HbBefore:        itemDB.HbBefore,
			HbAfter:         itemDB.HbAfter,
			SurgeryDuration: itemDB.SurgeryDuration,
			TotalBloodLoss:  itemDB.TotalBloodLoss,
		})
	}

	vm := BloodLossCalcVM{
		ID:            bl.ID,
		PatientHeight: bl.PatientHeight,
		PatientWeight: bl.PatientWeight,
		Items:         items,
	}

	ctx.HTML(http.StatusOK, "blood_loss_calc.html", gin.H{
		"bloodlosscalc": vm,
	})
}

func (h *Handler) GetOperationsWithRequestInfo(ctx *gin.Context) {
	operationsearch := ctx.Query("operationsearch")
	userID := h.getCurrentUserID(ctx)

	// Получаем операции
	var ops []OperationCard
	var err error

	if operationsearch == "" {
		all, e := h.Repository.GetOperations()
		err = e
		if err == nil {
			for _, o := range all {
				ops = append(ops, OperationCard{
					ID:             o.ID,
					Title:          o.Title,
					ImageURL:       o.ImageURL,
					BloodLossCoeff: o.BloodLossCoeff,
				})
			}
		}
	} else {
		found, e := h.Repository.GetOperationsByTitle(operationsearch)
		err = e
		if err == nil {
			for _, o := range found {
				ops = append(ops, OperationCard{
					ID:             o.ID,
					Title:          o.Title,
					ImageURL:       o.ImageURL,
					BloodLossCoeff: o.BloodLossCoeff,
				})
			}
		}
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// ПРОВЕРЯЕМ НАЛИЧИЕ АКТИВНОЙ ЗАЯВКИ
	currentRequest, err := h.Repository.GetCurrentBloodlosscalc(userID)
	hasActiveRequest := (err == nil)

	var serviceCount int64 = 0
	var currentRequestId int = 0

	if hasActiveRequest {
		currentRequestId = currentRequest.ID
		serviceCount = h.Repository.CountOperationsInBloodlosscalc(currentRequest.ID)
	}

	ctx.HTML(http.StatusOK, "operations.html", gin.H{
		"time":             time.Now().Format("15:04:05"),
		"operations":       ops,
		"operationsearch":  operationsearch,
		"hasActiveRequest": hasActiveRequest,
		"currentRequestId": currentRequestId,
		"serviceCount":     serviceCount,
	})
}

// POST /bloodlosscalc/add_operation - добавление операции в заявку
func (h *Handler) AddOperationToBloodlosscalc(ctx *gin.Context) {
	userID := h.getCurrentUserID(ctx)

	operationIDStr := ctx.PostForm("operation_id")
	operationID, _ := strconv.Atoi(operationIDStr)

	hbBefore, _ := strconv.Atoi(ctx.PostForm("hb_before"))
	hbAfter, _ := strconv.Atoi(ctx.PostForm("hb_after"))
	duration, _ := strconv.ParseFloat(ctx.PostForm("duration"), 64)
	totalLoss, _ := strconv.Atoi(ctx.PostForm("total_loss"))

	// Проверяем существование операции
	_, err := h.Repository.GetOperation(operationID)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// ПОЛУЧАЕМ ИЛИ СОЗДАЕМ ЗАЯВКУ
	bloodlosscalc, err := h.Repository.GetCurrentBloodlosscalc(userID)
	if err != nil {
		// Создаем новую заявку
		height, _ := strconv.ParseFloat(ctx.PostForm("height"), 64)
		weight, _ := strconv.Atoi(ctx.PostForm("weight"))

		bloodlosscalc, err = h.Repository.CreateBloodlosscalc(userID, height, weight)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	// Проверяем дубликаты и добавляем операцию
	if !h.Repository.OperationExistsInBloodlosscalc(bloodlosscalc.ID, operationID) {
		err = h.Repository.AddOperationToBloodlosscalc(bloodlosscalc.ID, operationID, hbBefore, hbAfter, duration, totalLoss)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	ctx.Redirect(http.StatusFound, "/")
}

// POST /bloodlosscalc/:id/delete - логическое удаление заявки через SQL
func (h *Handler) DeleteBloodlosscalc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	bloodlosscalcID, _ := strconv.Atoi(idStr)

	userID := h.getCurrentUserID(ctx)

	// Проверяем права доступа
	bloodlosscalc, err := h.Repository.GetBloodlosscalcByID(bloodlosscalcID)
	if err != nil || bloodlosscalc.CreatorID != userID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("нет прав для удаления"))
		return
	}

	// Удаляем через SQL
	err = h.Repository.DeleteBloodlosscalcSQL(bloodlosscalcID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/")
}

func (h *Handler) getCurrentUserID(ctx *gin.Context) int {
	return 1
}
