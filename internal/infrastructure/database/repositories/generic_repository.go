package repositories

import (
    "context"
    "fmt"
    "reflect"
    "strings"

    "github.com/jmoiron/sqlx"
)

type GenericRepository[T any] struct {
	db        *sqlx.DB
	tableName string
	columns   []string
	keyColumn string
}

func NewGenericRepository[T any](db *sqlx.DB, tableName, keyColumn string) *GenericRepository[T] {
	var zero T
	columns := extractColumns(zero)

	return &GenericRepository[T]{
		db:        db,
		tableName: tableName,
		columns:   columns,
		keyColumn: keyColumn,
	}
}

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

func (r *GenericRepository[T]) Upsert(ctx context.Context, entity T) error {
	columnsStr := strings.Join(r.columns, ", ")
	placeholders := make([]string, len(r.columns))
	for i, col := range r.columns {
		placeholders[i] = ":" + col
	}
	valuesStr := strings.Join(placeholders, ", ")

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

func (r *GenericRepository[T]) GetByField(ctx context.Context, field, value string) (*T, error) {
	var entity T
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 LIMIT 1",
		strings.Join(r.columns, ", "), r.tableName, field)

	if err := r.db.GetContext(ctx, &entity, query, value); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GenericRepository[T]) List(ctx context.Context, limit, page int, orderBy, orderDirection string) ([]T, error) {
    var results []T

    resolvedOrderBy := orderBy
    if resolvedOrderBy == "" || !r.isValidColumn(orderBy) {
        resolvedOrderBy = r.keyColumn
    }

    resolvedDirection := strings.ToUpper(orderDirection)
    if resolvedDirection != "DESC" {
        resolvedDirection = "ASC"
    }

    if limit <= 0 {
        limit = 10
    }
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * limit

    query := fmt.Sprintf(
        "SELECT %s FROM %s ORDER BY %s %s LIMIT $1 OFFSET $2",
        strings.Join(r.columns, ", "),
        r.tableName,
        resolvedOrderBy,
        resolvedDirection,
    )

    if err := r.db.SelectContext(ctx, &results, query, limit, offset); err != nil {
        return nil, err
    }
    return results, nil
}

func (r *GenericRepository[T]) DeleteByField(ctx context.Context, field, value string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", r.tableName, field)
	_, err := r.db.ExecContext(ctx, query, value)
	return err
}

func (r *GenericRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	return r.GetByField(ctx, r.keyColumn, fmt.Sprintf("%v", id))
}

func (r *GenericRepository[T]) DeleteByID(ctx context.Context, id interface{}) error {
	return r.DeleteByField(ctx, r.keyColumn, fmt.Sprintf("%v", id))
}

func (r *GenericRepository[T]) isValidColumn(column string) bool {
    for _, c := range r.columns {
        if c == column {
            return true
        }
    }
    return false
}
