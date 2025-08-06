package models

import (
	"time"
	"encoding/json"
)

// AdminLog represents an admin action log entry
type AdminLog struct {
	ID          int64           `json:"id" db:"id"`
	AdminID     int64           `json:"admin_id" db:"admin_id"`
	AdminEmail  string          `json:"admin_email" db:"admin_email"`
	Action      string          `json:"action" db:"action"`
	Resource    string          `json:"resource" db:"resource"`
	ResourceID  *int64          `json:"resource_id,omitempty" db:"resource_id"`
	OldValue    json.RawMessage `json:"old_value,omitempty" db:"old_value"`
	NewValue    json.RawMessage `json:"new_value,omitempty" db:"new_value"`
	IPAddress   string          `json:"ip_address" db:"ip_address"`
	UserAgent   string          `json:"user_agent" db:"user_agent"`
	Status      string          `json:"status" db:"status"` // success, failed
	ErrorMsg    *string         `json:"error_msg,omitempty" db:"error_msg"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// AdminAction constants
const (
	ActionUserCreate     = "user.create"
	ActionUserUpdate     = "user.update"
	ActionUserDelete     = "user.delete"
	ActionUserBan        = "user.ban"
	ActionUserUnban      = "user.unban"
	ActionUserActivate   = "user.activate"
	ActionUserDeactivate = "user.deactivate"
	
	ActionProductCreate  = "product.create"
	ActionProductUpdate  = "product.update"
	ActionProductDelete  = "product.delete"
	ActionProductApprove = "product.approve"
	ActionProductReject  = "product.reject"
	
	ActionOrderUpdate    = "order.update"
	ActionOrderCancel    = "order.cancel"
	ActionOrderRefund    = "order.refund"
	
	ActionSellerApprove  = "seller.approve"
	ActionSellerReject   = "seller.reject"
	ActionSellerSuspend  = "seller.suspend"
	
	ActionSystemBackup   = "system.backup"
	ActionSystemRestore  = "system.restore"
	ActionSettingsUpdate = "settings.update"
)

// AdminPermission represents admin permissions
type AdminPermission struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AdminRole represents admin roles
type AdminRole struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AdminRolePermission represents the many-to-many relationship
type AdminRolePermission struct {
	RoleID       int64     `json:"role_id" db:"role_id"`
	PermissionID int64     `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserRole represents user-role assignments
type UserRole struct {
	UserID    int64     `json:"user_id" db:"user_id"`
	RoleID    int64     `json:"role_id" db:"role_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}