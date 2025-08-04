package repository

import (
	"database/sql"
	"fmt"
	"time"
	"kolajAi/internal/models"
	"kolajAi/internal/database"
)

// AdminRepository handles admin-specific database operations
type AdminRepository struct {
	*BaseRepository
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *database.MySQLRepository) *AdminRepository {
	return &AdminRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Dashboard Statistics

// GetDashboardStats returns dashboard statistics
func (r *AdminRepository) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total users
	totalUsers, err := r.Count("users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}
	stats["TotalUsers"] = totalUsers

	// Active users (logged in within last 30 days)
	var activeUsers int
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE is_active = 1 AND updated_at > DATE_SUB(NOW(), INTERVAL 30 DAY)
	`).Scan(&activeUsers)
	if err != nil {
		activeUsers = 0
	}
	stats["ActiveUsers"] = activeUsers

	// New users today
	var newUsersToday int
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE DATE(created_at) = CURDATE()
	`).Scan(&newUsersToday)
	if err != nil {
		newUsersToday = 0
	}
	stats["NewUsersToday"] = newUsersToday

	// Total orders
	totalOrders, err := r.Count("orders", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total orders: %w", err)
	}
	stats["TotalOrders"] = totalOrders

	// Pending orders
	pendingOrders, err := r.Count("orders", map[string]interface{}{"status": "pending"})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	stats["PendingOrders"] = pendingOrders

	// Total products
	totalProducts, err := r.Count("products", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get total products: %w", err)
	}
	stats["TotalProducts"] = totalProducts

	// Total revenue
	var totalRevenue float64
	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(total_amount), 0) FROM orders 
		WHERE payment_status = 'paid'
	`).Scan(&totalRevenue)
	if err != nil {
		totalRevenue = 0
	}
	stats["TotalRevenue"] = fmt.Sprintf("₺%.2f", totalRevenue)

	// Total sellers/vendors
	totalSellers, err := r.Count("users", map[string]interface{}{"is_seller": true})
	if err != nil {
		return nil, fmt.Errorf("failed to get total sellers: %w", err)
	}
	stats["TotalSellers"] = totalSellers

	// Active sellers
	activeSellers, err := r.Count("users", map[string]interface{}{
		"is_seller": true,
		"is_active": true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get active sellers: %w", err)
	}
	stats["ActiveSellers"] = activeSellers

	// Pending sellers (assuming there's an approval process)
	pendingSellers, err := r.Count("users", map[string]interface{}{
		"is_seller": true,
		"is_active": false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pending sellers: %w", err)
	}
	stats["PendingSellers"] = pendingSellers

	return stats, nil
}

// GetRecentOrders returns recent orders for dashboard
func (r *AdminRepository) GetRecentOrders(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT o.id, o.order_number, o.total_amount, o.status, o.created_at,
		       u.name as customer_name, u.email as customer_email
		FROM orders o
		JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var order struct {
			ID            int64     `db:"id"`
			OrderNumber   string    `db:"order_number"`
			TotalAmount   float64   `db:"total_amount"`
			Status        string    `db:"status"`
			CreatedAt     time.Time `db:"created_at"`
			CustomerName  string    `db:"customer_name"`
			CustomerEmail string    `db:"customer_email"`
		}

		err := rows.Scan(&order.ID, &order.OrderNumber, &order.TotalAmount, 
			&order.Status, &order.CreatedAt, &order.CustomerName, &order.CustomerEmail)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, map[string]interface{}{
			"ID":            fmt.Sprintf("ORD-%d", order.ID),
			"OrderNumber":   order.OrderNumber,
			"CustomerName":  order.CustomerName,
			"CustomerEmail": order.CustomerEmail,
			"Amount":        fmt.Sprintf("₺%.2f", order.TotalAmount),
			"Status":        order.Status,
			"Date":          order.CreatedAt.Format("2006-01-02"),
		})
	}

	return orders, nil
}

// GetRecentUsers returns recent users for dashboard
func (r *AdminRepository) GetRecentUsers(limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent users: %w", err)
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var user struct {
			ID        int64     `db:"id"`
			Name      string    `db:"name"`
			Email     string    `db:"email"`
			CreatedAt time.Time `db:"created_at"`
		}

		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, map[string]interface{}{
			"ID":    user.ID,
			"Name":  user.Name,
			"Email": user.Email,
			"Date":  user.CreatedAt.Format("2006-01-02"),
		})
	}

	return users, nil
}

// User Management

// GetUsers returns paginated users with filters
func (r *AdminRepository) GetUsers(page, limit int, filters map[string]interface{}) ([]models.User, int64, error) {
	offset := (page - 1) * limit
	
	// Build where clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	
	if status, ok := filters["status"]; ok && status != "" {
		whereClause += " AND is_active = ?"
		args = append(args, status == "active")
	}
	
	if role, ok := filters["role"]; ok && role != "" {
		switch role {
		case "admin":
			whereClause += " AND is_admin = 1"
		case "seller":
			whereClause += " AND is_seller = 1"
		case "user":
			whereClause += " AND is_admin = 0 AND is_seller = 0"
		}
	}
	
	if search, ok := filters["search"]; ok && search != "" {
		whereClause += " AND (name LIKE ? OR email LIKE ?)"
		searchTerm := "%" + search.(string) + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM users " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	// Get users
	query := fmt.Sprintf(`
		SELECT id, name, email, phone, role, is_active, is_admin, is_seller, created_at, updated_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, 
			&user.Role, &user.IsActive, &user.IsAdmin, &user.IsSeller,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// UpdateUserStatus updates user status
