package repositories

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

type GenericRepository[T any] struct {
	database   *sqlx.DB
	tableName  string
	columnList []string
	keyColumn  string
}

func NewGenericRepository[T any](database *sqlx.DB, tableName, keyColumn string) *GenericRepository[T] {
	var zeroValue T
	columnList := extractColumnNames(zeroValue)

	return &GenericRepository[T]{
		database:   database,
		tableName:  tableName,
		columnList: columnList,
		keyColumn:  keyColumn,
	}
}

func extractColumnNames[T any](zeroValue T) []string {
	entityType := reflect.TypeOf(zeroValue)
	var columnNames []string

	for index := 0; index < entityType.NumField(); index++ {
		field := entityType.Field(index)
		if dbTag := field.Tag.Get("db"); dbTag != "" && dbTag != "-" {
			columnNames = append(columnNames, dbTag)
		}
	}
	return columnNames
}

func (repository *GenericRepository[T]) Upsert(ctx context.Context, entity T) error {
	columns := strings.Join(repository.columnList, ", ")
	placeholders := make([]string, len(repository.columnList))
	for index, column := range repository.columnList {
		placeholders[index] = ":" + column
	}
	values := strings.Join(placeholders, ", ")

	updateAssignments := make([]string, 0, len(repository.columnList))
	for _, column := range repository.columnList {
		if column != repository.keyColumn {
			updateAssignments = append(updateAssignments, fmt.Sprintf("%s = EXCLUDED.%s", column, column))
		}
	}
	updateClause := strings.Join(updateAssignments, ", ")

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (%s) DO UPDATE SET %s, updated_at = NOW()`,
		repository.tableName, columns, values, repository.keyColumn, updateClause)

	_, err := repository.database.NamedExecContext(ctx, query, entity)
	return err
}

func (repository *GenericRepository[T]) GetByField(ctx context.Context, fieldName, fieldValue string) (*T, error) {
	var entity T
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = $1 LIMIT 1",
		strings.Join(repository.columnList, ", "),
		repository.tableName,
		fieldName,
	)

	if err := repository.database.GetContext(ctx, &entity, query, fieldValue); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (repository *GenericRepository[T]) List(
	ctx context.Context,
	limit int,
	page int,
	orderBy string,
	orderDirection string,
) ([]T, error) {
	var results []T

	sortColumn := orderBy
	if sortColumn == "" || !repository.isValidColumn(orderBy) {
		sortColumn = repository.keyColumn
	}

	sortDirection := strings.ToUpper(orderDirection)
	if sortDirection != "DESC" {
		sortDirection = "ASC"
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
		strings.Join(repository.columnList, ", "),
		repository.tableName,
		sortColumn,
		sortDirection,
	)

	if err := repository.database.SelectContext(ctx, &results, query, limit, offset); err != nil {
		return nil, err
	}
	return results, nil
}

func (repository *GenericRepository[T]) DeleteByField(ctx context.Context, fieldName, fieldValue string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", repository.tableName, fieldName)
	_, err := repository.database.ExecContext(ctx, query, fieldValue)
	return err
}

func (repository *GenericRepository[T]) GetByID(ctx context.Context, identifier interface{}) (*T, error) {
	return repository.GetByField(ctx, repository.keyColumn, fmt.Sprintf("%v", identifier))
}

func (repository *GenericRepository[T]) DeleteByID(ctx context.Context, identifier interface{}) error {
	return repository.DeleteByField(ctx, repository.keyColumn, fmt.Sprintf("%v", identifier))
}

func (repository *GenericRepository[T]) isValidColumn(columnName string) bool {
	for _, column := range repository.columnList {
		if column == columnName {
			return true
		}
	}
	return false
}
