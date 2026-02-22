package gorm

import (
	"GoGorm/dialect"
	"GoGorm/log"
	"GoGorm/session"
	"database/sql"
)

type Engine struct {
	dialect dialect.Dialect
	db      *sql.DB
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if err = db.Ping(); err != nil {
		log.Error(err.Error())
		return
	}
	dial := dialect.GetDialect(driver)

	e = &Engine{db: db, dialect: dial}
	log.Info("Connected to database")
	return
}
func (engine *Engine) Close() (err error) {
	if err = engine.db.Close(); err != nil {
		log.Info(err.Error())
		return
	}
	return
}
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

func (engine *Engine) Transaction(fn func(s *session.Session) error) (err error) {
	s := engine.NewSession()
	if err = s.Begin(); err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		}
	}()
	if err = fn(s); err != nil {
		_ = s.Rollback()
		return err
	}
	return s.Commit()
}
