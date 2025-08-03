package services

import (
	"fmt"
	"kolajAi/internal/database"
	"kolajAi/internal/models"
	"time"
)

type OrderService struct {
	repo database.SimpleRepository
}

func NewOrderService(repo database.SimpleRepository) *OrderService {
	return &OrderService{repo: repo}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(order *models.Order) error {
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	if order.Status == "" {
		order.Status = "pending"
	}
	if order.PaymentStatus == "" {
		order.PaymentStatus = "pending"
	}
	if order.Currency == "" {
		order.Currency = "TRY"
	}

	// Generate order number if not provided
	if order.OrderNumber == "" {
		order.OrderNumber = s.generateOrderNumber()
	}

	id, err := s.repo.CreateStruct("orders", order)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = id
	return nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(id int) (*models.Order, error) {
	var order models.Order
	err := s.repo.FindByID("orders", id, &order)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return &order, nil
}

// GetOrderByNumber retrieves an order by order number
func (s *OrderService) GetOrderByNumber(orderNumber string) (*models.Order, error) {
	var order models.Order
	conditions := map[string]interface{}{"order_number": orderNumber}
	err := s.repo.FindOne("orders", &order, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get order by number: %w", err)
	}
	return &order, nil
}

// UpdateOrder updates an order
func (s *OrderService) UpdateOrder(id int, order *models.Order) error {
	order.UpdatedAt = time.Now()
	err := s.repo.Update("orders", id, order)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

// GetOrdersByUser retrieves orders by user ID
func (s *OrderService) GetOrdersByUser(userID int, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	conditions := map[string]interface{}{"user_id": userID}

	err := s.repo.FindAll("orders", &orders, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by user: %w", err)
	}
	return orders, nil
}

// GetOrdersByStatus retrieves orders by status
func (s *OrderService) GetOrdersByStatus(status string, limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	conditions := map[string]interface{}{"status": status}

	err := s.repo.FindAll("orders", &orders, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	return orders, nil
}

// AddOrderItem adds an item to an order
func (s *OrderService) AddOrderItem(item *models.OrderItem) error {
	id, err := s.repo.CreateStruct("order_items", item)
	if err != nil {
		return fmt.Errorf("failed to add order item: %w", err)
	}
	item.ID = id
	return nil
}

// GetOrderItems retrieves items for an order
func (s *OrderService) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	var items []models.OrderItem
	conditions := map[string]interface{}{"order_id": orderID}

	err := s.repo.FindAll("order_items", &items, conditions, "id ASC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	return items, nil
}

// AddOrderAddress adds an address to an order
func (s *OrderService) AddOrderAddress(address *models.OrderAddress) error {
	id, err := s.repo.CreateStruct("order_addresses", address)
	if err != nil {
		return fmt.Errorf("failed to add order address: %w", err)
	}
	address.ID = id
	return nil
}

// GetOrderAddresses retrieves addresses for an order
func (s *OrderService) GetOrderAddresses(orderID int) ([]models.OrderAddress, error) {
	var addresses []models.OrderAddress
	conditions := map[string]interface{}{"order_id": orderID}

	err := s.repo.FindAll("order_addresses", &addresses, conditions, "type ASC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get order addresses: %w", err)
	}
	return addresses, nil
}

// ConfirmOrder confirms an order
func (s *OrderService) ConfirmOrder(orderID int) error {
	order := &models.Order{
		Status:    "confirmed",
		UpdatedAt: time.Now(),
	}
	return s.UpdateOrder(orderID, order)
}

// ShipOrder marks an order as shipped
func (s *OrderService) ShipOrder(orderID int, trackingNumber string) error {
	now := time.Now()
	order := &models.Order{
		Status:         "shipped",
		TrackingNumber: trackingNumber,
		ShippedAt:      &now,
		UpdatedAt:      now,
	}
	return s.UpdateOrder(orderID, order)
}

// DeliverOrder marks an order as delivered
func (s *OrderService) DeliverOrder(orderID int) error {
	now := time.Now()
	order := &models.Order{
		Status:      "delivered",
		DeliveredAt: &now,
		UpdatedAt:   now,
	}
	return s.UpdateOrder(orderID, order)
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(orderID int) error {
	order := &models.Order{
		Status:    "cancelled",
		UpdatedAt: time.Now(),
	}
	return s.UpdateOrder(orderID, order)
}

// UpdatePaymentStatus updates the payment status of an order
func (s *OrderService) UpdatePaymentStatus(orderID int, paymentStatus string) error {
	order := &models.Order{
		PaymentStatus: paymentStatus,
		UpdatedAt:     time.Now(),
	}
	return s.UpdateOrder(orderID, order)
}

// GetOrderStats returns order statistics
func (s *OrderService) GetOrderStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total orders
	totalOrders, err := s.repo.Count("orders", nil)
	if err == nil {
		stats["total_orders"] = totalOrders
	}

	// Pending orders
	pendingOrders, err := s.repo.Count("orders", map[string]interface{}{"status": "pending"})
	if err == nil {
		stats["pending_orders"] = pendingOrders
	}

	// Confirmed orders
	confirmedOrders, err := s.repo.Count("orders", map[string]interface{}{"status": "confirmed"})
	if err == nil {
		stats["confirmed_orders"] = confirmedOrders
	}

	// Shipped orders
	shippedOrders, err := s.repo.Count("orders", map[string]interface{}{"status": "shipped"})
	if err == nil {
		stats["shipped_orders"] = shippedOrders
	}

	// Delivered orders
	deliveredOrders, err := s.repo.Count("orders", map[string]interface{}{"status": "delivered"})
	if err == nil {
		stats["delivered_orders"] = deliveredOrders
	}

	return stats, nil
}

// GetUserOrderStats returns order statistics for a specific user
func (s *OrderService) GetUserOrderStats(userID int) (*models.OrderStats, error) {
	stats := &models.OrderStats{}

	// This is a placeholder implementation
	// You would implement actual database queries here
	stats.TotalOrders = 0
	stats.PendingOrders = 0
	stats.ConfirmedOrders = 0
	stats.DeliveredOrders = 0
	stats.TotalRevenue = 0.0
	stats.AverageValue = 0.0

	return stats, nil
}

// GetUserOrders retrieves orders for a specific user
func (s *OrderService) GetUserOrders(userID, page, limit int, status, dateRange string) ([]*models.Order, int, error) {
	// This is a placeholder implementation
	// You would implement actual database queries here with filtering by status and dateRange
	orders := []*models.Order{}
	totalCount := 0
	
	return orders, totalCount, nil
}

// GetVendorOrders retrieves orders for a vendor
func (s *OrderService) GetVendorOrders(vendorID int, limit, offset int) ([]models.OrderItem, error) {
	var items []models.OrderItem
	conditions := map[string]interface{}{"vendor_id": vendorID}

	err := s.repo.FindAll("order_items", &items, conditions, "id DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendor orders: %w", err)
	}
	return items, nil
}

// generateOrderNumber generates a unique order number
func (s *OrderService) generateOrderNumber() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("ORD-%d", timestamp)
}

// Cart Management

// CreateCart creates a new cart
func (s *OrderService) CreateCart(cart *models.Cart) error {
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = time.Now()

	id, err := s.repo.CreateStruct("carts", cart)
	if err != nil {
		return fmt.Errorf("failed to create cart: %w", err)
	}
	cart.ID = int(id)
	return nil
}

// GetCartByUser retrieves cart by user ID
func (s *OrderService) GetCartByUser(userID int) (*models.Cart, error) {
	var cart models.Cart
	conditions := map[string]interface{}{"user_id": userID}
	err := s.repo.FindOne("carts", &cart, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart by user: %w", err)
	}
	return &cart, nil
}

// GetCartBySession retrieves cart by session ID
func (s *OrderService) GetCartBySession(sessionID string) (*models.Cart, error) {
	var cart models.Cart
	conditions := map[string]interface{}{"session_id": sessionID}
	err := s.repo.FindOne("carts", &cart, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart by session: %w", err)
	}
	return &cart, nil
}

// AddCartItem adds an item to cart
func (s *OrderService) AddCartItem(item *models.CartItem) error {
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	// Check if item already exists in cart
	existing, err := s.GetCartItem(item.CartID, item.ProductID, item.VariantID)
	if err == nil {
		// Update quantity
		existing.Quantity += item.Quantity
		existing.UpdatedAt = time.Now()
		return s.repo.Update("cart_items", existing.ID, existing)
	}

	id, err := s.repo.CreateStruct("cart_items", item)
	if err != nil {
		return fmt.Errorf("failed to add cart item: %w", err)
	}
	item.ID = int(id)
	return nil
}

// GetCartItem retrieves a specific cart item
func (s *OrderService) GetCartItem(cartID, productID int, variantID *int) (*models.CartItem, error) {
	var item models.CartItem
	conditions := map[string]interface{}{
		"cart_id":    cartID,
		"product_id": productID,
	}
	if variantID != nil {
		conditions["variant_id"] = *variantID
	}

	err := s.repo.FindOne("cart_items", &item, conditions)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}
	return &item, nil
}

