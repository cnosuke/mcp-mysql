package server

import (
	"context"
	"fmt"

	"github.com/cnosuke/mcp-mysql/config"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// DB connection
	DB *sqlx.DB
)

// Run - Execute the MCP server
func Run(cfg *config.Config, name string, version string, revision string) error {
	zap.S().Infow("starting MCP MySQL Server")

	// Format version string with revision if available
	versionString := version
	if revision != "" && revision != "xxx" {
		versionString = versionString + " (" + revision + ")"
	}

	// Create custom hooks for error handling
	hooks := &server.Hooks{}
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		zap.S().Errorw("MCP error occurred",
			"id", id,
			"method", method,
			"error", err,
		)
	})

	// Create MCP server with server name and version
	zap.S().Debugw("creating MCP server",
		"name", name,
		"version", versionString,
	)
	mcpServer := server.NewMCPServer(
		name,
		versionString,
		server.WithHooks(hooks),
	)

	// Register all tools
	zap.S().Debugw("registering MySQL tools")
	if err := RegisterAllTools(mcpServer, cfg); err != nil {
		zap.S().Errorw("failed to register tools", "error", err)
		return err
	}

	// Start the server with stdio transport
	zap.S().Infow("starting MCP server")
	err := server.ServeStdio(mcpServer)
	if err != nil {
		zap.S().Errorw("failed to start server", "error", err)
		return errors.Wrap(err, "failed to start server")
	}

	// ServeStdio will block until the server is terminated
	zap.S().Infow("server shutting down")
	return nil
}

// GetDB - Get database connection
func GetDB(cfg *config.Config, toolDSN string) (*sqlx.DB, error) {
	// Reset DB connection if new DSN is provided
	if toolDSN != "" {
		DB = nil
	}
	
	if DB != nil {
		return DB, nil
	}

	// Determine which DSN to use - tool parameter takes precedence
	dsn := toolDSN
	if dsn == "" {
		dsn = cfg.MySQL.DSN
		// If config DSN is empty, build it from individual parameters
		if dsn == "" {
			// If no connection parameters are provided anywhere, return error
			if cfg.MySQL.Host == "" || cfg.MySQL.User == "" {
				return nil, fmt.Errorf("MySQL connection information is required. Please provide a valid DSN parameter or configure MySQL connection in config file")
			}
			
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
				cfg.MySQL.User,
				cfg.MySQL.Password,
				cfg.MySQL.Host,
				cfg.MySQL.Port,
				cfg.MySQL.Database)
		}
	}

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to establish database connection: %v", err)
	}

	DB = db
	return DB, nil
}
