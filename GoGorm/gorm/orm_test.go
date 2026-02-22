package gorm_test

import (
	"GoGorm/gorm"
	"GoGorm/session"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"testing"
)

type TxUser struct {
	Name string `foundry:"PRIMARY KEY"`
	Age  int
}

func initTxTable(t *testing.T, engine *gorm.Engine) {
	t.Helper()
	s := engine.NewSession().Model(&TxUser{})
	if err := s.DropTable(); err != nil {
		t.Fatalf("drop table failed: %v", err)
	}
	if err := s.CreateTable(); err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	if _, err := s.Insert(&TxUser{Name: "Tom", Age: 18}); err != nil {
		t.Fatalf("seed data failed: %v", err)
	}
}

func TestEngine_TransactionCommitAndRollback(t *testing.T) {
	engine, err := gorm.NewEngine("sqlite3", "gee.db")
	if err != nil {
		t.Fatalf("new engine failed: %v", err)
	}
	defer engine.Close()
	initTxTable(t, engine)

	err = engine.Transaction(func(tx *session.Session) error {
		tx.Model(&TxUser{})
		_, err := tx.Insert(&TxUser{Name: "CommitUser", Age: 20})
		return err
	})
	if err != nil {
		t.Fatalf("transaction commit failed: %v", err)
	}

	s := engine.NewSession().Model(&TxUser{})
	count, err := s.Count()
	if err != nil {
		t.Fatalf("count after commit failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2 after commit, got %d", count)
	}

	err = engine.Transaction(func(tx *session.Session) error {
		tx.Model(&TxUser{})
		if _, err := tx.Insert(&TxUser{Name: "RollbackUser", Age: 21}); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected rollback error, got nil")
	}

	count, err = s.Count()
	if err != nil {
		t.Fatalf("count after rollback failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2 after rollback, got %d", count)
	}
}
