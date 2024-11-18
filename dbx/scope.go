package dbx

import (
	"fmt"
	"github.com/cylScripter/chest/utils"
	"reflect"
	"strings"
)

type Scope struct {
	cond                Cond
	db                  string
	table               string
	m                   *Model
	limit               uint32
	offset              uint32
	needCount           bool
	orderDesc           bool
	selects             []string
	skips               []string
	groups              []string
	orders              []string
	trId                string
	ignoreConflict      bool
	unscoped            bool
	returnUnknownFields bool
	enableCache         bool
	showSql             bool
	ignoreBroken        bool
	rowsAffected        uint64
}

func (s *Scope) GetModel() *Model {
	return s.m
}

func (s *Scope) Model(model interface{}) *Scope {
	m := &Model{
		ModelConfig: ModelConfig{Type: model},
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
	s.m = m
	return s
}

func (p *Model) Select(fields ...string) *Scope {
	s := p.NewScope()
	s.Select(fields...)
	return s
}

func (s *Scope) SetTablePrefix(prefix string) *Scope {
	s.cond.tablePrefix = prefix
	return s
}
func (s *Scope) SetLimit(limit uint32) *Scope {
	s.limit = limit
	return s
}

func (s *Scope) SetOffset(offset uint32) *Scope {
	s.offset = offset
	return s
}
func (s *Scope) Omit(columns ...string) *Scope {
	s.skips = append(s.skips, columns...)
	return s
}
func (s *Scope) Group(fields ...string) *Scope {
	s.groups = append(s.groups, fields...)
	return s
}
func (s *Scope) EnableCache() *Scope {
	s.enableCache = true
	return s
}

func (s *Scope) ResetGroup(fields ...string) *Scope {
	s.groups = append([]string{}, fields...)
	return s
}
func (s *Scope) Transaction(trId string) *Scope {
	s.trId = trId
	return s
}

func (s *Scope) ResetSelect(fields ...string) *Scope {
	s.selects = append([]string{}, fields...)
	return s
}
func (s *Scope) OrderAsc(fields ...string) *Scope {
	s.orders = append(s.orders, fields...)
	s.orderDesc = false
	return s
}
func (s *Scope) OrderDesc(fields ...string) *Scope {
	s.orders = append(s.orders, fields...)
	s.orderDesc = true
	return s
}
func (s *Scope) ResetOrderAsc(fields ...string) *Scope {
	s.orders = append([]string{}, fields...)
	s.orderDesc = false
	return s
}
func (s *Scope) ResetOrderDesc(fields ...string) *Scope {
	s.orders = append([]string{}, fields...)
	s.orderDesc = true
	return s
}
func (s *Scope) getOrder() string {
	if len(s.orders) > 0 {
		o := strings.Join(s.orders, ",")
		var c string
		if s.orderDesc {
			c = "DESC"
		} else {
			c = "ASC"
		}
		return fmt.Sprintf("%s %s", o, c)
	}
	return ""
}
func (s *Scope) getGroup() string {
	return strings.Join(s.groups, ",")
}
func (s *Scope) GetTableName() string {
	if s.table != "" {
		return s.table
	}
	return s.m.tableName
}

func (s *Scope) Unscoped() *Scope {
	s.unscoped = true
	return s
}

func (s *Scope) Where(args ...interface{}) *Scope {
	s.cond.Where(args...)
	return s
}

func (s *Scope) Lt(f string, v interface{}) *Scope {
	s.Where(fmt.Sprintf("`%s` < ?", f), v)
	return s
}

func (s *Scope) Lte(f string, v interface{}) *Scope {
	s.Where(fmt.Sprintf("`%s` <= ?", f), v)
	return s
}

func (s *Scope) Gt(f string, v interface{}) *Scope {
	s.Where(fmt.Sprintf("`%s` > ?", f), v)
	return s
}

func (s *Scope) Gte(f string, v interface{}) *Scope {
	s.Where(fmt.Sprintf("`%s` >= ?", f), v)
	return s
}

func (s *Scope) OrWhere(args ...interface{}) *Scope {
	s.cond.OrWhere(args...)
	return s
}

func (s *Scope) GetCondString() string {
	return s.cond.ToString()
}

func (s *Scope) Select(fields ...string) *Scope {
	s.selects = append(s.selects, fields...)
	return s
}
