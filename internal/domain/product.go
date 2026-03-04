package domain

import (
	"strings"
	"time"
)

var ValidUnits = []string{"Box", "Strip", "Blister", "Sheet", "Pcs"}

var ValidCategories = []string{
	"Medicine - OTC",
	"Medicine - Prescription",
	"Supplements",
	"Medical Supplies",
	"Maternal & Baby",
}

var ValidTherapeuticClasses = []string{
	"Analgesics",
	"Antibiotics",
	"Antihypertensives",
	"Antidiabetics",
	"Antihistamines",
	"Vitamins",
	"Antacids",
	"Antipyretics",
	"Antiseptics",
	"Bronchodilators",
}

func IsValidCategory(cat string) bool {
	for _, v := range ValidCategories {
		if strings.EqualFold(v, cat) {
			return true
		}
	}
	return false
}

func IsValidUnit(unit string) bool {
	for _, v := range ValidUnits {
		if strings.EqualFold(v, unit) {
			return true
		}
	}
	return false
}

func IsValidTherapeuticClass(class string) bool {
	if class == "" {
		return true // Optional
	}
	for _, v := range ValidTherapeuticClasses {
		if strings.EqualFold(v, class) {
			return true
		}
	}
	return false
}

type Product struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	SKUCode          string    `json:"sku_code"`
	Category         string    `json:"category"` // Kept for backward compatibility
	TherapeuticClass string    `json:"therapeutic_class"`
	Unit             string    `json:"unit"`
	BaseUnit         string    `json:"base_unit"`
	ItemsPerUnit     int       `json:"items_per_unit"`
	StorageLocation  string    `json:"storage_location"`
	PurchasePrice    float64   `json:"purchase_price"`
	SellingPrice     float64   `json:"selling_price"`
	MinStock         int       `json:"min_stock"`
	Status           string    `json:"status"` // active, inactive
	IsVerified       bool      `json:"is_verified"`
	CreatedAt        time.Time `json:"created_at"`
}
