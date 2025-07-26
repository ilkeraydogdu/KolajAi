package models

import (
	"testing"
	"time"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{
			name: "valid user",
			user: User{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "hashedpassword123",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			user: User{
				Name:     "",
				Email:    "test@example.com",
				Password: "hashedpassword123",
			},
			wantErr: true,
		},
		{
			name: "empty email",
			user: User{
				Name:     "Test User",
				Email:    "",
				Password: "hashedpassword123",
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			user: User{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "hashedpassword123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			user: User{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	activeUser := User{
		Name:      "Active User",
		Email:     "active@example.com",
		Password:  "hashedpassword123",
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	inactiveUser := User{
		Name:      "Inactive User",
		Email:     "inactive@example.com",
		Password:  "hashedpassword123",
		IsActive:  false,
		CreatedAt: time.Now(),
	}

	if !activeUser.IsActive {
		t.Error("Expected active user to be active")
	}

	if inactiveUser.IsActive {
		t.Error("Expected inactive user to be inactive")
	}
}
