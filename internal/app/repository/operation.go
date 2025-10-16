package repository

import (
	"fmt"
	"strings"

	"blood_loss_calc/internal/app/ds"
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

func (r *Repository) GetBloodlosscalcByID(id int) (ds.Bloodlosscalc, error) {
	var bloodlosscalc ds.Bloodlosscalc
	err := r.db.Where("bloodlosscalc_id = ?", id).First(&bloodlosscalc).Error
	if err != nil {
		return ds.Bloodlosscalc{}, err
	}
	return bloodlosscalc, nil
}

func (r Repository) GetBloodlosscalcOperations(bloodlosscalcID int) ([]ds.BloodlosscalcOperation, error) {
	var items []ds.BloodlosscalcOperation
	err := r.db.Where("bloodlosscalc_id = ?", bloodlosscalcID).Find(&items).Error
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
