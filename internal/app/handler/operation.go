package handler

import (
	"blood_loss_calc/internal/app/ds"
	"context"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	HbBefore        *int
	HbAfter         *int
	SurgeryDuration *float64
	TotalBloodLoss  *int
}

type BloodLossCalcVM struct {
	ID            int
	PatientHeight *float64
	PatientWeight *int
	Items         []BLItem
}

type OperationCardWithCart struct {
	ID             int
	Title          string
	ImageURL       *string
	BloodLossCoeff float64
	HasActiveCart  bool
}

// formatDate форматирует дату в формат DD.MM.YYYY
func formatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

// formatDatePtr форматирует указатель на дату в формат DD.MM.YYYY
func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format("02.01.2006")
	return &formatted
}

func (h Handler) GetOperation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	operation, err := h.Repository.GetOperation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, operation)
}

func (h Handler) GetBloodlosscalcByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, _ := strconv.Atoi(idStr)

	bl, err := h.Repository.GetBloodlosscalcByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	itemsDB, err := h.Repository.GetBloodlosscalcOperations(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

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

	// Формируем ответ с форматированными датами
	type BloodlosscalcDetailResponse struct {
		ID            int      `json:"id"`
		Status        string   `json:"status"`
		CreatedAt     string   `json:"created_at"`
		FormedAt      *string  `json:"formed_at"`
		CompletedAt   *string  `json:"completed_at"`
		PatientHeight *float64 `json:"patient_height"`
		PatientWeight *int     `json:"patient_weight"`
		Creator       string   `json:"creator"`
		Moderator     *string  `json:"moderator"`
		Items         []BLItem `json:"items"`
		ServiceCount  *int     `json:"service_count,omitempty"` // Количество услуг (только для завершенных)
	}

	response := BloodlosscalcDetailResponse{
		ID:            bl.ID,
		Status:        bl.Status,
		CreatedAt:     formatDate(bl.CreatedAt),
		FormedAt:      formatDatePtr(bl.FormedAt),
		CompletedAt:   formatDatePtr(bl.CompletedAt),
		PatientHeight: bl.PatientHeight,
		PatientWeight: bl.PatientWeight,
		Creator:       bl.Creator.Username,
		Items:         items,
		Moderator:     nil,
	}

	if bl.ModeratorID != nil && bl.Moderator.UserID != 0 {
		response.Moderator = &bl.Moderator.Username
	}

	// Добавляем количество услуг для завершенных заявок
	if bl.Status == "завершена" {
		count := int(h.Repository.CountOperationsInBloodlosscalc(bl.ID))
		response.ServiceCount = &count
	}

	ctx.JSON(http.StatusOK, response)
}

// POST /api/operations - добавление операции
func (h *Handler) CreateOperation(ctx *gin.Context) {
	var operation struct {
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		BloodLossCoeff float64 `json:"blood_loss_coeff"`
		AvgBloodLoss   int     `json:"avg_blood_loss"`
	}

	if err := ctx.BindJSON(&operation); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	newOperation := &ds.Operation{
		Title:          operation.Title,
		Description:    operation.Description,
		Status:         "активна",
		BloodLossCoeff: operation.BloodLossCoeff,
		AvgBloodLoss:   operation.AvgBloodLoss,
	}

	err := h.Repository.CreateOperation(newOperation)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, newOperation)
}

// PUT /api/operations/:id - изменение операции
func (h *Handler) UpdateOperation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем существование операции
	_, err = h.Repository.GetOperation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Создаем структуру для частичного обновления
	var updateData struct {
		Title          *string  `json:"title"`
		Description    *string  `json:"description"`
		Status         *string  `json:"status"`
		ImageURL       *string  `json:"image_url"`
		BloodLossCoeff *float64 `json:"blood_loss_coeff"`
		AvgBloodLoss   *int     `json:"avg_blood_loss"`
	}

	if err := ctx.BindJSON(&updateData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Обновляем только переданные поля
	updates := make(map[string]interface{})

	if updateData.Title != nil {
		updates["title"] = *updateData.Title
	}
	if updateData.Description != nil {
		updates["description"] = *updateData.Description
	}
	if updateData.Status != nil {
		updates["status"] = *updateData.Status
	}
	if updateData.ImageURL != nil {
		updates["image_url"] = *updateData.ImageURL
	}
	if updateData.BloodLossCoeff != nil {
		updates["blood_loss_coeff"] = *updateData.BloodLossCoeff
	}
	if updateData.AvgBloodLoss != nil {
		updates["avg_blood_loss"] = *updateData.AvgBloodLoss
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
		return
	}

	// Выполняем частичное обновление
	err = h.Repository.PartialUpdateOperation(id, updates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Возвращаем обновленную операцию
	updatedOperation, err := h.Repository.GetOperation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, updatedOperation)
}

