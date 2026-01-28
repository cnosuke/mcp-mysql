package server

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/cnosuke/mcp-mysql/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	StatementTypeNoExplainCheck = ""
	StatementTypeSelect         = "SELECT"
	StatementTypeInsert         = "INSERT"
	StatementTypeUpdate         = "UPDATE"
	StatementTypeDelete         = "DELETE"
)

// ExplainResult represents the output of EXPLAIN query
type ExplainResult struct {
	Id           *string `db:"id"`
	SelectType   *string `db:"select_type"`
	Table        *string `db:"table"`
	Partitions   *string `db:"partitions"`
	Type         *string `db:"type"`
	PossibleKeys *string `db:"possible_keys"`
	Key          *string `db:"key"`
	KeyLen       *string `db:"key_len"`
	Ref          *string `db:"ref"`
	Rows         *string `db:"rows"`
	Filtered     *string `db:"filtered"`
	Extra        *string `db:"Extra"`
}

// ShowCreateTableResult represents the output of SHOW CREATE TABLE query
type ShowCreateTableResult struct {
	Table       string `db:"Table"`
	CreateTable string `db:"Create Table"`
}

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *server.MCPServer, cfg *config.Config) error {
	// Schema Tools
	listDatabaseTool := mcp.NewTool(
		"list_database",
		mcp.WithDescription("List all databases in the MySQL server"),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	listTableTool := mcp.NewTool(
		"list_table",
		mcp.WithDescription("List all tables in the MySQL server"),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	createTableTool := mcp.NewTool(
		"create_table",
		mcp.WithDescription("Create a new table in the MySQL server. Make sure you have added proper comments for each column and the table itself"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to create the table"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	alterTableTool := mcp.NewTool(
		"alter_table",
		mcp.WithDescription("Alter an existing table in the MySQL server. Make sure you have updated comments for each modified column. DO NOT drop table or existing columns!"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to alter the table"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	descTableTool := mcp.NewTool(
		"desc_table",
		mcp.WithDescription("Describe the structure of a table"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("The name of the table to describe"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	// Data Tools
	readQueryTool := mcp.NewTool(
		"read_query",
		mcp.WithDescription("Execute a read-only SQL query. Make sure you have knowledge of the table structure before writing WHERE conditions. Call `desc_table` first if necessary"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	writeQueryTool := mcp.NewTool(
		"write_query",
		mcp.WithDescription("Execute a write SQL query. Make sure you have knowledge of the table structure before executing the query. Make sure the data types match the columns' definitions"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	updateQueryTool := mcp.NewTool(
		"update_query",
		mcp.WithDescription("Execute an update SQL query. Make sure you have knowledge of the table structure before executing the query. Make sure there is always a WHERE condition. Call `desc_table` first if necessary"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	deleteQueryTool := mcp.NewTool(
		"delete_query",
		mcp.WithDescription("Execute a delete SQL query. Make sure you have knowledge of the table structure before executing the query. Make sure there is always a WHERE condition. Call `desc_table` first if necessary"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The SQL query to execute"),
		),
		mcp.WithString("dsn",
			mcp.Description("MySQL DSN (Data Source Name) string. If provided, this overrides the configuration."),
		),
	)

	// Register handlers for each tool
	mcpServer.AddTool(listDatabaseTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dsn := request.GetString("dsn", "")
		result, err := HandleQuery(cfg, "SHOW DATABASES", StatementTypeNoExplainCheck, dsn)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	mcpServer.AddTool(listTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dsn := request.GetString("dsn", "")
		result, err := HandleQuery(cfg, "SHOW TABLES", StatementTypeNoExplainCheck, dsn)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	if !cfg.MySQL.ReadOnly {
		mcpServer.AddTool(createTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := request.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			dsn := request.GetString("dsn", "")
			result, err := HandleExec(cfg, query, StatementTypeNoExplainCheck, dsn)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	if !cfg.MySQL.ReadOnly {
		mcpServer.AddTool(alterTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := request.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			dsn := request.GetString("dsn", "")
			result, err := HandleExec(cfg, query, StatementTypeNoExplainCheck, dsn)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	mcpServer.AddTool(descTableTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dsn := request.GetString("dsn", "")
		result, err := HandleDescTable(cfg, name, dsn)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	mcpServer.AddTool(readQueryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		dsn := request.GetString("dsn", "")
		result, err := HandleQuery(cfg, query, StatementTypeSelect, dsn)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(result), nil
	})

	if !cfg.MySQL.ReadOnly {
		mcpServer.AddTool(writeQueryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := request.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			dsn := request.GetString("dsn", "")
			result, err := HandleExec(cfg, query, StatementTypeInsert, dsn)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	if !cfg.MySQL.ReadOnly {
		mcpServer.AddTool(updateQueryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := request.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			dsn := request.GetString("dsn", "")
			result, err := HandleExec(cfg, query, StatementTypeUpdate, dsn)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	if !cfg.MySQL.ReadOnly {
		mcpServer.AddTool(deleteQueryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			query, err := request.RequireString("query")
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			dsn := request.GetString("dsn", "")
			result, err := HandleExec(cfg, query, StatementTypeDelete, dsn)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(result), nil
		})
	}

	return nil
}

// HandleQuery executes a read query and returns the result as CSV
func HandleQuery(cfg *config.Config, query, expect string, toolDSN string) (string, error) {
	result, headers, err := DoQuery(cfg, query, expect, toolDSN)
	if err != nil {
		return "", err
	}

	s, err := MapToCSV(result, headers)
	if err != nil {
		return "", err
	}

	return s, nil
}

// DoQuery executes a query and returns the result rows and headers
func DoQuery(cfg *config.Config, query, expect string, toolDSN string) ([]map[string]interface{}, []string, error) {
	db, err := GetDB(cfg, toolDSN)
	if err != nil {
		return nil, nil, err
	}

	if len(expect) > 0 && cfg.MySQL.ExplainCheck {
		if err := HandleExplain(cfg, query, expect, toolDSN); err != nil {
			return nil, nil, err
		}
	}

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	result := []map[string]interface{}{}
	for rows.Next() {
		row, err := rows.SliceScan()
		if err != nil {
			return nil, nil, err
		}

		resultRow := map[string]interface{}{}
		for i, col := range cols {
			switch v := row[i].(type) {
			case []byte:
				resultRow[col] = string(v)
			default:
				resultRow[col] = v
			}
		}
		result = append(result, resultRow)
	}

	return result, cols, nil
}

// HandleExec executes a write query and returns the result summary
func HandleExec(cfg *config.Config, query, expect string, toolDSN string) (string, error) {
	db, err := GetDB(cfg, toolDSN)
	if err != nil {
		return "", err
	}

	if len(expect) > 0 && cfg.MySQL.ExplainCheck {
		if err := HandleExplain(cfg, query, expect, toolDSN); err != nil {
			return "", err
		}
	}

	result, err := db.Exec(query)
	if err != nil {
		return "", err
	}

	ra, err := result.RowsAffected()
	if err != nil {
		return "", err
	}

	switch expect {
	case StatementTypeInsert:
		li, err := result.LastInsertId()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d rows affected, last insert id: %d", ra, li), nil
	default:
		return fmt.Sprintf("%d rows affected", ra), nil
	}
}

// HandleExplain checks query plan with EXPLAIN query
func HandleExplain(cfg *config.Config, query, expect string, toolDSN string) error {
	if !cfg.MySQL.ExplainCheck {
		return nil
	}

	db, err := GetDB(cfg, toolDSN)
	if err != nil {
		return err
	}

	rows, err := db.Queryx(fmt.Sprintf("EXPLAIN %s", query))
	if err != nil {
		return err
	}

	result := []ExplainResult{}
	for rows.Next() {
		var row ExplainResult
		if err := rows.StructScan(&row); err != nil {
			return err
		}
		result = append(result, row)
	}

	if len(result) != 1 {
		return fmt.Errorf("unable to check query plan, denied")
	}

	match := false
	switch expect {
	case StatementTypeInsert:
		fallthrough
	case StatementTypeUpdate:
		fallthrough
	case StatementTypeDelete:
		if *result[0].SelectType == expect {
			match = true
		}
	default:
		// for SELECT type query, the select_type will be multiple values
		// here we check if it's not INSERT, UPDATE or DELETE
		match = true
		for _, typ := range []string{StatementTypeInsert, StatementTypeUpdate, StatementTypeDelete} {
			if *result[0].SelectType == typ {
				match = false
				break
			}
		}
	}

	if !match {
		return fmt.Errorf("query plan does not match expected pattern, denied")
	}

	return nil
}

// HandleDescTable describes a table structure
func HandleDescTable(cfg *config.Config, name string, toolDSN string) (string, error) {
	db, err := GetDB(cfg, toolDSN)
	if err != nil {
		return "", err
	}

	rows, err := db.Queryx(fmt.Sprintf("SHOW CREATE TABLE %s", name))
	if err != nil {
		return "", err
	}

	result := []ShowCreateTableResult{}
	for rows.Next() {
		var row ShowCreateTableResult
		if err := rows.StructScan(&row); err != nil {
			return "", err
		}
		result = append(result, row)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("table %s does not exist", name)
	}

	return result[0].CreateTable, nil
}

// MapToCSV converts map result to CSV format
func MapToCSV(m []map[string]interface{}, headers []string) (string, error) {
	var csvBuf strings.Builder
	writer := csv.NewWriter(&csvBuf)

	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %v", err)
	}

	for _, item := range m {
		row := make([]string, len(headers))
		for i, header := range headers {
			value, exists := item[header]
			if !exists {
				return "", fmt.Errorf("key '%s' not found in map", header)
			}
			row[i] = fmt.Sprintf("%v", value)
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write row: %v", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("error flushing CSV writer: %v", err)
	}

	return csvBuf.String(), nil
}
