package dbx

import (
	"context"
	"fmt"
	"github.com/cylScripter/chest/utils"
	"reflect"
)

type ModelConfig struct {
	Type            interface{}
	NotFoundErrCode int
	Db              string
}

type Model struct {
	ModelConfig
	proxy       DbProxy
	typ         reflect.Type
	tableName   string
	modelType   string
	notFoundErr error
}

func NewModel(c *ModelConfig, proxy DbProxy) *Model {
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
	m.tableName = utils.CamelToSnake(m.modelType)
	m.proxy = proxy
	return m
}
func (p *Model) NewScope() *Scope {
	s := &Scope{}
	s.m = p
	s.db = p.Db
	s.cond.isTopLevel = true
	return s
}
func (p *Model) UnScoped() *Scope {
	s := p.NewScope()
	s.Unscoped()
	return s
}
func (p *Model) Where(whereCond ...interface{}) *Scope {
	s := p.NewScope()
	s.cond.Where(whereCond...)
	return s
}
func (p *Model) OrWhere(whereCond ...interface{}) *Scope {
	s := p.NewScope()
	s.cond.isOr = true
	s.cond.OrWhere(whereCond...)
	return s
}

func (p *Model) getModel() interface{} {
	return p.Type
}
func (p *Model) Create(ctx context.Context, dest interface{}) error {
	s := p.NewScope()
	return s.Create(ctx, dest)
}

func (p *Model) FirstOrCreate(ctx context.Context, attributes map[string]interface{}, values map[string]interface{}, obj interface{}) (FirstOrCreateResult, error) {
	return p.NewScope().FirstOrCreate(ctx, attributes, values, obj)
}

func (p *Model) FirstOrUpdate(ctx context.Context, attributes map[string]interface{}, values map[string]interface{}, obj interface{}) (FirstOrCreateResult, error) {
	return p.NewScope().FirstOrUpdate(ctx, attributes, values, obj)
}

func (p *Model) Save(ctx context.Context, dest interface{}) error {
	s := p.NewScope()
	return s.Save(ctx, dest)
}