// DELETE /api/operations/:id - удаление операции
func (h *Handler) DeleteOperation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteOperation(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Операция удалена"})
}

// GET /api/operations - список операций с фильтрацией
func (h *Handler) GetOperations(ctx *gin.Context) {
	titleFilter := ctx.Query("title")
	statusFilter := ctx.Query("status")

	var operations []ds.Operation
	var err error

	if titleFilter != "" {
		operations, err = h.Repository.GetOperationsByTitle(titleFilter)
	} else {
		operations, err = h.Repository.GetOperations()
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if statusFilter != "" {
		var filtered []ds.Operation
		for _, op := range operations {
			if op.Status == statusFilter {
				filtered = append(filtered, op)
			}
		}
		operations = filtered
	}

	ctx.JSON(http.StatusOK, gin.H{
		"operations": operations,
	})
}

// POST /api/operations/:id/image - добавление изображения к операции
func (h *Handler) UploadOperationImage(ctx *gin.Context) {
	idStr := ctx.Param("id")
	operationID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid operation ID"))
		return
	}

	operation, err := h.Repository.GetOperation(operationID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("operation not found"))
		return
	}

	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("failed to get file: %v", err))
		return
	}
	defer file.Close()

	if !isImageFile(header.Filename) {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("file must be an image (jpg, jpeg, png, gif)"))
		return
	}

	if header.Size > 10*1024*1024 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("file size must be less than 10MB"))
		return
	}

	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	fileName := generateFileName(fileExt)

	imageURL, err := h.Repository.UploadFile(ctx, fileName, file, header.Size)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to upload image: %v", err))
		return
	}

	if operation.ImageURL != nil && *operation.ImageURL != "" {
		oldFileName := getFileNameFromURL(*operation.ImageURL)
		if oldFileName != "" {
			go func() {
				deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				h.Repository.DeleteFile(deleteCtx, oldFileName)
			}()
		}
	}

	err = h.Repository.UpdateOperationImage(operationID, &imageURL)
	if err != nil {
		go func() {
			deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			h.Repository.DeleteFile(deleteCtx, fileName)
		}()

		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to update operation: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Image uploaded successfully",
		"image_url": imageURL,
	})
}

// GET /api/bloodlosscalcs - список заявок с фильтрацией
func (h *Handler) GetBloodlosscalcs(ctx *gin.Context) {
	statusFilter := ctx.Query("status")
	dateFromStr := ctx.Query("date_from")
	dateToStr := ctx.Query("date_to")

	var dateFrom, dateTo *time.Time

	if dateFromStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			dateFrom = &parsed
		}
	}
	if dateToStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateToStr); err == nil {
			dateTo = &parsed
		}
	}

	bloodlosscalcs, err := h.Repository.GetBloodlosscalcsFiltered(statusFilter, dateFrom, dateTo)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Структура для ответа с форматированными датами
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
		ServiceCount   *int     `json:"service_count,omitempty"` // Количество услуг (только для завершенных)
	}

	var response []BloodlosscalcResponse
	for _, bl := range bloodlosscalcs {
		item := BloodlosscalcResponse{
			ID:             bl.ID,
			Status:         bl.Status,
			CreatedAt:      formatDate(bl.CreatedAt),
			FormedAt:       formatDatePtr(bl.FormedAt),
			CompletedAt:    formatDatePtr(bl.CompletedAt),
			PatientHeight:  bl.PatientHeight,
			PatientWeight:  bl.PatientWeight,
			CreatorLogin:   bl.Creator.Username,
			ModeratorLogin: nil,
		}

		if bl.ModeratorID != nil && bl.Moderator.UserID != 0 {
			item.ModeratorLogin = &bl.Moderator.Username
		}

		// Добавляем количество услуг для завершенных заявок
		if bl.Status == "завершена" {
			count := int(h.Repository.CountOperationsInBloodlosscalc(bl.ID))
			item.ServiceCount = &count
		}

		response = append(response, item)
	}

	ctx.JSON(http.StatusOK, response)
}

