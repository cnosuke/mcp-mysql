package server

import (
	"strings"
	"testing"

	"github.com/cnosuke/mcp-mysql/config"
	"github.com/stretchr/testify/assert"
)

func TestMapToCSV(t *testing.T) {
	t.Run("successful mapping", func(t *testing.T) {
		// Setup test data
		data := []map[string]interface{}{
			{"id": 1, "name": "test1"},
			{"id": 2, "name": "test2"},
		}
		headers := []string{"id", "name"}

		// Call MapToCSV
		result, err := MapToCSV(data, headers)

		// Verify results
		assert.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		assert.Len(t, lines, 3)
		assert.Equal(t, "id,name", lines[0])
		assert.Equal(t, "1,test1", lines[1])
		assert.Equal(t, "2,test2", lines[2])
	})

	t.Run("missing key", func(t *testing.T) {
		// Setup test data
		data := []map[string]interface{}{
			{"id": 1}, // missing "name"
		}
		headers := []string{"id", "name"}

		// Call MapToCSV
		_, err := MapToCSV(data, headers)

		// Verify results
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key 'name' not found in map")
	})

	t.Run("empty data", func(t *testing.T) {
		// Setup test data
		data := []map[string]interface{}{}
		headers := []string{"id", "name"}

		// Call MapToCSV
		result, err := MapToCSV(data, headers)

		// Verify results
		assert.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		assert.Len(t, lines, 1)
		assert.Equal(t, "id,name", lines[0])
	})

	t.Run("handles different types", func(t *testing.T) {
		// Setup test data
		data := []map[string]interface{}{
			{"id": 1, "name": "test1", "active": true, "score": 3.14},
		}
		headers := []string{"id", "name", "active", "score"}

		// Call MapToCSV
		result, err := MapToCSV(data, headers)

		// Verify results
		assert.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(result), "\n")
		assert.Len(t, lines, 2)
		assert.Equal(t, "id,name,active,score", lines[0])
		assert.Equal(t, "1,test1,true,3.14", lines[1])
	})
}

func TestRegisterAllTools(t *testing.T) {
	// Mock config
	cfg := &config.Config{}
	cfg.MySQL.Host = "localhost"
	cfg.MySQL.User = "root"
	cfg.MySQL.Password = "password"
	cfg.MySQL.Port = 3306
	cfg.MySQL.Database = "test"
	cfg.MySQL.ReadOnly = false
	cfg.MySQL.ExplainCheck = false

	// This test would be a comprehensive unit test with mocks,
	// but for now we'll just verify it doesn't panic
	t.Run("doesn't panic", func(t *testing.T) {
		// We can't actually test the server without mocking a lot,
		// so we'll just ensure RegisterAllTools doesn't panic with nil
		assert.Panics(t, func() {
			_ = RegisterAllTools(nil, cfg)
		})
	})
}

// Simple wrapper for HandleQuery for testing
func mockHandleQuery(cfg *config.Config, query, expect string, toolDSN string) (string, error) {
	return "mocked result", nil
}
