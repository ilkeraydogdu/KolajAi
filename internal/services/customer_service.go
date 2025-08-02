package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"time"
)

type CustomerService struct {
	repo database.SimpleRepository
}

func NewCustomerService(repo database.SimpleRepository) *CustomerService {
	return &CustomerService{repo: repo}
}

// CreateCustomer creates a new customer
func (s *CustomerService) CreateCustomer(customer *models.Customer) error {
	if err := customer.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()
	
	if customer.Status == "" {
		customer.Status = "active"
	}

	id, err := s.repo.CreateStruct("customers", customer)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}
	customer.ID = id
	return nil
}

// GetCustomerByID retrieves a customer by ID
func (s *CustomerService) GetCustomerByID(id int64) (*models.Customer, error) {
	var customer models.Customer
	err := s.repo.FindByID("customers", id, &customer)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	return &customer, nil
}

// GetCustomerByUserID retrieves a customer by user ID
func (s *CustomerService) GetCustomerByUserID(userID int64) (*models.Customer, error) {
	var customer models.Customer
	conditions := map[string]interface{}{"user_id": userID}
	err := s.repo.FindOne("customers", &customer, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer by user ID: %w", err)
	}
	return &customer, nil
}

// GetCustomerByEmail retrieves a customer by email
func (s *CustomerService) GetCustomerByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	conditions := map[string]interface{}{"email": email}
	err := s.repo.FindOne("customers", &customer, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}
	return &customer, nil
}

// UpdateCustomer updates a customer
func (s *CustomerService) UpdateCustomer(id int64, customer *models.Customer) error {
	if err := customer.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	customer.UpdatedAt = time.Now()
	err := s.repo.Update("customers", id, customer)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}
	return nil
}

// DeleteCustomer soft deletes a customer
func (s *CustomerService) DeleteCustomer(id int64) error {
	err := s.repo.SoftDelete("customers", id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	return nil
}

// GetCustomers retrieves customers with pagination
func (s *CustomerService) GetCustomers(limit, offset int) ([]models.Customer, error) {
	var customers []models.Customer
	conditions := map[string]interface{}{}
	err := s.repo.FindAll("customers", &customers, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}
	return customers, nil
}

// SearchCustomers searches for customers
func (s *CustomerService) SearchCustomers(term string, limit, offset int) ([]models.Customer, error) {
	var customers []models.Customer
	fields := []string{"first_name", "last_name", "email", "phone", "company_name"}
	err := s.repo.Search("customers", fields, term, limit, offset, &customers)
	if err != nil {
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}
	return customers, nil
}

// UpdateLoyaltyPoints updates customer loyalty points
func (s *CustomerService) UpdateLoyaltyPoints(customerID int64, points int) error {
	customer, err := s.GetCustomerByID(customerID)
	if err != nil {
		return err
	}
	
	customer.LoyaltyPoints += points
	customer.UpdatedAt = time.Now()
	
	return s.repo.Update("customers", customerID, customer)
}

// UpdateTotalSpent updates customer total spent amount
func (s *CustomerService) UpdateTotalSpent(customerID int64, amount float64) error {
	customer, err := s.GetCustomerByID(customerID)
	if err != nil {
		return err
	}
	
	customer.TotalSpent += amount
	customer.OrderCount++
	customer.LastOrderDate = &time.Time{}
	*customer.LastOrderDate = time.Now()
	customer.UpdatedAt = time.Now()
	
	return s.repo.Update("customers", customerID, customer)
}

// CreateAddress creates a new address for a customer
func (s *CustomerService) CreateAddress(address *models.Address) error {
	if err := address.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	address.CreatedAt = time.Now()
	address.UpdatedAt = time.Now()
	
	if address.IsDefault {
		// Set other addresses as non-default
		if err := s.unsetDefaultAddresses(address.CustomerID); err != nil {
			return fmt.Errorf("failed to unset default addresses: %w", err)
		}
	}

	id, err := s.repo.CreateStruct("addresses", address)
	if err != nil {
		return fmt.Errorf("failed to create address: %w", err)
	}
	address.ID = id
	return nil
}

// GetAddressByID retrieves an address by ID
func (s *CustomerService) GetAddressByID(id int64) (*models.Address, error) {
	var address models.Address
	err := s.repo.FindByID("addresses", id, &address)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	return &address, nil
}

// GetCustomerAddresses retrieves all addresses for a customer
func (s *CustomerService) GetCustomerAddresses(customerID int64) ([]models.Address, error) {
	var addresses []models.Address
	conditions := map[string]interface{}{
		"customer_id": customerID,
		"is_active":   true,
	}
	err := s.repo.FindAll("addresses", &addresses, conditions, "is_default DESC, created_at DESC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer addresses: %w", err)
	}
	return addresses, nil
}

// GetDefaultAddress retrieves the default address for a customer
func (s *CustomerService) GetDefaultAddress(customerID int64) (*models.Address, error) {
	var address models.Address
	conditions := map[string]interface{}{
		"customer_id": customerID,
		"is_default":  true,
		"is_active":   true,
	}
	err := s.repo.FindOne("addresses", &address, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get default address: %w", err)
	}
	return &address, nil
}

// UpdateAddress updates an address
func (s *CustomerService) UpdateAddress(id int64, address *models.Address) error {
	if err := address.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	address.UpdatedAt = time.Now()
	
	if address.IsDefault {
		// Set other addresses as non-default
		if err := s.unsetDefaultAddresses(address.CustomerID); err != nil {
			return fmt.Errorf("failed to unset default addresses: %w", err)
		}
	}

	err := s.repo.Update("addresses", id, address)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}
	return nil
}

// DeleteAddress soft deletes an address
func (s *CustomerService) DeleteAddress(id int64) error {
	// Instead of hard delete, set is_active to false
	address, err := s.GetAddressByID(id)
	if err != nil {
		return err
	}
	
	address.IsActive = false
	address.UpdatedAt = time.Now()
	
	return s.repo.Update("addresses", id, address)
}

// SetDefaultAddress sets an address as default
func (s *CustomerService) SetDefaultAddress(customerID, addressID int64) error {
	// First, unset all default addresses for the customer
	if err := s.unsetDefaultAddresses(customerID); err != nil {
		return fmt.Errorf("failed to unset default addresses: %w", err)
	}
	
	// Then set the specified address as default
	address, err := s.GetAddressByID(addressID)
	if err != nil {
		return err
	}
	
	if address.CustomerID != customerID {
		return fmt.Errorf("address does not belong to customer")
	}
	
	address.IsDefault = true
	address.UpdatedAt = time.Now()
	
	return s.repo.Update("addresses", addressID, address)
}

// unsetDefaultAddresses sets all addresses for a customer as non-default
func (s *CustomerService) unsetDefaultAddresses(customerID int64) error {
	addresses, err := s.GetCustomerAddresses(customerID)
	if err != nil {
		return err
	}
	
	for _, addr := range addresses {
		if addr.IsDefault {
			addr.IsDefault = false
			addr.UpdatedAt = time.Now()
			if err := s.repo.Update("addresses", addr.ID, &addr); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// GetCustomerStats returns customer statistics
func (s *CustomerService) GetCustomerStats() (map[string]interface{}, error) {
	totalCustomers, err := s.repo.Count("customers", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}
	
	activeCustomers, err := s.repo.Count("customers", map[string]interface{}{"status": "active"})
	if err != nil {
		return nil, fmt.Errorf("failed to count active customers: %w", err)
	}
	
	stats := map[string]interface{}{
		"total_customers":  totalCustomers,
		"active_customers": activeCustomers,
	}
	
	return stats, nil
}