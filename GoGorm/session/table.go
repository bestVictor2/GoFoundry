package session

import (
	"GoGorm/log"
	"GoGorm/schema"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
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

func (s *Session) Columns() ([]string, error) {
	rows, err := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1;", s.refTable.Name)).QueryRows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return rows.Columns()
}

func (s *Session) columnTypes() (map[string]string, error) {
	rows, err := s.Raw(fmt.Sprintf("PRAGMA table_info(%s);", s.refTable.Name)).QueryRows()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	types := make(map[string]string)
	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err = rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		types[name] = strings.ToLower(typ)
	}
	return types, rows.Err()
}

func (s *Session) AutoMigrate() (err error) {
	if !s.HasTable() {
		return s.CreateTable()
	}

	oldColumns, err := s.Columns()
	if err != nil {
		return err
	}
	oldTypes, err := s.columnTypes()
	if err != nil {
		return err
	}
	newColumns := s.RefTable().FieldNames
	addCols, delCols := diffColumns(newColumns, oldColumns)
	typeChanged := hasTypeChange(s.RefTable(), oldTypes)
	if len(addCols) == 0 && len(delCols) == 0 && !typeChanged {
		return nil
	}

	managedTx := false
	if s.tx == nil {
		if err = s.Begin(); err != nil {
			return err
		}
		managedTx = true
	}
	defer func() {
		if managedTx && err != nil {
			_ = s.Rollback()
		}
	}()

	tableName := s.RefTable().Name
	tmpTable := fmt.Sprintf("%s_tmp_%d", tableName, time.Now().UnixNano())
	if _, err = s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tableName, tmpTable)).Exec(); err != nil {
		return err
	}
	if err = s.CreateTable(); err != nil {
		return err
	}

	common := intersectColumns(newColumns, oldColumns)
	if len(common) > 0 {
		oldSet := make(map[string]struct{}, len(oldColumns))
		for _, col := range oldColumns {
			oldSet[col] = struct{}{}
		}
		var selectParts []string
		for _, col := range newColumns {
			if _, ok := oldSet[col]; ok {
				selectParts = append(selectParts, col)
				continue
			}
			field := s.RefTable().GetField(col)
			selectParts = append(selectParts, defaultValueExpr(field.Type))
		}
		columns := strings.Join(newColumns, ",")
		selectExpr := strings.Join(selectParts, ",")
		if _, err = s.Raw(fmt.Sprintf("INSERT INTO %s(%s) SELECT %s FROM %s;", tableName, columns, selectExpr, tmpTable)).Exec(); err != nil {
			return err
		}
	}
	if _, err = s.Raw(fmt.Sprintf("DROP TABLE %s;", tmpTable)).Exec(); err != nil {
		return err
	}
	if managedTx {
		return s.Commit()
	}
	return nil
}

func diffColumns(target []string, source []string) (add []string, remove []string) {
	targetSet := make(map[string]struct{}, len(target))
	sourceSet := make(map[string]struct{}, len(source))
	for _, col := range target {
		targetSet[col] = struct{}{}
	}
	for _, col := range source {
		sourceSet[col] = struct{}{}
	}
	for _, col := range target {
		if _, ok := sourceSet[col]; !ok {
			add = append(add, col)
		}
	}
	for _, col := range source {
		if _, ok := targetSet[col]; !ok {
			remove = append(remove, col)
		}
	}
	return add, remove
}

func intersectColumns(a []string, b []string) []string {
	bSet := make(map[string]struct{}, len(b))
	for _, col := range b {
		bSet[col] = struct{}{}
	}
	var res []string
	for _, col := range a {
		if _, ok := bSet[col]; ok {
			res = append(res, col)
		}
	}
	return res
}

func defaultValueExpr(typ string) string {
	switch strings.ToLower(typ) {
	case "integer", "bigint", "real", "bool":
		return "0"
	case "blob":
		return "x''"
	case "text", "datetime":
		return "''"
	default:
		return "NULL"
	}
}

func hasTypeChange(table *schema.Schema, oldTypes map[string]string) bool {
	for _, field := range table.Fields {
		oldType, ok := oldTypes[field.Name]
		if !ok {
			continue
		}
		if oldType != strings.ToLower(field.Type) {
			return true
		}
	}
	return false
}
