package repositories

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

// GenericRepository provides generic CRUD operations for any table/struct
type GenericRepository[T any] struct {
	db        *sqlx.DB
	tableName string
	columns   []string
	keyColumn string
}

// NewGenericRepository creates a new generic repository
func NewGenericRepository[T any](db *sqlx.DB, tableName, keyColumn string) *GenericRepository[T] {
	// Extract column names from struct tags
	var zero T
	columns := extractColumns(zero)

	return &GenericRepository[T]{
		db:        db,
		tableName: tableName,
		columns:   columns,
		keyColumn: keyColumn,
	}
}

// extractColumns extracts column names from struct db tags
func extractColumns[T any](zero T) []string {
	t := reflect.TypeOf(zero)
	var columns []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if dbTag := field.Tag.Get("db"); dbTag != "" && dbTag != "-" {
			columns = append(columns, dbTag)
		}
	}
	return columns
}

// Upsert performs INSERT ... ON CONFLICT DO UPDATE
func (r *GenericRepository[T]) Upsert(ctx context.Context, entity T) error {
	columnsStr := strings.Join(r.columns, ", ")
	placeholders := make([]string, len(r.columns))
	for i, col := range r.columns {
		placeholders[i] = ":" + col
	}
	valuesStr := strings.Join(placeholders, ", ")

	// Build SET clause for ON CONFLICT
	setClause := make([]string, 0, len(r.columns))
	for _, col := range r.columns {
		if col != r.keyColumn {
			setClause = append(setClause, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
	}
	setStr := strings.Join(setClause, ", ")

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (%s) DO UPDATE SET %s, updated_at = NOW()`,
		r.tableName, columnsStr, valuesStr, r.keyColumn, setStr)

	_, err := r.db.NamedExecContext(ctx, query, entity)
	return err
}

// GetByField retrieves a single record by a specific field
func (r *GenericRepository[T]) GetByField(ctx context.Context, field, value string) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 LIMIT 1",
		strings.Join(r.columns, ", "), r.tableName, field)

	if err := r.db.GetContext(ctx, &entity, query, value); err != nil {
		return nil, err
	}
	return &entity, nil
}

// List retrieves all records
func (r *GenericRepository[T]) List(ctx context.Context) ([]T, error) {
	var entities []T
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(r.columns, ", "), r.tableName)

	if err := r.db.SelectContext(ctx, &entities, query); err != nil {
		return nil, err
	}
	return entities, nil
}

// DeleteByField deletes records by a specific field
func (r *GenericRepository[T]) DeleteByField(ctx context.Context, field, value string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", r.tableName, field)
	_, err := r.db.ExecContext(ctx, query, value)
	return err
}

// GetByID retrieves a record by its primary key
func (r *GenericRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	return r.GetByField(ctx, r.keyColumn, fmt.Sprintf("%v", id))
}

// DeleteByID deletes a record by its primary key
func (r *GenericRepository[T]) DeleteByID(ctx context.Context, id interface{}) error {
	return r.DeleteByField(ctx, r.keyColumn, fmt.Sprintf("%v", id))
}