// PUT /api/bloodlosscalc/:id - изменение заявки
func (h *Handler) UpdateBloodlosscalc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем существование заявки
	_, err = h.Repository.GetBloodlosscalcByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	var updateData struct {
		PatientHeight *float64 `json:"patient_height"`
		PatientWeight *int     `json:"patient_weight"`
	}

	if err := ctx.BindJSON(&updateData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Создаем карту обновлений
	updates := make(map[string]interface{})

	if updateData.PatientHeight != nil {
		updates["patient_height"] = updateData.PatientHeight
	}
	if updateData.PatientWeight != nil {
		updates["patient_weight"] = updateData.PatientWeight
	}

	// Если нет полей для обновления
	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
		return
	}

	err = h.Repository.UpdateBloodlosscalc(id, updates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка обновлена"})
}

// POST /api/operations/:id/add_to_bloodlosscalc - добавление операции в заявку
func (h *Handler) AddOperationToBloodlosscalc(ctx *gin.Context) {
	operationIDStr := ctx.Param("id")
	operationID, err := strconv.Atoi(operationIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid operation ID"))
		return
	}

	userID := h.getCurrentUserID(ctx)

	// Получаем операцию по ID из URL
	operation, err := h.Repository.GetOperation(operationID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("operation not found"))
		return
	}

	var requestData struct {
		PatientHeight   *float64 `json:"patient_height"`
		PatientWeight   *int     `json:"patient_weight"`
		HbBefore        *int     `json:"hb_before"`
		HbAfter         *int     `json:"hb_after"`
		SurgeryDuration *float64 `json:"surgery_duration"`
		TotalBloodLoss  *int     `json:"total_blood_loss"`
	}

	// Биндим данные, но даже если тело пустое - продолжаем
	if ctx.Request.ContentLength > 0 {
		if err := ctx.BindJSON(&requestData); err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
	}

	var bloodlosscalc ds.Bloodlosscalc
	var isNewRequest bool

	// Пытаемся получить текущую заявку пользователя
	currentBloodlosscalc, err := h.Repository.GetCurrentBloodlosscalc(userID)
	if err != nil {
		// Если нет текущей заявки - создаем новую ПУСТУЮ заявку
		bloodlosscalc, err = h.Repository.CreateBloodlosscalc(userID, requestData.PatientHeight, requestData.PatientWeight)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		isNewRequest = true
	} else {
		// Используем существующую заявку
		bloodlosscalc = currentBloodlosscalc
		isNewRequest = false

		// Если переданы новые данные роста/веса - обновляем заявку
		if requestData.PatientHeight != nil || requestData.PatientWeight != nil {
			updates := make(map[string]interface{})
			if requestData.PatientHeight != nil {
				updates["patient_height"] = requestData.PatientHeight
			}
			if requestData.PatientWeight != nil {
				updates["patient_weight"] = requestData.PatientWeight
			}
			err = h.Repository.UpdateBloodlosscalc(bloodlosscalc.ID, updates)
			if err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
		}
	}

	// Проверяем дубликаты
	if h.Repository.OperationExistsInBloodlosscalc(bloodlosscalc.ID, operationID) {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("operation already exists in bloodlosscalc"))
		return
	}

	// Добавляем операцию в заявку
	err = h.Repository.AddOperationToBloodlosscalc(
		bloodlosscalc.ID,
		operationID,
		requestData.HbBefore,
		requestData.HbAfter,
		requestData.SurgeryDuration,
		requestData.TotalBloodLoss,
	)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Получаем актуальное количество операций в заявке
	serviceCount := h.Repository.CountOperationsInBloodlosscalc(bloodlosscalc.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Operation added to bloodlosscalc successfully",
		"bloodlosscalc": gin.H{
			"id":             bloodlosscalc.ID,
			"status":         bloodlosscalc.Status,
			"patient_height": bloodlosscalc.PatientHeight,
			"patient_weight": bloodlosscalc.PatientWeight,
		},
		"operation": gin.H{
			"id":               operation.ID,
			"title":            operation.Title,
			"blood_loss_coeff": operation.BloodLossCoeff,
			"avg_blood_loss":   operation.AvgBloodLoss,
			"image_url":        operation.ImageURL,
		},
		"operation_data": gin.H{
			"hb_before":        requestData.HbBefore,
			"hb_after":         requestData.HbAfter,
			"surgery_duration": requestData.SurgeryDuration,
			"total_blood_loss": requestData.TotalBloodLoss,
		},
		"service_count":  serviceCount,
		"is_new_request": isNewRequest,
	})
}

