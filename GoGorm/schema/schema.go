package schema

import (
	"GoGorm/dialect"
	"go/ast"
	"reflect"
)

type Field struct {
	Name string
	Type string
	Tag  string
}
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field
	FieldNames []string
	FieldMap   map[string]*Field
}

func (s *Schema) GetField(name string) *Field {
	return s.FieldMap[name]
}
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	model := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model: dest,
		Name:  model.Name(),
		//Fields:   make([]*Field, 0),
		FieldMap: make(map[string]*Field),
	}
	for i := 0; i < model.NumField(); i++ {
		p := model.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("foundry"); ok {
				field.Tag = v
			} else if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.FieldMap[p.Name] = field
		}
	}
	return schema
}
func (s *Schema) RecordsValues(dest interface{}) []interface{} {
	destVal := reflect.Indirect(reflect.ValueOf(dest))
	var feilds []interface{}
	for _, field := range s.Fields {
		feilds = append(feilds, destVal.FieldByName(field.Name).Interface())
	}
	return feilds
}
