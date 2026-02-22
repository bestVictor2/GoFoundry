package session

import (
	"GoGorm/clause"
	"GoGorm/dialect"
	"GoGorm/log"
	"GoGorm/schema"
	"context"
	"database/sql"
	"errors"
	"strings"
)

type Session struct {
	db       *sql.DB
	tx       *sql.Tx
	dialect  dialect.Dialect
	refTable *schema.Schema
	clause   clause.Clause
	sql      strings.Builder
	sqlVals  []interface{}
	ctx      context.Context
	selects  []string
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{db: db, sql: strings.Builder{}, sqlVals: make([]interface{}, 0), dialect: dialect, ctx: context.Background()}
}
func (s *Session) reSet() {
	s.sql.Reset()
	s.sqlVals = nil
	s.clause = clause.Clause{}
	s.selects = nil
}
func (s *Session) DB() *sql.DB {
	return s.db
}
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sqlVals = append(s.sqlVals, values...)
	return s
}
func (s *Session) WithContext(ctx context.Context) *Session {
	if ctx == nil {
		s.ctx = context.Background()
		return s
	}
	s.ctx = ctx
	return s
}
func (s *Session) Select(fields ...string) *Session {
	s.selects = append(s.selects[:0], fields...)
	return s
}
func (s *Session) Begin() error {
	if s.tx != nil {
		return errors.New("transaction already started")
	}
	tx, err := s.db.Begin()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	s.tx = tx
	return nil
}
func (s *Session) Commit() error {
	if s.tx == nil {
		return errors.New("transaction not started")
	}
	if err := s.tx.Commit(); err != nil {
		log.Error(err.Error())
		return err
	}
	s.tx = nil
	return nil
}
func (s *Session) Rollback() error {
	if s.tx == nil {
		return errors.New("transaction not started")
	}
	if err := s.tx.Rollback(); err != nil {
		log.Error(err.Error())
		return err
	}
	s.tx = nil
	return nil
}
func (s *Session) Exec() (sql.Result, error) {
	defer s.reSet()
	log.Info(s.sql.String(), s.sqlVals)
	var result sql.Result
	var err error
	if s.tx != nil {
		result, err = s.tx.ExecContext(s.ctx, s.sql.String(), s.sqlVals...)
	} else {
		result, err = s.db.ExecContext(s.ctx, s.sql.String(), s.sqlVals...)
	}
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return result, nil
}
func (s *Session) QueryRow() *sql.Row {
	defer s.reSet()
	log.Info(s.sql.String(), s.sqlVals)
	if s.tx != nil {
		return s.tx.QueryRowContext(s.ctx, s.sql.String(), s.sqlVals...)
	}
	return s.db.QueryRowContext(s.ctx, s.sql.String(), s.sqlVals...)
}
func (s *Session) QueryRows() (*sql.Rows, error) {
	defer s.reSet()
	log.Info(s.sql.String(), s.sqlVals)
	var rows *sql.Rows
	var err error
	if s.tx != nil {
		rows, err = s.tx.QueryContext(s.ctx, s.sql.String(), s.sqlVals...)
	} else {
		rows, err = s.db.QueryContext(s.ctx, s.sql.String(), s.sqlVals...)
	}
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return rows, nil
}