func (r *AdminRepository) UpdateUserStatus(userID int64, isActive bool) error {
	query := "UPDATE users SET is_active = ?, updated_at = NOW() WHERE id = ?"
	_, err := r.db.Exec(query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

// DeleteUser soft deletes a user
func (r *AdminRepository) DeleteUser(userID int64) error {
	// Instead of hard delete, we deactivate the user
	query := "UPDATE users SET is_active = 0, updated_at = NOW() WHERE id = ?"
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// Order Management

// GetOrders returns paginated orders with filters
func (r *AdminRepository) GetOrders(page, limit int, filters map[string]interface{}) ([]map[string]interface{}, int64, error) {
	offset := (page - 1) * limit
	
	// Build where clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	
	if status, ok := filters["status"]; ok && status != "" {
		whereClause += " AND o.status = ?"
		args = append(args, status)
	}
	
	if paymentStatus, ok := filters["payment_status"]; ok && paymentStatus != "" {
		whereClause += " AND o.payment_status = ?"
		args = append(args, paymentStatus)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM orders o " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get order count: %w", err)
	}

	// Get orders
	query := fmt.Sprintf(`
		SELECT o.id, o.order_number, o.user_id, o.status, o.payment_status, 
		       o.total_amount, o.created_at, u.name as customer_name, u.email as customer_email
		FROM orders o
		JOIN users u ON o.user_id = u.id
		%s
		ORDER BY o.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []map[string]interface{}
	for rows.Next() {
		var order struct {
			ID            int64     `db:"id"`
			OrderNumber   string    `db:"order_number"`
			UserID        int64     `db:"user_id"`
			Status        string    `db:"status"`
			PaymentStatus string    `db:"payment_status"`
			TotalAmount   float64   `db:"total_amount"`
			CreatedAt     time.Time `db:"created_at"`
			CustomerName  string    `db:"customer_name"`
			CustomerEmail string    `db:"customer_email"`
		}

		err := rows.Scan(&order.ID, &order.OrderNumber, &order.UserID,
			&order.Status, &order.PaymentStatus, &order.TotalAmount,
			&order.CreatedAt, &order.CustomerName, &order.CustomerEmail)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		orders = append(orders, map[string]interface{}{
			"ID":            order.ID,
			"OrderNumber":   order.OrderNumber,
			"CustomerName":  order.CustomerName,
			"CustomerEmail": order.CustomerEmail,
			"Amount":        fmt.Sprintf("₺%.2f", order.TotalAmount),
			"Status":        order.Status,
			"PaymentStatus": order.PaymentStatus,
			"CreatedAt":     order.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	return orders, total, nil
}

// UpdateOrderStatus updates order status
func (r *AdminRepository) UpdateOrderStatus(orderID int64, status string) error {
	query := "UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?"
	_, err := r.db.Exec(query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	return nil
}

// Product Management

// GetProducts returns paginated products with filters
func (r *AdminRepository) GetProducts(page, limit int, filters map[string]interface{}) ([]map[string]interface{}, int64, error) {
	offset := (page - 1) * limit
	
	// Build where clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	
	if status, ok := filters["status"]; ok && status != "" {
		whereClause += " AND p.status = ?"
		args = append(args, status)
	}
	
	if categoryID, ok := filters["category_id"]; ok && categoryID != "" {
		whereClause += " AND p.category_id = ?"
		args = append(args, categoryID)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM products p " + whereClause
	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get product count: %w", err)
	}

	// Get products
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.sku, p.price, p.stock, p.status, p.created_at,
		       u.name as vendor_name, c.name as category_name
		FROM products p
		LEFT JOIN users u ON p.vendor_id = u.id
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var product struct {
			ID           int       `db:"id"`
			Name         string    `db:"name"`
			SKU          string    `db:"sku"`
			Price        float64   `db:"price"`
			Stock        int       `db:"stock"`
			Status       string    `db:"status"`
			CreatedAt    time.Time `db:"created_at"`
			VendorName   sql.NullString `db:"vendor_name"`
			CategoryName sql.NullString `db:"category_name"`
		}

		err := rows.Scan(&product.ID, &product.Name, &product.SKU,
			&product.Price, &product.Stock, &product.Status,
			&product.CreatedAt, &product.VendorName, &product.CategoryName)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		vendorName := "N/A"
		if product.VendorName.Valid {
			vendorName = product.VendorName.String
		}

		categoryName := "N/A"
		if product.CategoryName.Valid {
			categoryName = product.CategoryName.String
		}

		products = append(products, map[string]interface{}{
			"ID":          product.ID,
			"Name":        product.Name,
			"SKU":         product.SKU,
			"Price":       fmt.Sprintf("₺%.2f", product.Price),
			"Stock":       product.Stock,
			"Status":      product.Status,
			"Category":    categoryName,
			"VendorName":  vendorName,
			"CreatedAt":   product.CreatedAt.Format("2006-01-02"),
		})
	}

	return products, total, nil
}

// GetSystemHealth returns system health metrics
func (r *AdminRepository) GetSystemHealth() (map[string]interface{}, error) {
	health := make(map[string]interface{})

	// Database connection test
	err := database.DB.Ping()
	if err != nil {
		health["DatabaseStatus"] = "disconnected"
		health["OverallStatus"] = "unhealthy"
		health["HealthScore"] = 0
	} else {
		health["DatabaseStatus"] = "connected"
		health["OverallStatus"] = "healthy"
		health["HealthScore"] = 95 // This could be calculated based on various metrics
	}

	// Get database size
	var dbSize float64
	err = r.db.QueryRow(`
		SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 1) as db_size_mb
		FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`).Scan(&dbSize)
	if err != nil {
		dbSize = 0
	}
	
	health["DatabaseSize"] = fmt.Sprintf("%.1f MB", dbSize)

	// Get table count
	var tableCount int
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = DATABASE()
	`).Scan(&tableCount)
	if err != nil {
		tableCount = 0
	}
	
	health["TableCount"] = tableCount

	return health, nil
}