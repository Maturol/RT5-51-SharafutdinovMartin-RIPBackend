package repository

import (
	"blood_loss_calc/internal/app/ds"
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetOperation(id int) (ds.Operation, error) {
	var operation ds.Operation
	err := r.db.Where("operation_id = ?", id).First(&operation).Error
	if err != nil {
		return ds.Operation{}, fmt.Errorf("услуга не найдена")
	}
	return operation, nil
}

func (r *Repository) GetOperations() ([]ds.Operation, error) {
	var operations []ds.Operation
	err := r.db.Find(&operations).Error
	if err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *Repository) GetOperationsByTitle(title string) ([]ds.Operation, error) {
	var operations []ds.Operation
	err := r.db.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(title)+"%").Find(&operations).Error
	if err != nil {
		return nil, err
	}
	return operations, nil
}

func (r *Repository) GetBloodlosscalc() (ds.Bloodlosscalc, error) {
	var bloodlosscalc ds.Bloodlosscalc
	err := r.db.First(&bloodlosscalc).Error
	if err != nil {
		return ds.Bloodlosscalc{}, err
	}
	return bloodlosscalc, nil
}

// GetBloodlosscalcByID возвращает заявку по ID
func (r *Repository) GetBloodlosscalcByID(id int) (ds.Bloodlosscalc, error) {
	var bloodlosscalc ds.Bloodlosscalc
	err := r.db.
		Preload("Creator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		}).
		Preload("Moderator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		}).
		Where("bloodlosscalc_id = ?", id).
		First(&bloodlosscalc).Error
	if err != nil {
		return ds.Bloodlosscalc{}, err
	}
	return bloodlosscalc, nil
}

// GetBloodlosscalcOperations возвращает операции заявки
func (r *Repository) GetBloodlosscalcOperations(bloodlosscalcID int) ([]ds.BloodlosscalcOperation, error) {
	var items []ds.BloodlosscalcOperation
	err := r.db.
		Preload("Operation").
		Where("bloodlosscalc_id = ?", bloodlosscalcID).
		Find(&items).Error
	return items, err
}

// Получение текущей заявки пользователя (статус "черновик")
func (r *Repository) GetCurrentBloodlosscalc(userID int) (ds.Bloodlosscalc, error) {
	var bloodlosscalc ds.Bloodlosscalc
	err := r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&bloodlosscalc).Error
	if err != nil {
		return ds.Bloodlosscalc{}, err
	}
	return bloodlosscalc, nil
}

// Создание новой заявки для пользователя
func (r *Repository) CreateBloodlosscalc(userID int, height *float64, weight *int) (ds.Bloodlosscalc, error) {
	bloodlosscalc := ds.Bloodlosscalc{
		Status:        "черновик",
		CreatorID:     userID,
		PatientHeight: height,
		PatientWeight: weight,
	}

	err := r.db.Create(&bloodlosscalc).Error
	if err != nil {
		return ds.Bloodlosscalc{}, err
	}
	return bloodlosscalc, nil
}

// Добавление услуги (операции) в заявку
func (r *Repository) AddOperationToBloodlosscalc(bloodlosscalcID, operationID int, hbBefore, hbAfter *int, duration *float64, totalLoss *int) error {
	item := ds.BloodlosscalcOperation{
		BloodlosscalcID: bloodlosscalcID,
		OperationID:     operationID,
		HbBefore:        hbBefore,
		HbAfter:         hbAfter,
		SurgeryDuration: duration,
		TotalBloodLoss:  totalLoss,
	}

	return r.db.Create(&item).Error
}

// Проверка существования операции в заявке
func (r *Repository) OperationExistsInBloodlosscalc(bloodlosscalcID, operationID int) bool {
	var count int64
	r.db.Model(&ds.BloodlosscalcOperation{}).
		Where("bloodlosscalc_id = ? AND operation_id = ?", bloodlosscalcID, operationID).
		Count(&count)
	return count > 0
}

// Подсчет количества операций в заявке
func (r *Repository) CountOperationsInBloodlosscalc(bloodlosscalcID int) int64 {
	var count int64
	r.db.Model(&ds.BloodlosscalcOperation{}).
		Where("bloodlosscalc_id = ?", bloodlosscalcID).
		Count(&count)
	return count
}

// ЛОГИЧЕСКОЕ УДАЛЕНИЕ ЗАЯВКИ ЧЕРЕЗ ПРЯМОЙ SQL
func (r *Repository) DeleteBloodlosscalcSQL(bloodlosscalcID int) error {
	result := r.db.Exec("UPDATE bloodlosscalcs SET status = 'удален' WHERE bloodlosscalc_id = ?", bloodlosscalcID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("заявка не найдена")
	}
	return nil
}

// UpdateOperationImage обновляет изображение операции
func (r *Repository) UpdateOperationImage(operationID int, imageURL *string) error {
	return r.db.Model(&ds.Operation{}).
		Where("operation_id = ?", operationID).
		Update("image_url", imageURL).Error
}

// CreateOperation создает новую операцию
func (r *Repository) CreateOperation(operation *ds.Operation) error {
	return r.db.Create(operation).Error
}

// UpdateOperation обновляет операцию
func (r *Repository) UpdateOperation(operation *ds.Operation) error {
	return r.db.Save(operation).Error
}

// DeleteOperation удаляет операцию и её изображение
func (r *Repository) DeleteOperation(id int) error {
	// Сначала получаем операцию для удаления изображения
	operation, err := r.GetOperation(id)
	if err != nil {
		return err
	}

	// Удаляем изображение из Minio если есть
	if operation.ImageURL != nil && *operation.ImageURL != "" {
		fileName := getFileNameFromURL(*operation.ImageURL)
		if fileName != "" {
			go func() {
				ctx := context.Background()
				r.DeleteFile(ctx, fileName)
			}()
		}
	}

	return r.db.Delete(&ds.Operation{}, id).Error
}

// RemoveOperationFromBloodlosscalc удаляет операцию из заявки
func (r *Repository) RemoveOperationFromBloodlosscalc(bloodlosscalcID, operationID int) error {
	return r.db.Where("bloodlosscalc_id = ? AND operation_id = ?", bloodlosscalcID, operationID).
		Delete(&ds.BloodlosscalcOperation{}).Error
}

// UpdateBloodlosscalcOperation обновляет данные операции в заявке
func (r *Repository) UpdateBloodlosscalcOperation(bloodlosscalcID, operationID int, hbBefore, hbAfter *int, duration *float64, totalLoss *int) error {
	return r.db.Model(&ds.BloodlosscalcOperation{}).
		Where("bloodlosscalc_id = ? AND operation_id = ?", bloodlosscalcID, operationID).
		Updates(map[string]interface{}{
			"hb_before":        hbBefore,
			"hb_after":         hbAfter,
			"surgery_duration": duration,
			"total_blood_loss": totalLoss,
		}).Error
}

// GetBloodlosscalcsFiltered возвращает отфильтрованные заявки (исключая черновики и удаленные)
func (r *Repository) GetBloodlosscalcsFiltered(status string, dateFrom, dateTo *time.Time) ([]ds.Bloodlosscalc, error) {
	var bloodlosscalcs []ds.Bloodlosscalc

	query := r.db.
		Preload("Creator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		}).
		Preload("Moderator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		})

	// Исключаем черновики и удаленные заявки
	query = query.Where("status NOT IN (?, ?)", "черновик", "удален")

	// Дополнительная фильтрация по статусу если указана
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if dateFrom != nil {
		query = query.Where("formed_at >= ?", dateFrom)
	}

	if dateTo != nil {
		query = query.Where("formed_at <= ?", dateTo)
	}

	query = query.Order("created_at DESC")

	err := query.Find(&bloodlosscalcs).Error
	return bloodlosscalcs, err
}

// UpdateBloodlosscalc обновляет данные заявки
func (r *Repository) UpdateBloodlosscalc(id int, updates map[string]interface{}) error {
	return r.db.Model(&ds.Bloodlosscalc{}).
		Where("bloodlosscalc_id = ?", id).
		Updates(updates).Error
}

// UpdateBloodlosscalcStatus обновляет статус заявки
func (r *Repository) UpdateBloodlosscalcStatus(id int, status string, formedAt, completedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if formedAt != nil {
		updates["formed_at"] = formedAt
	}

	if completedAt != nil {
		updates["completed_at"] = completedAt
		updates["moderator_id"] = 1 // временная заглушка
	}

	return r.db.Model(&ds.Bloodlosscalc{}).
		Where("bloodlosscalc_id = ?", id).
		Updates(updates).Error
}

// Вспомогательная функция для извлечения имени файла из URL
func getFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// PartialUpdateOperation частично обновляет операцию
func (r *Repository) PartialUpdateOperation(id int, updates map[string]interface{}) error {
	return r.db.Model(&ds.Operation{}).
		Where("operation_id = ?", id).
		Updates(updates).Error
}

// UpdateBloodlosscalcOperationTotalLoss обновляет поле TotalBloodLoss для операции в заявке
func (r *Repository) UpdateBloodlosscalcOperationTotalLoss(bloodlosscalcID, operationID int, totalBloodLoss int) error {
	return r.db.Model(&ds.BloodlosscalcOperation{}).
		Where("bloodlosscalc_id = ? AND operation_id = ?", bloodlosscalcID, operationID).
		Update("total_blood_loss", totalBloodLoss).Error
}

// CreateUser создает нового пользователя
func (r *Repository) CreateUser(user *ds.User) error {
	return r.db.Create(user).Error
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(userID int) (ds.User, error) {
	var user ds.User
	err := r.db.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return ds.User{}, err
	}
	return user, nil
}

// GetUserByUsername возвращает пользователя по username
func (r *Repository) GetUserByUsername(username string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return ds.User{}, err
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя
func (r *Repository) UpdateUser(userID int, updates map[string]interface{}) error {
	return r.db.Model(&ds.User{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// GetUserBloodlosscalcs возвращает заявки пользователя с фильтрацией
func (r *Repository) GetUserBloodlosscalcs(userID int, status string, dateFrom, dateTo *time.Time) ([]ds.Bloodlosscalc, error) {
	var bloodlosscalcs []ds.Bloodlosscalc

	query := r.db.
		Preload("Creator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		}).
		Preload("Moderator", func(db *gorm.DB) *gorm.DB {
			return db.Select("user_id, username")
		}).
		Where("creator_id = ? AND status != ?", userID, "удален")

	// Дополнительная фильтрация по статусу если указана
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if dateFrom != nil {
		query = query.Where("created_at >= ?", dateFrom)
	}

	if dateTo != nil {
		query = query.Where("created_at <= ?", dateTo)
	}

	query = query.Order("created_at DESC")

	err := query.Find(&bloodlosscalcs).Error
	return bloodlosscalcs, err
}
