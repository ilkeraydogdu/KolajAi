package models

import "time"

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalUsers       int64   `json:"total_users"`
	ActiveUsers      int64   `json:"active_users"`
	TotalProducts    int64   `json:"total_products"`
	TotalOrders      int64   `json:"total_orders"`
	TotalRevenue     float64 `json:"total_revenue"`
	NewUsersToday    int64   `json:"new_users_today"`
	OrdersToday      int64   `json:"orders_today"`
	RevenueToday     float64 `json:"revenue_today"`
	PendingTasks     int64   `json:"pending_tasks"`
	UnreadNotifications int64 `json:"unread_notifications"`
}

// DashboardChart represents chart data for dashboard
type DashboardChart struct {
	Labels []string      `json:"labels"`
	Data   []interface{} `json:"data"`
	Type   string        `json:"type"`
}

// UserActivity represents user activity data
type UserActivity struct {
	Date        time.Time `json:"date"`
	LoginCount  int       `json:"login_count"`
	ActiveUsers int       `json:"active_users"`
}

// RevenueData represents revenue chart data
type RevenueData struct {
	Date    time.Time `json:"date"`
	Revenue float64   `json:"revenue"`
	Orders  int       `json:"orders"`
}

// Task represents a user task
type Task struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"` // pending, completed, cancelled
	Priority    string    `json:"priority" db:"priority"` // low, medium, high
	DueDate     *time.Time `json:"due_date" db:"due_date"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// QuickAction represents a quick action for dashboard
type QuickAction struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	URL         string `json:"url"`
	Color       string `json:"color"`
	Badge       string `json:"badge,omitempty"`
}

// DashboardWidget represents a dashboard widget configuration
type DashboardWidget struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"` // stats, chart, list, table
	Title    string      `json:"title"`
	Size     string      `json:"size"` // small, medium, large, full
	Position int         `json:"position"`
	Data     interface{} `json:"data"`
	Settings map[string]interface{} `json:"settings"`
}