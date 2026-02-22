package dialect

import "reflect"

var dialectMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string
	TableExistSQL(tableName string) (string, []interface{})
}

func GetDialect(dialect string) Dialect {
	return dialectMap[dialect]
}
func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}
