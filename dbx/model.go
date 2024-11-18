package dbx

import (
	"fmt"
	"github.com/cylScripter/chest/log"
	"reflect"
)

type ModelConfig struct {
	Type            interface{}
	NotFoundErrCode int
	Db              string
}

type Model struct {
	ModelConfig
	typ         reflect.Type
	tableName   string
	modelType   string
	notFoundErr error
}

func NewModel(c *ModelConfig) *Model {
	m := &Model{
		ModelConfig: *c,
	}
	if m.Type == nil {
		panic("Type nil")
	}
	typ := reflect.TypeOf(m.Type)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	m.typ = typ
	m.modelType = fmt.Sprintf("%T", m.Type)
	return m
}
func (p *Model) NewScope() *Scope {
	s := &Scope{
		m:  p,
		db: p.Db,
	}
	s.cond.isTopLevel = true
	return s
}

func (p *Model) Where(whereCond ...interface{}) *Scope {
	s := p.NewScope()
	log.Infof(" whereCond:%v", s.GetCondString())
	s.cond.Where(whereCond...)
	return s
}

func (p *Model) OrWhere(whereCond ...interface{}) *Scope {
	s := p.NewScope()
	s.cond.OrWhere(whereCond...)
	return s
}
