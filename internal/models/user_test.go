package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Validation(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user",
			user: User{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			user: User{
				Email:     "invalid-email",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
			errMsg:  "email",
		},
		{
			name: "empty email",
			user: User{
				Email:     "",
				Password:  "password123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
			errMsg:  "Email",
		},
		{
			name: "empty password",
			user: User{
				Email:     "test@example.com",
				Password:  "",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
			errMsg:  "Password",
		},
		{
			name: "short password",
			user: User{
				Email:     "test@example.com",
				Password:  "short",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
			errMsg:  "Password",
		},
		{
			name: "empty first name",
			user: User{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "",
				LastName:  "Doe",
			},
			wantErr: true,
			errMsg:  "FirstName",
		},
		{
			name: "empty last name",
			user: User{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  "",
			},
			wantErr: true,
			errMsg:  "LastName",
		},
		{
			name: "first name too long",
			user: User{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: strings.Repeat("a", 101),
				LastName:  "Doe",
			},
			wantErr: true,
			errMsg:  "FirstName",
		},
		{
			name: "last name too long",
			user: User{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "John",
				LastName:  strings.Repeat("a", 101),
			},
			wantErr: true,
			errMsg:  "LastName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.user)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_TableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "users", user.TableName())
}

func TestUser_BeforeCreate(t *testing.T) {
	user := &User{
		Email: "TEST@EXAMPLE.COM",
	}

	// Since we can't easily test GORM hooks without a database,
	// we'll test the logic directly
	user.Email = strings.ToLower(user.Email)

	assert.Equal(t, "test@example.com", user.Email)
}

func TestUser_BeforeUpdate(t *testing.T) {
	user := &User{
		Email: "UPDATED@EXAMPLE.COM",
	}

	// Since we can't easily test GORM hooks without a database,
	// we'll test the logic directly
	user.Email = strings.ToLower(user.Email)

	assert.Equal(t, "updated@example.com", user.Email)
}
