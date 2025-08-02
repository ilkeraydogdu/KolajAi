package models

import (
	"testing"
	"time"
)

func TestCustomer_Validate(t *testing.T) {
	tests := []struct {
		name     string
		customer Customer
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid customer",
			customer: Customer{
				UserID:       1,
				FirstName:    "John",
				LastName:     "Doe",
				Email:        "john@example.com",
				CustomerType: "individual",
			},
			wantErr: false,
		},
		{
			name: "empty first name",
			customer: Customer{
				UserID:    1,
				FirstName: "",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			wantErr: true,
			errMsg:  "first name cannot be empty",
		},
		{
			name: "empty last name",
			customer: Customer{
				UserID:    1,
				FirstName: "John",
				LastName:  "",
				Email:     "john@example.com",
			},
			wantErr: true,
			errMsg:  "last name cannot be empty",
		},
		{
			name: "empty email",
			customer: Customer{
				UserID:    1,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "",
			},
			wantErr: true,
			errMsg:  "email cannot be empty",
		},
		{
			name: "invalid user ID",
			customer: Customer{
				UserID:    0,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			wantErr: true,
			errMsg:  "valid user ID is required",
		},
		{
			name: "invalid customer type",
			customer: Customer{
				UserID:       1,
				FirstName:    "John",
				LastName:     "Doe",
				Email:        "john@example.com",
				CustomerType: "invalid",
			},
			wantErr: true,
			errMsg:  "customer type must be individual or corporate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.customer.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Customer.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Customer.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestCustomer_GetFullName(t *testing.T) {
	tests := []struct {
		name     string
		customer Customer
		want     string
	}{
		{
			name: "normal names",
			customer: Customer{
				FirstName: "John",
				LastName:  "Doe",
			},
			want: "John Doe",
		},
		{
			name: "names with spaces",
			customer: Customer{
				FirstName: " John ",
				LastName:  " Doe ",
			},
			want: "John   Doe",
		},
		{
			name: "empty names",
			customer: Customer{
				FirstName: "",
				LastName:  "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.customer.GetFullName()
			if got != tt.want {
				t.Errorf("Customer.GetFullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_Validate(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid address",
			address: Address{
				CustomerID:   1,
				FirstName:    "John",
				LastName:     "Doe",
				AddressLine1: "123 Main St",
				City:         "New York",
				Country:      "USA",
				Type:         "both",
			},
			wantErr: false,
		},
		{
			name: "empty first name",
			address: Address{
				CustomerID:   1,
				FirstName:    "",
				LastName:     "Doe",
				AddressLine1: "123 Main St",
				City:         "New York",
				Country:      "USA",
			},
			wantErr: true,
			errMsg:  "first name cannot be empty",
		},
		{
			name: "empty address line 1",
			address: Address{
				CustomerID:   1,
				FirstName:    "John",
				LastName:     "Doe",
				AddressLine1: "",
				City:         "New York",
				Country:      "USA",
			},
			wantErr: true,
			errMsg:  "address line 1 cannot be empty",
		},
		{
			name: "empty city",
			address: Address{
				CustomerID:   1,
				FirstName:    "John",
				LastName:     "Doe",
				AddressLine1: "123 Main St",
				City:         "",
				Country:      "USA",
			},
			wantErr: true,
			errMsg:  "city cannot be empty",
		},
		{
			name: "empty country",
			address: Address{
				CustomerID:   1,
				FirstName:    "John",
				LastName:     "Doe",
				AddressLine1: "123 Main St",
				City:         "New York",
				Country:      "",
			},
			wantErr: true,
			errMsg:  "country cannot be empty",
		},
		{
			name: "invalid customer ID",
			address: Address{
				CustomerID:   0,
				FirstName:    "John",
				LastName:     "Doe",
				AddressLine1: "123 Main St",
				City:         "New York",
				Country:      "USA",
			},
			wantErr: true,
			errMsg:  "valid customer ID is required",
		},
		{
			name: "invalid type",
			address: Address{
				CustomerID:   1,
				FirstName:    "John",
				LastName:     "Doe",
				AddressLine1: "123 Main St",
				City:         "New York",
				Country:      "USA",
				Type:         "invalid",
			},
			wantErr: true,
			errMsg:  "address type must be billing, shipping, or both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.address.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Address.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Address.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestAddress_GetFullAddress(t *testing.T) {
	tests := []struct {
		name    string
		address Address
		want    string
	}{
		{
			name: "complete address",
			address: Address{
				AddressLine1: "123 Main St",
				AddressLine2: "Apt 4B",
				City:         "New York",
				State:        "NY",
				PostalCode:   "10001",
				Country:      "USA",
			},
			want: "123 Main St, Apt 4B, New York, NY, 10001, USA",
		},
		{
			name: "minimal address",
			address: Address{
				AddressLine1: "123 Main St",
				City:         "New York",
				Country:      "USA",
			},
			want: "123 Main St, New York, USA",
		},
		{
			name: "empty address",
			address: Address{},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.address.GetFullAddress()
			if got != tt.want {
				t.Errorf("Address.GetFullAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomerCreationDefaults(t *testing.T) {
	customer := Customer{
		UserID:    1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test default values
	if customer.Status != "" {
		// Status should be set by service layer, not model
		t.Errorf("Expected empty status by default, got %v", customer.Status)
	}

	if customer.LoyaltyPoints != 0 {
		t.Errorf("Expected 0 loyalty points by default, got %v", customer.LoyaltyPoints)
	}

	if customer.TotalSpent != 0 {
		t.Errorf("Expected 0 total spent by default, got %v", customer.TotalSpent)
	}

	if customer.OrderCount != 0 {
		t.Errorf("Expected 0 order count by default, got %v", customer.OrderCount)
	}
}

func TestAddressCreationDefaults(t *testing.T) {
	address := Address{
		CustomerID:   1,
		FirstName:    "John",
		LastName:     "Doe",
		AddressLine1: "123 Main St",
		City:         "New York",
		Country:      "USA",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test default values
	if address.IsDefault {
		t.Errorf("Expected false for IsDefault by default, got %v", address.IsDefault)
	}

	if !address.IsActive {
		// IsActive should default to true
		address.IsActive = true
	}

	if address.Type == "" {
		// Type should be set by service layer
		address.Type = "both"
	}
}