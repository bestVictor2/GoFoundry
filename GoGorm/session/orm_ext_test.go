package session_test

import (
	"GoGorm/gorm"
	"GoGorm/session"
	"context"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"testing"
)

type HookUser struct {
	Name string `foundry:"PRIMARY KEY"`
	Age  int
}

func (u *HookUser) BeforeInsert(_ *session.Session) error {
	u.Name = strings.ToUpper(u.Name)
	return nil
}

func (u *HookUser) AfterQuery(_ *session.Session) error {
	u.Name = "HOOKED-" + u.Name
	return nil
}

func TestSession_HookAndSelect(t *testing.T) {
	engine, err := gorm.NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatalf("new engine failed: %v", err)
	}
	defer engine.Close()

	s := engine.NewSession().Model(&HookUser{})
	_ = s.DropTable()
	if err = s.AutoMigrate(); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
	hookUser := &HookUser{Name: "tom", Age: 18}
	if _, err = s.Insert(hookUser); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if hookUser.Name != "TOM" {
		t.Fatalf("before insert hook not called, got %s", hookUser.Name)
	}

	var got HookUser
	if err = s.Where("Name = ?", "TOM").First(&got); err != nil {
		t.Fatalf("first failed: %v", err)
	}
	if got.Name != "HOOKED-TOM" {
		t.Fatalf("after query hook not called, got %s", got.Name)
	}

	var list []HookUser
	if err = s.Select("Name").Find(&list); err != nil {
		t.Fatalf("select failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 row, got %d", len(list))
	}
	if list[0].Age != 0 {
		t.Fatalf("expected age zero when not selected, got %d", list[0].Age)
	}
}

func TestSession_WithContextCanceled(t *testing.T) {
	engine, err := gorm.NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatalf("new engine failed: %v", err)
	}
	defer engine.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := engine.NewSession().WithContext(ctx)
	row := s.Raw("SELECT 1").QueryRow()
	var n int
	if err = row.Scan(&n); err == nil {
		t.Fatal("expected canceled context error")
	}
}
