package dbx

import "fmt"

type Scope struct {
	cond   Cond
	db     string
	table  string
	m      *Model
	limit  uint32
	offset uint32
}

func (s *Scope) GetModel() *Model {
	return s.m
}

func (s *Scope) SetTablePrefix(prefix string) *Scope {
	s.cond.tablePrefix = prefix
	return s
}

func (p *Scope) Where(args ...interface{}) *Scope {
	p.cond.Where(args...)
	return p
}

func (p *Scope) Lt(f string, v interface{}) *Scope {
	p.Where(fmt.Sprintf("`%s` < ?", f), v)
	return p
}

func (p *Scope) Lte(f string, v interface{}) *Scope {
	p.Where(fmt.Sprintf("`%s` <= ?", f), v)
	return p
}

func (p *Scope) Gt(f string, v interface{}) *Scope {
	p.Where(fmt.Sprintf("`%s` > ?", f), v)
	return p
}

func (p *Scope) Gte(f string, v interface{}) *Scope {
	p.Where(fmt.Sprintf("`%s` >= ?", f), v)
	return p
}

func (p *Scope) OrWhere(args ...interface{}) *Scope {
	p.cond.OrWhere(args...)
	return p
}
