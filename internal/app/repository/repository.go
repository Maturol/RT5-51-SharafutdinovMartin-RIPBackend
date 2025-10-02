package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
	operations    []Operation
	bloodlosscalc Bloodlosscalc
}

func NewRepository() (*Repository, error) {
	repo := &Repository{}
	repo.initializeTestData()
	return repo, nil
}

func (r *Repository) initializeTestData() {
	r.operations = []Operation{
		{
			ID:             1,
			Title:          "Кесарево сечение",
			ImageURL:       "cesarean.jpg",
			BloodLossCoeff: 0.12,
			AvgBloodLoss:   500,
			Description:    "Операция родоразрешения путем извлечения плода через разрез на матке",
		},
		{
			ID:             2,
			Title:          "Эндопротезирование тазобедренного сустава",
			ImageURL:       "hip_replacement.jpg",
			BloodLossCoeff: 0.22,
			AvgBloodLoss:   800,
			Description:    "Замена поврежденных частей тазобедренного сустава на искусственные имплантаты",
		},
		{
			ID:             3,
			Title:          "Спондилодез",
			ImageURL:       "spondylodesis.jpg",
			BloodLossCoeff: 0.19,
			AvgBloodLoss:   600,
			Description:    "Операция по сращению позвонков для стабилизации позвоночника",
		},
		{
			ID:             4,
			Title:          "Аппендэктомия",
			ImageURL:       "appendectomy.jpg",
			BloodLossCoeff: 0.04,
			AvgBloodLoss:   150,
			Description:    "Удаление червеобразного отростка слепой кишки",
		},
	}

	r.bloodlosscalc = Bloodlosscalc{
		ID:     1,
		Height: 1.75,
		Weight: 70,
		Items: []BloodlosscalcItem{
			{
				OperationID:     1,
				HbBefore:        140,
				HbAfter:         120,
				SurgeryDuration: 2,
				BloodLossResult: 350,
			},
			{
				OperationID:     3,
				HbBefore:        145,
				HbAfter:         130,
				SurgeryDuration: 3,
				BloodLossResult: 420,
			},
		},
		TotalBloodLoss: 770,
	}
}

type Operation struct {
	ID             int
	Title          string
	ImageURL       string
	BloodLossCoeff float64
	AvgBloodLoss   float64
	Description    string
}

type BloodlosscalcItem struct {
	OperationID     int
	HbBefore        float64
	HbAfter         float64
	SurgeryDuration float64
	BloodLossResult float64
}

type Bloodlosscalc struct {
	ID             int
	Height         float64
	Weight         float64
	Items          []BloodlosscalcItem
	TotalBloodLoss float64
}

// Методы для работы с услугами
func (r *Repository) GetOperation(id int) (Operation, error) {
	for _, operation := range r.operations {
		if operation.ID == id {
			return operation, nil
		}
	}
	return Operation{}, fmt.Errorf("услуга не найдена")
}

func (r *Repository) GetOperations() ([]Operation, error) {
	return r.operations, nil
}

func (r *Repository) GetOperationsByTitle(title string) ([]Operation, error) {
	var result []Operation
	for _, operation := range r.operations {
		if strings.Contains(strings.ToLower(operation.Title), strings.ToLower(title)) {
			result = append(result, operation)
		}
	}
	return result, nil
}

func (r *Repository) GetBloodlosscalc() Bloodlosscalc {
	return r.bloodlosscalc
}

func (r *Repository) GetBloodlosscalcByID(id int) Bloodlosscalc {
	if r.bloodlosscalc.ID == id {
		return r.bloodlosscalc
	}
	return Bloodlosscalc{}
}
