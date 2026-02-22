package session

import (
	"GoGorm/clause"
	"fmt"
	"reflect"
)

func (s *Session) Insert(values ...interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, nil
	}
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		if err := callBeforeInsert(value, s); err != nil {
			return 0, err
		}
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordsValues(value))
	}
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	for _, value := range values {
		if err = callAfterInsert(value, s); err != nil {
			return 0, err
		}
	}
	return result.RowsAffected()
}
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	modelValue := reflect.New(destType).Interface()
	if err := callBeforeQuery(modelValue, s); err != nil {
		return err
	}
	fields := table.FieldNames
	if len(s.selects) > 0 {
		fields = s.selects
	}
	for _, field := range fields {
		if table.GetField(field) == nil {
			return fmt.Errorf("unknown field %s", field)
		}
	}
	s.clause.Set(clause.SELECT, table.Name, fields)
	sql, vals := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vals...).QueryRows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, value := range fields {
			values = append(values, dest.FieldByName(value).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		if err := callAfterQuery(dest.Addr().Interface(), s); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return nil
}
func (s *Session) Update(values ...interface{}) (int64, error) {
	m, ok := values[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			m[values[i].(string)] = values[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.refTable.Name, m)
	sql, vals := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vals...).Exec()
	if err != nil {
		//fmt.Print(111)
		return 0, err
	}
	//fmt.Println(result.RowsAffected())
	return result.RowsAffected()
}
func (s *Session) Delete() (int64, error) {
	s.clause.Set(clause.DELETE, s.refTable.Name)
	sql, vals := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vals...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.refTable.Name)
	sql, vals := s.clause.Build(clause.COUNT, clause.WHERE)
	result := s.Raw(sql, vals...).QueryRow()
	var tmp int64
	if err := result.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}
func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}
func (s *Session) First(dest interface{}) error {
	destval := reflect.Indirect(reflect.ValueOf(dest))
	destSlice := reflect.New(reflect.SliceOf(destval.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return ErrRecordNotFound
	}
	destval.Set(destSlice.Index(0))
	return nil
}
