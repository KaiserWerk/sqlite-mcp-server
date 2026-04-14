package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/KaiserWerk/sqlite-mcp-server/internal/config"
	"github.com/KaiserWerk/sqlite-mcp-server/internal/handlers"
	"github.com/KaiserWerk/sqlite-mcp-server/internal/repository"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	dbPath string
	debug  bool
)

func main() {
	flag.StringVar(&dbPath, "database", "", "The path to the SQLite file to use")
	flag.BoolVar(&debug, "debug", false, "Whether to set the log level to 'Debug'")
	flag.Parse()

	if dbPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --database flag is required")
		os.Exit(1)
	}

	cfg := &config.Config{
		DatabasePath: dbPath,
		Debug:        debug,
	}

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	fh, err := os.Create("server.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log file: %v\n", err)
		os.Exit(1)
	}
	defer fh.Close()

	logger := slog.New(slog.NewTextHandler(fh, &slog.HandlerOptions{
		Level: level,
	}))

	err = runServer(context.Background(), cfg, logger)
	_ = fh.Close()
	if err != nil {
		logger.Error("Server error: " + err.Error())
		os.Exit(1)
	}
}

func runServer(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Initialize database
	repo, err := repository.NewSQLiteDB(cfg.DatabasePath, logger)
	if err != nil {
		logger.Error("Failed to initialize database: " + err.Error())
		return err
	}
	defer repo.Close()

	// Initialize MCP handler
	mcpHandler := handlers.NewMCPHandler(repo, logger)

	mcpServer := server.NewMCPServer(
		"sqlite-mcp-server",
		"-",
	)

	// Get Schema Tool - No parameters needed
	listTablesTool := mcp.NewTool("sqlite_get_schema",
		mcp.WithDescription("List all tables in the SQLite database with their schema information including columns, types, constraints, and indexes"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	)
	mcpServer.AddTool(listTablesTool, mcpHandler.GetSchema)

	// Query Database Tool
	queryDatabaseTool := mcp.NewTool("sqlite_query",
		mcp.WithDescription("Execute SELECT queries against the SQLite database. Only SELECT, WITH, and EXPLAIN queries are allowed."),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL SELECT query to execute"),
			mcp.MinLength(1),
			mcp.MaxLength(10000),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(true),
	)
	mcpServer.AddTool(queryDatabaseTool, mcpHandler.Query)

	// Execute Database Tool
	executeDatabaseTool := mcp.NewTool("sqlite_execute",
		mcp.WithDescription("Execute DDL/DML operations (INSERT, UPDATE) against the SQLite database. SELECT queries are not allowed - use sqlite_query instead."),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL statement to execute (non-SELECT operations only)"),
			mcp.MinLength(1),
			mcp.MaxLength(10000),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	)
	mcpServer.AddTool(executeDatabaseTool, mcpHandler.Execute)

	// Execute administrative tool
	executeAdminDatabaseTool := mcp.NewTool("sqlite_execute_admin",
		mcp.WithDescription("Execute DDL/DML operations (DELETE, CREATE, ALTER, DROP, etc.) against the SQLite database. SELECT queries are not allowed - use sqlite_query instead."),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL statement to execute (non-SELECT operations only) administrative or destructive operations that may modify or delete data, alter schema, or perform other high-impact actions"),
			mcp.MinLength(1),
			mcp.MaxLength(10000),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	)
	mcpServer.AddTool(executeAdminDatabaseTool, mcpHandler.ExecuteAdmin)

	//Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Start server
	// TODO: Look into alternatives for transport layer (sse, streamable-http)
	logger.Info("SQLite MCP Server started successfully")
	if err := server.ServeStdio(mcpServer); err != nil {
		logger.Error("Server error: " + err.Error())
	}

	logger.Info("SQLite MCP Server stopped")

	return nil
}
