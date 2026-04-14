package repository

import "github.com/KaiserWerk/sqlite-mcp-server/internal/models"

type Repository interface {
	GetSchema() ([]models.Table, error)
	Query(sqlQuery string) (*models.QueryResult, error)
	Execute(sqlQuery string) (*models.ExecuteResult, error)
	Close() error
}
