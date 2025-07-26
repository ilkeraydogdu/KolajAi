package models

import (
	"testing"
	"time"
)

func TestProduct_Validate(t *testing.T) {
	tests := []struct {
		name    string
		product Product
		wantErr bool
	}{
		{
			name: "valid product",
			product: Product{
				Name:        "Test Product",
				Description: "A test product description",
				Price:       99.99,
				Stock:       10,
				VendorID:    1,
				CategoryID:  1,
				Status:      "active",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			product: Product{
				Name:        "",
				Description: "A test product description",
				Price:       99.99,
				Stock:       10,
				VendorID:    1,
				CategoryID:  1,
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "negative price",
			product: Product{
				Name:        "Test Product",
				Description: "A test product description",
				Price:       -10.00,
				Stock:       10,
				VendorID:    1,
				CategoryID:  1,
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "negative stock",
			product: Product{
				Name:        "Test Product",
				Description: "A test product description",
				Price:       99.99,
				Stock:       -5,
				VendorID:    1,
				CategoryID:  1,
				Status:      "active",
			},
			wantErr: true,
		},
		{
			name: "invalid vendor ID",
			product: Product{
				Name:        "Test Product",
				Description: "A test product description",
				Price:       99.99,
				Stock:       10,
				VendorID:    0,
				CategoryID:  1,
				Status:      "active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Product.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProduct_IsAvailable(t *testing.T) {
	availableProduct := Product{
		Name:        "Available Product",
		Description: "An available product",
		Price:       99.99,
		Stock:       10,
		VendorID:    1,
		CategoryID:  1,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	unavailableProduct := Product{
		Name:        "Unavailable Product",
		Description: "An unavailable product",
		Price:       99.99,
		Stock:       0,
		VendorID:    1,
		CategoryID:  1,
		Status:      "active",
		CreatedAt:   time.Now(),
	}

	inactiveProduct := Product{
		Name:        "Inactive Product",
		Description: "An inactive product",
		Price:       99.99,
		Stock:       10,
		VendorID:    1,
		CategoryID:  1,
		Status:      "inactive",
		CreatedAt:   time.Now(),
	}

	if availableProduct.Status != "active" || availableProduct.Stock <= 0 {
		t.Error("Expected available product to be active and have stock")
	}

	if unavailableProduct.Stock > 0 {
		t.Error("Expected unavailable product to have no stock")
	}

	if inactiveProduct.Status == "active" {
		t.Error("Expected inactive product to be inactive")
	}
}

func TestProduct_ComparePrice(t *testing.T) {
	product := Product{
		Name:         "Test Product",
		Price:        80.00,
		ComparePrice: 100.00,
	}

	if product.ComparePrice <= product.Price {
		t.Error("Expected compare price to be higher than regular price")
	}

	savings := product.ComparePrice - product.Price
	expectedSavings := 20.00

	if savings != expectedSavings {
		t.Errorf("Expected savings %v, got %v", expectedSavings, savings)
	}
}
