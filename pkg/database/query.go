package database

import (
	"database/sql"
	"fmt"
	"strings"
)

// Query represents a database query builder
type Query struct {
	db       *DB
	table    string
	selects  []string
	wheres   []whereClause
	joins    []joinClause
	orderBys []orderByClause
	groupBys []string
	limit    int
	offset   int
}

type whereClause struct {
	field    string
	operator string
	value    interface{}
	boolean  string // AND or OR
}

type joinClause struct {
	joinType string // INNER, LEFT, RIGHT
	table    string
	first    string
	operator string
	second   string
}

type orderByClause struct {
	field     string
	direction string // ASC or DESC
}

// NewQuery creates a new query builder
func NewQuery(db *DB) *Query {
	return &Query{
		db:       db,
		selects:  []string{"*"},
		wheres:   []whereClause{},
		joins:    []joinClause{},
		orderBys: []orderByClause{},
		groupBys: []string{},
	}
}

// Table sets the table for the query
func (q *Query) Table(table string) *Query {
	q.table = table
	return q
}

// Select sets the columns to select
func (q *Query) Select(columns ...string) *Query {
	q.selects = columns
	return q
}

// Where adds a where condition
func (q *Query) Where(field string, operatorOrValue interface{}, value ...interface{}) *Query {
	var op string
	var val interface{}
	
	if len(value) > 0 {
		op = operatorOrValue.(string)
		val = value[0]
	} else {
		op = "="
		val = operatorOrValue
	}
	
	q.wheres = append(q.wheres, whereClause{
		field:    field,
		operator: op,
		value:    val,
		boolean:  "AND",
	})
	
	return q
}

// OrWhere adds an OR where condition
func (q *Query) OrWhere(field string, operatorOrValue interface{}, value ...interface{}) *Query {
	var op string
	var val interface{}
	
	if len(value) > 0 {
		op = operatorOrValue.(string)
		val = value[0]
	} else {
		op = "="
		val = operatorOrValue
	}
	
	q.wheres = append(q.wheres, whereClause{
		field:    field,
		operator: op,
		value:    val,
		boolean:  "OR",
	})
	
	return q
}

// WhereIn adds a WHERE IN condition
func (q *Query) WhereIn(field string, values []interface{}) *Query {
	q.wheres = append(q.wheres, whereClause{
		field:    field,
		operator: "IN",
		value:    values,
		boolean:  "AND",
	})
	
	return q
}

// WhereNull adds a WHERE IS NULL condition
func (q *Query) WhereNull(field string) *Query {
	q.wheres = append(q.wheres, whereClause{
		field:    field,
		operator: "IS NULL",
		value:    nil,
		boolean:  "AND",
	})
	
	return q
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (q *Query) WhereNotNull(field string) *Query {
	q.wheres = append(q.wheres, whereClause{
		field:    field,
		operator: "IS NOT NULL",
		value:    nil,
		boolean:  "AND",
	})
	
	return q
}

// Join adds an inner join
func (q *Query) Join(table, first, operator, second string) *Query {
	q.joins = append(q.joins, joinClause{
		joinType: "INNER",
		table:    table,
		first:    first,
		operator: operator,
		second:   second,
	})
	
	return q
}

// LeftJoin adds a left join
func (q *Query) LeftJoin(table, first, operator, second string) *Query {
	q.joins = append(q.joins, joinClause{
		joinType: "LEFT",
		table:    table,
		first:    first,
		operator: operator,
		second:   second,
	})
	
	return q
}

// OrderBy adds an order by clause
func (q *Query) OrderBy(field string, direction ...string) *Query {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	
	q.orderBys = append(q.orderBys, orderByClause{
		field:     field,
		direction: dir,
	})
	
	return q
}

// GroupBy adds a group by clause
func (q *Query) GroupBy(fields ...string) *Query {
	q.groupBys = append(q.groupBys, fields...)
	return q
}

// Limit sets the limit
func (q *Query) Limit(limit int) *Query {
	q.limit = limit
	return q
}

// Offset sets the offset
func (q *Query) Offset(offset int) *Query {
	q.offset = offset
	return q
}

// Get executes the query and returns all results
func (q *Query) Get() (*sql.Rows, error) {
	query, args := q.toSQL()
	return q.db.Query(query, args...)
}

// First executes the query and returns the first result
func (q *Query) First() (*sql.Row, error) {
	q.limit = 1
	query, args := q.toSQL()
	return q.db.QueryRow(query, args...), nil
}

// Count returns the count of matching records
func (q *Query) Count() (int64, error) {
	// Save current selects
	oldSelects := q.selects
	q.selects = []string{"COUNT(*) as count"}
	
	query, args := q.toSQL()
	
	// Restore selects
	q.selects = oldSelects
	
	var count int64
	err := q.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// Exists checks if any records exist
func (q *Query) Exists() (bool, error) {
	count, err := q.Count()
	return count > 0, err
}

// Update updates records
func (q *Query) Update(updates map[string]interface{}) (sql.Result, error) {
	if len(updates) == 0 {
		return nil, fmt.Errorf("no updates provided")
	}
	
	var setClauses []string
	var args []interface{}
	
	for field, value := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field))
		args = append(args, value)
	}
	
	query := fmt.Sprintf("UPDATE %s SET %s", q.table, strings.Join(setClauses, ", "))
	
	// Add where clauses
	whereSQL, whereArgs := q.buildWhere()
	if whereSQL != "" {
		query += " WHERE " + whereSQL
		args = append(args, whereArgs...)
	}
	
	return q.db.Exec(query, args...)
}

