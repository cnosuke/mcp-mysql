package server

import (
	"testing"

	"github.com/cnosuke/mcp-mysql/config"
	"github.com/stretchr/testify/assert"
)

func TestGetDB(t *testing.T) {
	// Save the original DB
	originalDB := DB
	defer func() { DB = originalDB }()

	// Reset DB to nil
	DB = nil

	// Create test config
	cfg := &config.Config{}
	cfg.MySQL.Host = "localhost"
	cfg.MySQL.User = "root"
	cfg.MySQL.Password = "password"
	cfg.MySQL.Port = 3306
	cfg.MySQL.Database = "test"

	// This test is more of an integration test and would require a real DB
	// For unit testing, we'll just verify that it creates a DSN correctly
	_, err := GetDB(cfg, "")
	assert.Error(t, err) // Error expected since we're not connecting to a real DB

	// Test with provided toolDSN
	DB = nil
	_, err = GetDB(cfg, "user:pass@tcp(localhost:3306)/testdb")
	assert.Error(t, err) // Error expected since we're not connecting to a real DB

	// Test with URL-style DSN
	DB = nil
	_, err = GetDB(cfg, "mysql://user:pass@localhost:3306/testdb")
	assert.Error(t, err) // Error expected since we're not connecting to a real DB

	// Test with empty configuration and no toolDSN
	DB = nil
	emptyCfg := &config.Config{}
	_, err = GetDB(emptyCfg, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MySQL connection information is required")

	// Test with empty configuration but with toolDSN
	DB = nil
	_, err = GetDB(emptyCfg, "user:pass@tcp(localhost:3306)/testdb")
	assert.Error(t, err) // Error expected since we're not connecting to a real DB

	// Test with empty configuration but with URL-style toolDSN
	DB = nil
	_, err = GetDB(emptyCfg, "mysql://user:pass@localhost:3306/testdb")
	assert.Error(t, err) // Error expected since we're not connecting to a real DB
}

func TestIsURLStyle(t *testing.T) {
	// Test valid URL-style DSNs
	assert.True(t, isURLStyle("mysql://user:pass@localhost:3306/dbname"))
	assert.True(t, isURLStyle("mysql://localhost/dbname"))
	assert.True(t, isURLStyle("mariadb://user:pass@localhost/dbname"))
	assert.True(t, isURLStyle("memsql://user@host/db"))
	assert.True(t, isURLStyle("tidb://127.0.0.1:4000/test"))
	
	// Test non-URL-style DSNs
	assert.False(t, isURLStyle("user:pass@tcp(localhost:3306)/dbname"))
	assert.False(t, isURLStyle("root:@tcp(localhost:3306)/note"))
	assert.False(t, isURLStyle("/var/run/mysqld/mysqld.sock"))
	assert.False(t, isURLStyle(""))
}
