package session

import (
	"GoGorm/log"
	"GoGorm/schema"
	"fmt"
	"reflect"
	"strings"
)

func (s *Session) Model(model interface{}) *Session {
	if s.refTable == nil || reflect.TypeOf(model) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schema.Parse(model, s.dialect)
	}
	return s
}
func (s *Session) RefTable() *schema.Schema {
	if s.refTable == nil {
		log.Error("refTable is nil")
		return nil
	} else {
		return s.refTable
	}
}
func (s *Session) CreateTable() error {
	table := s.RefTable()
	var columns []string
	for _, feilds := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", feilds.Name, feilds.Type, feilds.Tag))
	}
	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}
func (s *Session) DropTable() error {
	_, err := s.Raw("DROP TABLE IF EXISTS " + s.refTable.Name).Exec()
	return err
}
func (s *Session) HasTable() bool {
	sql, values := s.dialect.TableExistSQL(s.refTable.Name)
	row := s.Raw(sql, values...).QueryRow()
	var tmp string
	_ = row.Scan(&tmp)
	return tmp == s.refTable.Name
}

func (s *Session) AutoMigrate() error {
	if s.HasTable() {
		return nil
	}
	return s.CreateTable()
}
