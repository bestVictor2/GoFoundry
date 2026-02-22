package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}
func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}
func _insert(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	args := strings.Join(values[1].([]string), ", ")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, args), []interface{}{}
}

func _values(values ...interface{}) (string, []interface{}) {
	var bindstr string
	var args []interface{}
	var sql strings.Builder
	sql.WriteString("VALUES ")
	for i, value := range values {
		v := value.([]interface{})
		if bindstr == "" {
			bindstr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindstr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		args = append(args, v...)
	}
	return sql.String(), args
}
func _select(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}
func _limit(values ...interface{}) (string, []interface{}) {
	return "LIMIT ?", values
}
func _where(values ...interface{}) (string, []interface{}) {
	desc, values := values[0], values[1:]
	return fmt.Sprintf("WHERE %v", desc), values
}
func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %v", values[0]), []interface{}{}
}
func _update(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var args []interface{}
	for k, v := range m {
		keys = append(keys, k+" = ?")
		args = append(args, v)
	}
	return fmt.Sprintf("UPDATE %v SET %v", tableName, strings.Join(keys, ", ")), args
}
func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}
func _count(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"count(*)"})
}
