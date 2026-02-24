package domain

import (
	"strings"
	"time"
)

var ValidUnits = []string{"Box", "Strip", "Pcs", "Vial", "Botol", "Tube", "Sachet", "Ampul", "Pot", "Dus"}

var ValidLegalCategories = []string{"OTC", "Rx"}

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

func IsValidUnit(unit string) bool {
	for _, v := range ValidUnits {
		if strings.EqualFold(v, unit) {
			return true
		}
	}
	return false
}

func IsValidLegalCategory(cat string) bool {
	for _, v := range ValidLegalCategories {
		if strings.EqualFold(v, cat) {
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
	LegalCategory    string    `json:"legal_category"`
	TherapeuticClass string    `json:"therapeutic_class"`
	Unit             string    `json:"unit"`
	ItemsPerUnit     int       `json:"items_per_unit"`
	StorageLocation  string    `json:"storage_location"`
	PurchasePrice    float64   `json:"purchase_price"`
	SellingPrice     float64   `json:"selling_price"`
	MinStock         int       `json:"min_stock"`
	Status           string    `json:"status"` // active, inactive
	IsVerified       bool      `json:"is_verified"`
	CreatedAt        time.Time `json:"created_at"`
}