// Delete deletes records
func (q *Query) Delete() (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM %s", q.table)
	
	// Add where clauses
	whereSQL, args := q.buildWhere()
	if whereSQL != "" {
		query += " WHERE " + whereSQL
	}
	
	return q.db.Exec(query, args...)
}

// toSQL builds the SQL query
func (q *Query) toSQL() (string, []interface{}) {
	var args []interface{}
	
	// SELECT
	query := fmt.Sprintf("SELECT %s FROM %s", strings.Join(q.selects, ", "), q.table)
	
	// JOIN
	for _, join := range q.joins {
		query += fmt.Sprintf(" %s JOIN %s ON %s %s %s",
			join.joinType, join.table, join.first, join.operator, join.second)
	}
	
	// WHERE
	whereSQL, whereArgs := q.buildWhere()
	if whereSQL != "" {
		query += " WHERE " + whereSQL
		args = append(args, whereArgs...)
	}
	
	// GROUP BY
	if len(q.groupBys) > 0 {
		query += " GROUP BY " + strings.Join(q.groupBys, ", ")
	}
	
	// ORDER BY
	if len(q.orderBys) > 0 {
		var orderClauses []string
		for _, order := range q.orderBys {
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", order.field, order.direction))
		}
		query += " ORDER BY " + strings.Join(orderClauses, ", ")
	}
	
	// LIMIT
	if q.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.limit)
	}
	
	// OFFSET
	if q.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.offset)
	}
	
	return query, args
}

// buildWhere builds the WHERE clause
func (q *Query) buildWhere() (string, []interface{}) {
	if len(q.wheres) == 0 {
		return "", nil
	}
	
	var clauses []string
	var args []interface{}
	
	for i, where := range q.wheres {
		clause := ""
		
		// Add boolean operator for subsequent clauses
		if i > 0 {
			clause += where.boolean + " "
		}
		
		switch where.operator {
		case "IN":
			values := where.value.([]interface{})
			placeholders := make([]string, len(values))
			for j := range placeholders {
				placeholders[j] = "?"
				args = append(args, values[j])
			}
			clause += fmt.Sprintf("%s IN (%s)", where.field, strings.Join(placeholders, ", "))
			
		case "IS NULL", "IS NOT NULL":
			clause += fmt.Sprintf("%s %s", where.field, where.operator)
			
		default:
			clause += fmt.Sprintf("%s %s ?", where.field, where.operator)
			args = append(args, where.value)
		}
		
		clauses = append(clauses, clause)
	}
	
	return strings.Join(clauses, " "), args
}