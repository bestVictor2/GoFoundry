package schema

import (
	"GoGorm/dialect"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}
type FoundryUser struct {
	Name string `foundry:"PRIMARY KEY"`
	Age  int
}

var TestDial = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, TestDial)
	if schema.Name != "User" || len(schema.Fields) != 2 {
		t.Fatal("failed to parse User struct")
	}
	if schema.GetField("Name").Tag != "PRIMARY KEY" {
		t.Fatal("failed to parse primary key")
	}
}

func TestParseFoundryTag(t *testing.T) {
	schema := Parse(&FoundryUser{}, TestDial)
	if schema.GetField("Name").Tag != "PRIMARY KEY" {
		t.Fatal("failed to parse foundry primary key")
	}
}
