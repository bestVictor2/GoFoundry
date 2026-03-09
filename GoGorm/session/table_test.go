package session_test

import (
	"GoGorm/gorm"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"testing"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

type UserOnlyName struct {
	Name string `geeorm:"PRIMARY KEY"`
}

type UserBigAge struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int64
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

func TestSession_AutoMigrate_AddColumnKeepData(t *testing.T) {
	engine, _ := gorm.NewEngine("sqlite3", "gee.db")
	s := engine.NewSession().Model(&User{})
	_ = s.DropTable()
	if _, err := s.Raw("CREATE TABLE User (Name text PRIMARY KEY);").Exec(); err != nil {
		t.Fatalf("create old table failed: %v", err)
	}
	if _, err := s.Raw("INSERT INTO User(Name) VALUES (?);", "Tom").Exec(); err != nil {
		t.Fatalf("insert old row failed: %v", err)
	}

	if err := s.AutoMigrate(); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	var got User
	if err := s.Where("Name = ?", "Tom").First(&got); err != nil {
		t.Fatalf("query row after migrate failed: %v", err)
	}
	if got.Name != "Tom" {
		t.Fatalf("name should be kept, got %s", got.Name)
	}
	if got.Age != 0 {
		t.Fatalf("new column should use zero value, got %d", got.Age)
	}
}

func TestSession_AutoMigrate_RemoveColumnKeepData(t *testing.T) {
	engine, _ := gorm.NewEngine("sqlite3", "gee.db")
	s := engine.NewSession().Model(&UserOnlyName{})
	_ = s.DropTable()
	if _, err := s.Raw("CREATE TABLE UserOnlyName (Name text PRIMARY KEY, Age integer);").Exec(); err != nil {
		t.Fatalf("create old table failed: %v", err)
	}
	if _, err := s.Raw("INSERT INTO UserOnlyName(Name, Age) VALUES (?, ?);", "Sam", 18).Exec(); err != nil {
		t.Fatalf("insert old row failed: %v", err)
	}

	if err := s.AutoMigrate(); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	columns, err := s.Columns()
	if err != nil {
		t.Fatalf("query columns failed: %v", err)
	}
	if strings.Contains(strings.Join(columns, ","), "Age") {
		t.Fatalf("removed column should be dropped, got %v", columns)
	}

	var got UserOnlyName
	if err = s.Where("Name = ?", "Sam").First(&got); err != nil {
		t.Fatalf("query row after migrate failed: %v", err)
	}
	if got.Name != "Sam" {
		t.Fatalf("name should be kept, got %s", got.Name)
	}
}

func TestSession_AutoMigrate_TypeChangeKeepData(t *testing.T) {
	engine, _ := gorm.NewEngine("sqlite3", "gee.db")
	s := engine.NewSession().Model(&UserBigAge{})
	_ = s.DropTable()
	if _, err := s.Raw("CREATE TABLE UserBigAge (Name text PRIMARY KEY, Age integer);").Exec(); err != nil {
		t.Fatalf("create old table failed: %v", err)
	}
	if _, err := s.Raw("INSERT INTO UserBigAge(Name, Age) VALUES (?, ?);", "Neo", 21).Exec(); err != nil {
		t.Fatalf("insert old row failed: %v", err)
	}

	if err := s.AutoMigrate(); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	var got UserBigAge
	if err := s.Where("Name = ?", "Neo").First(&got); err != nil {
		t.Fatalf("query row after migrate failed: %v", err)
	}
	if got.Age != 21 {
		t.Fatalf("age should be kept, got %d", got.Age)
	}
}
