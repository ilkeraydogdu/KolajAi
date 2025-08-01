package migrations

// AddRoleToUsers migration adds role column to users table
var AddRoleToUsers = `
-- Add role column to users table
ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'user';

-- Update existing admin users
UPDATE users SET role = 'admin' WHERE is_admin = 1;

-- Create index on role for faster queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
`