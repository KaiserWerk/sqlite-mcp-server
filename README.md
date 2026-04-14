# SQLite MCP Server

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server for SQLite database operations. This server provides a standardized interface for SQLite database interactions including schema inspection, read, write and administrative operations.

## Available Tools

The MCP server exposes four tools:

#### get_schema

- Description: List all tables in the SQLite database with their schema information
- Parameters: None
- Usage: Provides complete schema introspection including columns, types, constraints, and indexes

#### query

- Description: Execute read-only queries against the SQLite database.
- Parameters: 
  - `sql` (required): Read-only SQL query to execute
-Usage: Only SELECT, WITH, and EXPLAIN queries are allowed
- Example: `SELECT * FROM users WHERE age > 25`

#### execute

- Description: Execute write operations against the SQLite database
- Parameters:
  - `sql` (required): SQL statement that modifies the database
- Usage: INSERT and UPDATE operations
- Example: `INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')`

#### execute_admin

- Description: Execute administrative or destructive operations against the SQLite database
- Parameters:
  - `sql` (required): SQL statement that modifies the database schema, drops data or tables or creates new ones
- Usage: DELETE, CREATE, ALTER, DROP operations
- Example: `DELETE FROM users`