// PUT /api/bloodlosscalc/:id/form - формирование заявки
func (h *Handler) FormBloodlosscalc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := 1 // id модератора

	bloodlosscalc, err := h.Repository.GetBloodlosscalcByID(id)
	if err != nil || bloodlosscalc.CreatorID == userID {
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("нет прав для формирования заявки"))
		return
	}

	if bloodlosscalc.Status != "черновик" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("можно формировать только черновики"))
		return
	}

	if bloodlosscalc.PatientHeight == nil || bloodlosscalc.PatientWeight == nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("для формирования укажите рост и вес пациента"))
		return
	}

	count := h.Repository.CountOperationsInBloodlosscalc(id)
	if count == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("добавьте хотя бы одну операцию в заявку"))
		return
	}

	now := time.Now()
	err = h.Repository.UpdateBloodlosscalcStatus(id, "сформирована", &now, nil)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка сформирована"})
}

// PUT /api/bloodlosscalc/:id/complete - завершение заявки
func (h *Handler) CompleteBloodlosscalc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	userID := 1 //id модератора

	if userID != 1 { // временная заглушка для модератора
		h.errorHandler(ctx, http.StatusForbidden, fmt.Errorf("только модератор может завершать заявки"))
		return
	}

	bloodlosscalc, err := h.Repository.GetBloodlosscalcByID(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if bloodlosscalc.Status != "сформирована" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("можно завершать только сформированные заявки"))
		return
	}

	// Получаем все операции в заявке
	items, err := h.Repository.GetBloodlosscalcOperations(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	totalBloodLoss := 0.0
	operationResults := make(map[int]float64) // operationID -> рассчитанная кровопотеря

	// Рассчитываем кровопотерю для каждой операции
	for _, item := range items {
		operation, err := h.Repository.GetOperation(item.OperationID)
		if err != nil {
			continue
		}

		var operationLoss float64

		if item.TotalBloodLoss != nil {
			operationLoss = float64(*item.TotalBloodLoss)
		} else {
			// Если нет данных гемоглобина и времени операции, используем базовый расчет
			if item.HbBefore == nil || item.HbAfter == nil || item.SurgeryDuration == nil {
				// Базовый расчет: коэффициент * средняя кровопотеря
				operationLoss = operation.BloodLossCoeff * float64(operation.AvgBloodLoss)
			} else {
				// Расчет по полной формуле Надлера
				calculatedLoss, err := h.calculateBloodLossByNadler(
					*bloodlosscalc.PatientHeight,
					float64(*bloodlosscalc.PatientWeight),
					*item.HbBefore,
					*item.HbAfter,
					*item.SurgeryDuration,
					operation.BloodLossCoeff,
				)
				if err != nil {
					// В случае ошибки расчета используем базовый метод
					operationLoss = operation.BloodLossCoeff * float64(operation.AvgBloodLoss)
				} else {
					operationLoss = calculatedLoss
				}
			}

			// Сохраняем рассчитанную кровопотерю в БД
			calculatedLossInt := int(math.Round(operationLoss))
			err = h.Repository.UpdateBloodlosscalcOperationTotalLoss(bloodlosscalc.ID, item.OperationID, calculatedLossInt)
			if err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("failed to save calculated blood loss: %v", err))
				return
			}
		}

		operationResults[item.OperationID] = operationLoss
		totalBloodLoss += operationLoss
	}

	now := time.Now()
	err = h.Repository.UpdateBloodlosscalcStatus(id, "завершена", nil, &now)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":           "Заявка завершена",
		"total_blood_loss":  math.Round(totalBloodLoss),
		"operation_results": operationResults,
	})
}

// GET /api/operationcart - иконка корзины
func (h *Handler) GetOperationCartInfo(ctx *gin.Context) {
	userID := h.getCurrentUserID(ctx)

	currentRequest, err := h.Repository.GetCurrentBloodlosscalc(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"current_request_id": 0,
			"service_count":      0,
		})
		return
	}

	serviceCount := h.Repository.CountOperationsInBloodlosscalc(currentRequest.ID)

	ctx.JSON(http.StatusOK, gin.H{
		"current_request_id": currentRequest.ID,
		"service_count":      serviceCount,
	})
}

