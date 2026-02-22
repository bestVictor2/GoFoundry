package clause

import "strings"

type Clause struct {
	sql     map[Type]string
	sqlVals map[Type][]interface{}
}
type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

func (c *Clause) Set(name Type, value ...interface{}) *Clause {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVals = make(map[Type][]interface{})
	}
	sql, val := generators[name](value...)
	c.sql[name] = sql
	c.sqlVals[name] = val
	return c
}
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vals []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vals = append(vals, c.sqlVals[order]...)
		}
	}
	return strings.Join(sqls, " "), vals
}
