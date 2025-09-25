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
	columns := []string{}
	placeholders := []string{}
	updateAssignments := []string{}

	entityValue := reflect.ValueOf(entity)
	entityType := reflect.TypeOf(entity)

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		valueField := entityValue.Field(i)
		if (dbTag == "created_at" || dbTag == "updated_at") && valueField.IsZero() {
			continue
		}

		columns = append(columns, dbTag)
		placeholders = append(placeholders, ":"+dbTag)

		if dbTag != repository.keyColumn && dbTag != "updated_at" {
			updateAssignments = append(updateAssignments, fmt.Sprintf("%s = VALUES(%s)", dbTag, dbTag))
		}
	}

	columnsStr := strings.Join(columns, ", ")
	valuesStr := strings.Join(placeholders, ", ")

	var updateClause string
	if len(updateAssignments) > 0 {
		updateClause = strings.Join(updateAssignments, ", ") + ", updated_at = CURRENT_TIMESTAMP"
	} else {
		updateClause = "updated_at = CURRENT_TIMESTAMP"
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON DUPLICATE KEY UPDATE %s`,
		repository.tableName, columnsStr, valuesStr, updateClause)

	fmt.Printf("Entity Values: %+v\n", entity)
	fmt.Printf("Query: %+v\n", query)

	result, err := repository.database.NamedExecContext(ctx, query, entity)

	fmt.Printf("Update Query: %s\n", result)

	return err
}

func (repository *GenericRepository[T]) Update(ctx context.Context, entity T) error {
	setAssignments := []string{}
	entityType := reflect.TypeOf(entity)

	for i := 0; i < entityType.NumField(); i++ {
		field := entityType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		if dbTag == repository.keyColumn || dbTag == "created_at" {
			continue
		}

		setAssignments = append(setAssignments, fmt.Sprintf("%s = :%s", dbTag, dbTag))
	}

	if len(setAssignments) == 0 {
		return nil
	}

	setClause := strings.Join(setAssignments, ", ") + ", updated_at = CURRENT_TIMESTAMP"

	query := fmt.Sprintf(`
		UPDATE %s
		SET %s
		WHERE %s = :%s`,
		repository.tableName,
		setClause,
		repository.keyColumn,
		repository.keyColumn,
	)

	result, err := repository.database.NamedExecContext(ctx, query, entity)

	fmt.Printf("Entity Values: %+v\n", entity)
	fmt.Printf("Update Query: %s\n", result)

	return err
}

func (repository *GenericRepository[T]) BatchUpsert(ctx context.Context, entitiesToUpsert []T) error {
	if len(entitiesToUpsert) == 0 {
		return nil
	}

	columns := []string{}
	updateAssignments := []string{}

	entityType := reflect.TypeOf(entitiesToUpsert[0])
	for fieldIndex := 0; fieldIndex < entityType.NumField(); fieldIndex++ {
		field := entityType.Field(fieldIndex)
		dbColumnName := field.Tag.Get("db")
		if dbColumnName == "" || dbColumnName == "-" {
			continue
		}

		columns = append(columns, dbColumnName)

		if dbColumnName != repository.keyColumn && dbColumnName != "updated_at" {
			updateAssignments = append(updateAssignments, fmt.Sprintf("%s = VALUES(%s)", dbColumnName, dbColumnName))
		}
	}

	columnsString := strings.Join(columns, ", ")
	updateClause := "updated_at = CURRENT_TIMESTAMP"
	if len(updateAssignments) > 0 {
		updateClause = strings.Join(updateAssignments, ", ") + ", " + updateClause
	}

	valuePlaceholdersList := []string{}
	argumentMaps := []map[string]interface{}{}
	for _, singleEntity := range entitiesToUpsert {
		singleEntityValue := reflect.ValueOf(singleEntity)
		singleEntityMap := map[string]interface{}{}
		placeholders := []string{}

		for fieldIndex := 0; fieldIndex < entityType.NumField(); fieldIndex++ {
			field := entityType.Field(fieldIndex)
			dbColumnName := field.Tag.Get("db")
			if dbColumnName == "" || dbColumnName == "-" {
				continue
			}

			fieldValue := singleEntityValue.Field(fieldIndex).Interface()
			singleEntityMap[dbColumnName] = fieldValue
			placeholders = append(placeholders, ":"+dbColumnName)
		}

		argumentMaps = append(argumentMaps, singleEntityMap)
		valuePlaceholdersList = append(valuePlaceholdersList, "("+strings.Join(placeholders, ", ")+")")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES %s
		ON DUPLICATE KEY UPDATE %s`,
		repository.tableName,
		columnsString,
		strings.Join(valuePlaceholdersList, ", "),
		updateClause,
	)

	fmt.Printf("Entities being upserted")

	for _, argumentMap := range argumentMaps {
		if _, err := repository.database.NamedExecContext(ctx, query, argumentMap); err != nil {
			return fmt.Errorf("failed to upsert entity: %w", err)
		}
	}

	return nil
}

func (repository *GenericRepository[T]) GetByField(ctx context.Context, fieldName, fieldValue string) (*T, error) {
	var entity T
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = ? LIMIT 1",
		strings.Join(repository.columnList, ", "),
		repository.tableName,
		fieldName,
	)

	if err := repository.database.GetContext(ctx, &entity, query, fieldValue); err != nil {
		return nil, err
	}

	fmt.Printf("Query: %+v\n", query)

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
		"SELECT %s FROM %s ORDER BY %s %s LIMIT ? OFFSET ?",
		strings.Join(repository.columnList, ", "),
		repository.tableName,
		sortColumn,
		sortDirection,
	)

	if err := repository.database.SelectContext(ctx, &results, query, limit, offset); err != nil {
		return nil, err
	}

	fmt.Printf("Query: %+v\n", query)

	return results, nil
}

func (repository *GenericRepository[T]) DeleteByField(ctx context.Context, fieldName, fieldValue string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", repository.tableName, fieldName)
	_, err := repository.database.ExecContext(ctx, query, fieldValue)

	fmt.Printf("Query: %+v\n", query)

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
