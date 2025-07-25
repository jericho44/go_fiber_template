package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigration_Validation(t *testing.T) {
	tests := []struct {
		name      string
		migration Migration
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid migration",
			migration: Migration{
				Version: "20240101_create_users_table",
			},
			wantErr: false,
		},
		{
			name: "empty version",
			migration: Migration{
				Version: "",
			},
			wantErr: true,
			errMsg:  "Version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.migration)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMigration_TableName(t *testing.T) {
	migration := Migration{}
	assert.Equal(t, "migrations", migration.TableName())
}
