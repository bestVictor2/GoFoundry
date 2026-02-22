package session_test

import "testing"

func TestSession_TransactionRollback(t *testing.T) {
	s := testRecordInit(t)
	if err := s.Begin(); err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}
	if _, err := s.Insert(&User{Name: "RollbackUser", Age: 28}); err != nil {
		t.Fatalf("insert in transaction failed: %v", err)
	}
	if err := s.Rollback(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}
	count, err := s.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected count 2 after rollback, got %d", count)
	}
}

func TestSession_TransactionCommit(t *testing.T) {
	s := testRecordInit(t)
	if err := s.Begin(); err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}
	if _, err := s.Insert(&User{Name: "CommitUser", Age: 19}); err != nil {
		t.Fatalf("insert in transaction failed: %v", err)
	}
	if err := s.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}
	count, err := s.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3 after commit, got %d", count)
	}
}
