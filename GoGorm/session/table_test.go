package session_test

import (
	"GoGorm/gorm"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	engine, _ := gorm.NewEngine("sqlite3", "gee.db")
	s := engine.NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}

func TestSession_AutoMigrate(t *testing.T) {
	engine, _ := gorm.NewEngine("sqlite3", "gee.db")
	s := engine.NewSession().Model(&User{})
	_ = s.DropTable()
	if err := s.AutoMigrate(); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	if !s.HasTable() {
		t.Fatal("auto migrate should create table")
	}
}