// GetCartItems retrieves all items in a cart
func (s *OrderService) GetCartItems(cartID int) ([]models.CartItem, error) {
	var items []models.CartItem
	conditions := map[string]interface{}{"cart_id": cartID}

	err := s.repo.FindAll("cart_items", &items, conditions, "created_at ASC", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart items: %w", err)
	}
	return items, nil
}

// UpdateCartItem updates a cart item
func (s *OrderService) UpdateCartItem(itemID int, item *models.CartItem) error {
	item.UpdatedAt = time.Now()
	err := s.repo.Update("cart_items", itemID, item)
	if err != nil {
		return fmt.Errorf("failed to update cart item: %w", err)
	}
	return nil
}

// RemoveCartItem removes an item from cart
func (s *OrderService) RemoveCartItem(itemID int) error {
	err := s.repo.Delete("cart_items", itemID)
	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}
	return nil
}

// ClearCart removes all items from a cart
func (s *OrderService) ClearCart(cartID int) error {
	items, err := s.GetCartItems(cartID)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := s.RemoveCartItem(item.ID); err != nil {
			return err
		}
	}

	return nil
}

// GetAllOrders returns all orders with pagination
func (s *OrderService) GetAllOrders(limit, offset int) ([]models.Order, error) {
	var orders []models.Order
	conditions := map[string]interface{}{}
	err := s.repo.FindAll("orders", &orders, conditions, "created_at DESC", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	return orders, nil
}

// GetOrderCount returns the total number of orders
func (s *OrderService) GetOrderCount() (int64, error) {
	conditions := map[string]interface{}{}
	return s.repo.Count("orders", conditions)
}