// DELETE /api/bloodlosscalc_operations - удаление операции из заявки
func (h *Handler) RemoveOperationFromBloodlosscalc(ctx *gin.Context) {
	bloodlosscalcIDStr := ctx.Query("bloodlosscalc_id")
	operationIDStr := ctx.Query("operation_id")

	bloodlosscalcID, err := strconv.Atoi(bloodlosscalcIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid bloodlosscalc ID"))
		return
	}

	operationID, err := strconv.Atoi(operationIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid operation ID"))
		return
	}

	err = h.Repository.RemoveOperationFromBloodlosscalc(bloodlosscalcID, operationID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Operation removed from bloodlosscalc"})
}

// PUT /api/bloodlosscalc_operations - обновление операции в заявке
func (h *Handler) UpdateBloodlosscalcOperation(ctx *gin.Context) {
	bloodlosscalcIDStr := ctx.Query("bloodlosscalc_id")
	operationIDStr := ctx.Query("operation_id")

	bloodlosscalcID, err := strconv.Atoi(bloodlosscalcIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid bloodlosscalc ID"))
		return
	}

	operationID, err := strconv.Atoi(operationIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid operation ID"))
		return
	}

	var updateData struct {
		HbBefore        *int     `json:"hb_before"`
		HbAfter         *int     `json:"hb_after"`
		SurgeryDuration *float64 `json:"surgery_duration"`
		TotalBloodLoss  *int     `json:"total_blood_loss"`
	}

	if err := ctx.BindJSON(&updateData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.UpdateBloodlosscalcOperation(bloodlosscalcID, operationID, updateData.HbBefore, updateData.HbAfter, updateData.SurgeryDuration, updateData.TotalBloodLoss)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bloodlosscalc operation updated"})
}

// CalculateTotalBloodLoss - расчет кровопотери по формуле Надлера
func (h *Handler) CalculateTotalBloodLoss(bloodlosscalcID int) (float64, error) {
	bloodlosscalc, err := h.Repository.GetBloodlosscalcByID(bloodlosscalcID)
	if err != nil {
		return 0, err
	}

	// Проверяем, что есть данные о росте и весе пациента
	if bloodlosscalc.PatientHeight == nil || bloodlosscalc.PatientWeight == nil {
		return 0, fmt.Errorf("для расчета кровопотери необходимы рост и вес пациента")
	}

	// Получаем все операции в заявке
	items, err := h.Repository.GetBloodlosscalcOperations(bloodlosscalcID)
	if err != nil {
		return 0, err
	}

	totalBloodLoss := 0.0

	for _, item := range items {
		operation, err := h.Repository.GetOperation(item.OperationID)
		if err != nil {
			continue
		}

		// Если указана фактическая кровопотеря - используем ее
		if item.TotalBloodLoss != nil {
			totalBloodLoss += float64(*item.TotalBloodLoss)
			continue
		}

		// Если нет данных гемоглобина и времени операции, используем базовый расчет
		if item.HbBefore == nil || item.HbAfter == nil || item.SurgeryDuration == nil {
			// Базовый расчет: коэффициент * средняя кровопотеря
			operationLoss := operation.BloodLossCoeff * float64(operation.AvgBloodLoss)
			totalBloodLoss += operationLoss
			continue
		}

		// Конвертируем типы для передачи в функцию расчета
		heightCm := *bloodlosscalc.PatientHeight
		weightKg := float64(*bloodlosscalc.PatientWeight)
		hbBefore := *item.HbBefore
		hbAfter := *item.HbAfter
		durationHours := *item.SurgeryDuration

		// Расчет по полной формуле Надлера
		operationLoss, err := h.calculateBloodLossByNadler(
			heightCm,
			weightKg,
			hbBefore,
			hbAfter,
			durationHours,
			operation.BloodLossCoeff,
		)
		if err != nil {
			// В случае ошибки расчета используем базовый метод
			operationLoss = operation.BloodLossCoeff * float64(operation.AvgBloodLoss)
		}

		totalBloodLoss += operationLoss
	}

	return totalBloodLoss, nil
}

// calculateBloodLossByNadler - расчет кровопотери по формуле Надлера
func (h *Handler) calculateBloodLossByNadler(heightCm, weightKg float64, hbBefore, hbAfter int, durationHours, bloodLossCoeff float64) (float64, error) {
	// Проверка валидности входных данных
	if heightCm <= 0 || weightKg <= 0 {
		return 0, fmt.Errorf("рост и вес должны быть положительными числами")
	}
	if hbBefore <= 0 || hbAfter < 0 {
		return 0, fmt.Errorf("гемоглобин должен быть положительным числом")
	}
	if hbAfter >= hbBefore {
		return 0, fmt.Errorf("гемоглобин после операции должен быть меньше чем до операции")
	}
	if durationHours < 0 {
		return 0, fmt.Errorf("длительность операции не может быть отрицательной")
	}

	// Переводим рост из см в метры
	heightM := heightCm / 100.0

	// 1. Расчет ОЦК (объема циркулирующей крови) по формуле Надлера
	// ОЦК (мл) = (0.3669 × рост³(м) + 0.03219 × вес(кг) + 0.6041) × 1000
	bv := (0.3669*math.Pow(heightM, 3) + 0.03219*weightKg + 0.6041) * 1000

	// 2. Расчет объем кровопотери по формуле:
	// Объем кровопотери (мл) = [ОЦК × (Hbдо - Hbпосле) / Hbдо] × (1 + k × T)
	hbDrop := float64(hbBefore - hbAfter)
	baseBloodLoss := bv * (hbDrop / float64(hbBefore))

	// Применяем коэффициент длительности операции
	// k - коэффициент кровопотери (используем bloodLossCoeff из операции)
	timeFactor := 1.0 + (bloodLossCoeff * durationHours)

	totalBloodLoss := baseBloodLoss * timeFactor

	// Округляем до целых миллилитров
	return math.Round(totalBloodLoss), nil
}

// DELETE /api/bloodlosscalcs/:id - удаление заявки
func (h *Handler) DeleteBloodlosscalc(ctx *gin.Context) {
	idStr := ctx.Param("id")
	bloodlosscalcID, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid bloodlosscalc ID"))
		return
	}

	// Удаляем через SQL (логическое удаление)
	err = h.Repository.DeleteBloodlosscalcSQL(bloodlosscalcID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Возвращаем JSON ответ вместо redirect
	ctx.JSON(http.StatusOK, gin.H{
		"message":          "Заявка успешно удалена",
		"bloodlosscalc_id": bloodlosscalcID,
	})
}

// POST /api/register - регистрация пользователя
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var requestData struct {
		Username string `json:"username" binding:"required,min=3,max=32"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := ctx.BindJSON(&requestData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем, не существует ли уже пользователь с таким username
	existingUser, err := h.Repository.GetUserByUsername(requestData.Username)
	if err == nil && existingUser.UserID != 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("пользователь с таким именем уже существует"))
		return
	}

	// Сохраняем пароль в открытом виде (без хеширования)
	newUser := &ds.User{
		Username:    requestData.Username,
		Password:    requestData.Password,
		IsModerator: false,
	}

	err = h.Repository.CreateUser(newUser)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно зарегистрирован",
		"user": gin.H{
			"user_id":      newUser.UserID,
			"username":     newUser.Username,
			"is_moderator": newUser.IsModerator,
		},
	})
}

// GET /api/user - получить профиль пользователя
func (h *Handler) GetUserProfile(ctx *gin.Context) {
	userID := h.getCurrentUserID(ctx)

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("пользователь не найден"))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"user_id":      user.UserID,
			"username":     user.Username,
			"is_moderator": user.IsModerator,
		},
	})
}

// PUT /api/user - обновить профиль пользователя
func (h *Handler) UpdateUserProfile(ctx *gin.Context) {
	userID := h.getCurrentUserID(ctx)

	var requestData struct {
		Username *string `json:"username"`
		Password *string `json:"password"`
	}

	if err := ctx.BindJSON(&requestData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	currentUser, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("пользователь не найден"))
		return
	}

	updates := make(map[string]interface{})

	if requestData.Username != nil {
		if *requestData.Username != currentUser.Username {
			existingUser, err := h.Repository.GetUserByUsername(*requestData.Username)
			if err == nil && existingUser.UserID != 0 {
				h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("пользователь с таким именем уже существует"))
				return
			}
		}
		updates["username"] = *requestData.Username
	}

	if requestData.Password != nil {
		updates["password"] = *requestData.Password
	}

	if len(updates) == 0 {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("нет полей для обновления"))
		return
	}

	err = h.Repository.UpdateUser(userID, updates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Профиль успешно обновлен",
	})
}

// POST /api/auth - аутентификация (заглушка)
func (h *Handler) AuthenticateUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Заглушка",
	})
}

// POST /api/logout - деавторизация (заглушка)
func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Заглушка",
	})
}

func (h *Handler) getCurrentUserID(ctx *gin.Context) int {
	return 2
}

// Вспомогательные функции для работы с изображениями
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func generateFileName(ext string) string {
	uuid := uuid.New().String()
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("operation_%s_%s%s", timestamp, uuid, ext)
}

func getFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
