package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Model is the base interface for all models
type Model interface {
	TableName() string
	PrimaryKey() string
}

// Timestamps adds created_at and updated_at fields
type Timestamps struct {
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// SoftDeletes adds deleted_at field for soft deletes
type SoftDeletes struct {
	DeletedAt *time.Time `db:"deleted_at"`
}

// BaseModel provides common functionality for models
type BaseModel struct {
	db    *DB
	model Model
}

// NewBaseModel creates a new BaseModel instance
func NewBaseModel(db *DB, model Model) *BaseModel {
	return &BaseModel{
		db:    db,
		model: model,
	}
}

// Find finds a record by primary key
func (m *BaseModel) Find(id interface{}, dest Model) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ? LIMIT 1", 
		dest.TableName(), dest.PrimaryKey())
	
	row := m.db.QueryRow(query, id)
	return scanStruct(row, dest)
}

// Create inserts a new record
func (m *BaseModel) Create(model Model) error {
	fields, values := getFieldsAndValues(model, true)
	
	// Handle timestamps
	if hasTimestamps(model) {
		now := time.Now()
		fields = append(fields, "created_at", "updated_at")
		values = append(values, now, now)
	}
	
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		model.TableName(),
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))
	
	result, err := m.db.Exec(query, values...)
	if err != nil {
		return err
	}
	
	// Set auto-increment ID if applicable
	if id, err := result.LastInsertId(); err == nil && id > 0 {
		setFieldValue(model, model.PrimaryKey(), id)
	}
	
	return nil
}

// Update updates a record
func (m *BaseModel) Update(model Model) error {
	fields, values := getFieldsAndValues(model, false)
	
	// Handle timestamps
	if hasTimestamps(model) {
		fields = append(fields, "updated_at")
		values = append(values, time.Now())
	}
	
	// Build SET clause
	setClauses := make([]string, len(fields))
	for i, field := range fields {
		setClauses[i] = field + " = ?"
	}
	
	// Add primary key to values
	pkValue := getFieldValue(model, model.PrimaryKey())
	values = append(values, pkValue)
	
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
		model.TableName(),
		strings.Join(setClauses, ", "),
		model.PrimaryKey())
	
	_, err := m.db.Exec(query, values...)
	return err
}

// Delete deletes a record
func (m *BaseModel) Delete(model Model) error {
	// Check for soft deletes
	if hasSoftDeletes(model) {
		now := time.Now()
		query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE %s = ?",
			model.TableName(), model.PrimaryKey())
		
		pkValue := getFieldValue(model, model.PrimaryKey())
		_, err := m.db.Exec(query, now, pkValue)
		return err
	}
	
	// Hard delete
	pkValue := getFieldValue(model, model.PrimaryKey())
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", 
		model.TableName(), model.PrimaryKey())
	
	_, err := m.db.Exec(query, pkValue)
	return err
}

// Where starts a new query
func (m *BaseModel) Where(field string, value interface{}) *Query {
	q := &Query{
		db:    m.db,
		table: m.model.TableName(),
		wheres: []whereClause{
			{field: field, operator: "=", value: value},
		},
	}
	return q
}

// All retrieves all records
func (m *BaseModel) All(dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s", m.model.TableName())
	
	// Handle soft deletes
	if hasSoftDeletes(m.model) {
		query += " WHERE deleted_at IS NULL"
	}
	
	rows, err := m.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	return scanStructs(rows, dest)
}

// Helper functions

func hasTimestamps(model interface{}) bool {
	_, ok := reflect.TypeOf(model).Elem().FieldByName("Timestamps")
	return ok
}

func hasSoftDeletes(model interface{}) bool {
	_, ok := reflect.TypeOf(model).Elem().FieldByName("SoftDeletes")
	return ok
}

func getFieldsAndValues(model interface{}, excludePK bool) ([]string, []interface{}) {
	var fields []string
	var values []interface{}
	
	v := reflect.ValueOf(model).Elem()
	t := v.Type()
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		
		if tag == "" || tag == "-" {
			continue
		}
		
		// Skip primary key if requested
		if excludePK && tag == model.(Model).PrimaryKey() {
			continue
		}
		
		// Skip timestamps and soft deletes fields
		if tag == "created_at" || tag == "updated_at" || tag == "deleted_at" {
			continue
		}
		
		fields = append(fields, tag)
		values = append(values, v.Field(i).Interface())
	}
	
	return fields, values
}

func getFieldValue(model interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(model).Elem()
	t := v.Type()
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		
		if tag == fieldName {
			return v.Field(i).Interface()
		}
	}
	
	return nil
}

func setFieldValue(model interface{}, fieldName string, value interface{}) {
	v := reflect.ValueOf(model).Elem()
	t := v.Type()
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		
		if tag == fieldName {
			v.Field(i).Set(reflect.ValueOf(value))
			return
		}
	}
}

func scanStruct(row *sql.Row, dest interface{}) error {
	// This is a simplified version - in production you'd use a library like sqlx
	// For now, we'll just return an error
	return fmt.Errorf("scanStruct not fully implemented - use sqlx for production")
}

func scanStructs(rows *sql.Rows, dest interface{}) error {
	// This is a simplified version - in production you'd use a library like sqlx
	// For now, we'll just return an error
	return fmt.Errorf("scanStructs not fully implemented - use sqlx for production")
}