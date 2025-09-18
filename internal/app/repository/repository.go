package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
	services []Service
	cart     Cart
}

func NewRepository() (*Repository, error) {
	repo := &Repository{}
	repo.initializeTestData()
	return repo, nil
}

func (r *Repository) initializeTestData() {
	r.services = []Service{
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
			ImageURL:       "hip_replacement.png",
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

	r.cart = Cart{
		ID:     1,
		Height: 1.75,
		Weight: 70,
		Items: []CartItem{
			{
				ServiceID:       1,
				HbBefore:        140,
				HbAfter:         120,
				SurgeryDuration: 2,
				BloodLossResult: 350,
			},
			{
				ServiceID:       3,
				HbBefore:        145,
				HbAfter:         130,
				SurgeryDuration: 3,
				BloodLossResult: 420,
			},
		},
		TotalBloodLoss: 770,
	}
}

type Service struct {
	ID             int
	Title          string
	ImageURL       string
	BloodLossCoeff float64
	AvgBloodLoss   float64
	Description    string
}

type CartItem struct {
	ServiceID       int
	HbBefore        float64
	HbAfter         float64
	SurgeryDuration float64
	BloodLossResult float64
}

type Cart struct {
	ID             int
	Height         float64
	Weight         float64
	Items          []CartItem
	TotalBloodLoss float64
}

// Методы для работы с услугами
func (r *Repository) GetService(id int) (Service, error) {
	for _, service := range r.services {
		if service.ID == id {
			return service, nil
		}
	}
	return Service{}, fmt.Errorf("услуга не найдена")
}

func (r *Repository) GetServices() ([]Service, error) {
	return r.services, nil
}

func (r *Repository) GetServicesByTitle(title string) ([]Service, error) {
	var result []Service
	for _, service := range r.services {
		if strings.Contains(strings.ToLower(service.Title), strings.ToLower(title)) {
			result = append(result, service)
		}
	}
	return result, nil
}

func (r *Repository) GetCart() Cart {
	return r.cart
}
