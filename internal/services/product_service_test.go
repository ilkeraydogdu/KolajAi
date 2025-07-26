package services

import (
	"kolajAi/internal/models"
	"testing"
	"time"
)

// MockRepository implements SimpleRepository for testing
type MockRepository struct {
	products []models.Product
}

func (m *MockRepository) CreateStruct(table string, data interface{}) (int64, error) {
	return 1, nil
}

func (m *MockRepository) Update(table string, id interface{}, data interface{}) error {
	return nil
}

func (m *MockRepository) Delete(table string, id interface{}) error {
	return nil
}

func (m *MockRepository) FindByID(table string, id interface{}, result interface{}) error {
	return nil
}

func (m *MockRepository) FindAll(table string, result interface{}, conditions map[string]interface{}, orderBy string, limit, offset int) error {
	return nil
}

func (m *MockRepository) FindOne(table string, result interface{}, conditions map[string]interface{}) error {
	return nil
}

func (m *MockRepository) Count(table string, conditions map[string]interface{}) (int64, error) {
	return 0, nil
}

func (m *MockRepository) Search(table string, fields []string, term string, limit, offset int, result interface{}) error {
	return nil
}

func (m *MockRepository) FindByDateRange(table, dateField string, start, end time.Time, limit, offset int, result interface{}) error {
	return nil
}

func (m *MockRepository) SetConnectionPool(maxOpen, maxIdle int, maxLifetime time.Duration) {
}

func (m *MockRepository) SoftDelete(table string, id interface{}) error {
	return nil
}

func (m *MockRepository) BulkCreate(table string, data []interface{}) ([]int64, error) {
	return []int64{}, nil
}

func (m *MockRepository) BulkUpdate(table string, ids []interface{}, data interface{}) error {
	return nil
}

func (m *MockRepository) BulkDelete(table string, ids []interface{}) error {
	return nil
}

func (m *MockRepository) Exists(table string, conditions map[string]interface{}) (bool, error) {
	return false, nil
}

func TestNewProductService(t *testing.T) {
	mockRepo := &MockRepository{}

	service := NewProductService(mockRepo)
	if service == nil {
		t.Error("Expected product service to be created, but got nil")
	}
}

func TestProductService_ValidateProduct(t *testing.T) {
	mockRepo := &MockRepository{}
	_ = NewProductService(mockRepo)

	tests := []struct {
		name    string
		product models.Product
		wantErr bool
	}{
		{
			name: "valid product",
			product: models.Product{
				Name:       "Test Product",
				Price:      99.99,
				Stock:      10,
				VendorID:   1,
				CategoryID: 1,
				Status:     "active",
			},
			wantErr: false,
		},
		{
			name: "invalid product - empty name",
			product: models.Product{
				Name:       "",
				Price:      99.99,
				Stock:      10,
				VendorID:   1,
				CategoryID: 1,
				Status:     "active",
			},
			wantErr: true,
		},
		{
			name: "invalid product - negative price",
			product: models.Product{
				Name:       "Test Product",
				Price:      -10.0,
				Stock:      10,
				VendorID:   1,
				CategoryID: 1,
				Status:     "active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Product validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
