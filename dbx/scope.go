package dbx

import (
	"context"
	"errors"
	"fmt"
	"github.com/cylScripter/chest/log"
	"github.com/cylScripter/chest/utils"
	"github.com/cylScripter/openapi/base"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/json"
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
		ModelConfig: ModelConfig{
			Type:            model,
			NotFoundErrCode: s.m.NotFoundErrCode,
			Db:              s.m.Db,
		},
		proxy: s.m.proxy,
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

func (p *Model) Unscoped() *Scope {
	return p.WithTrash()
}
func (p *Model) WithTrash() *Scope {
	s := p.NewScope()
	s.Unscoped()
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

func (s *Scope) Find(ctx context.Context, dest interface{}) error {

	if len(s.groups) > 0 {
		s.needCount = true
	}
	var orders []string
	if len(s.orders) > 0 {
		orders = append(orders, s.getOrder())
	}
	return s.m.proxy.Find(ctx, &WhereReq{
		Unscoped:  s.unscoped,
		Cond:      []string{s.GetCondString()},
		Groups:    []string{s.getGroup()},
		Limit:     s.limit,
		Offset:    s.offset,
		Orders:    orders,
		Selects:   s.selects,
		TableName: s.GetTableName(),
	}, dest)
}
func (s *Scope) ToSql(ctx context.Context, dest interface{}) (string, error) {
	model := s.m.getModel()
	s.Model(model)
	if len(s.groups) > 0 {
		s.needCount = true
	}
	return s.m.proxy.ToSql(ctx, &WhereReq{
		Cond:      []string{s.GetCondString()},
		Groups:    []string{s.getGroup()},
		Limit:     s.limit,
		Offset:    s.offset,
		Orders:    []string{s.getOrder()},
		Selects:   s.selects,
		needGroup: s.needCount,
		TableName: s.GetTableName(),
	}, dest)
}
func (s *Scope) First(ctx context.Context, dest interface{}) error {
	if len(s.groups) > 0 {
		s.needCount = true
	}
	var orders []string
	if len(s.orders) > 0 {
		orders = append(orders, s.getOrder())
	}
	return s.m.proxy.First(ctx, &WhereReq{
		needGroup: s.needCount,
		Unscoped:  s.unscoped,
		Cond:      []string{s.GetCondString()},
		Groups:    []string{s.getGroup()},
		Limit:     s.limit,
		Offset:    s.offset,
		Orders:    orders,
		Selects:   s.selects,
		TableName: s.GetTableName(),
	}, dest)
}
func (s *Scope) FindPaginate(ctx context.Context, dest interface{}) (*base.Paginate, error) {
	if len(s.groups) > 0 {
		s.needCount = true
	}
	var orders []string
	if len(s.orders) > 0 {
		orders = append(orders, s.getOrder())
	}
	return s.m.proxy.FindPaginate(ctx, &WhereReq{
		needGroup: s.needCount,
		Unscoped:  s.unscoped,
		Cond:      []string{s.GetCondString()},
		Groups:    []string{s.getGroup()},
		Limit:     s.limit,
		Offset:    s.offset,
		Orders:    orders,
		Selects:   s.selects,
		TableName: s.GetTableName(),
	}, dest)
}
func (s *Scope) Create(ctx context.Context, dest interface{}) error {
	return s.m.proxy.Create(ctx, &CreateReq{
		TableName: s.GetTableName(),
		Selects:   s.selects,
		Omit:      s.skips,
	}, dest)
}
func (s *Scope) Count(ctx context.Context) (int64, error) {
	model := s.m.getModel()
	if len(s.groups) > 0 {
		s.needCount = true
	}
	var orders []string
	if len(s.orders) > 0 {
		orders = append(orders, s.getOrder())
	}
	return s.m.proxy.Count(ctx, &WhereReq{
		Unscoped:  s.unscoped,
		Limit:     s.limit,
		Offset:    s.offset,
		Cond:      []string{s.GetCondString()},
		Groups:    []string{s.getGroup()},
		needGroup: s.needCount,
		Orders:    orders,
		TableName: s.GetTableName(),
	}, model)
}
func (s *Scope) UseDb(db string) *Scope {
	s.db = db
	return s
}

func (s *Scope) UseTable(table string) *Scope {
	s.table = table
	return s
}
func (s *Scope) Update(ctx context.Context, values map[string]interface{}) (UpdateResult, error) {
	model := s.m.getModel()
	return s.m.proxy.Update(ctx, &WhereReq{
		Unscoped:  s.unscoped,
		Cond:      []string{s.GetCondString()},
		TableName: s.GetTableName(),
	}, model, values)

}
func (s *Scope) Delete(ctx context.Context) (DeleteResult, error) {
	model := s.m.getModel()
	return s.m.proxy.Delete(ctx, &WhereReq{
		Unscoped:  s.unscoped,
		Cond:      []string{s.GetCondString()},
		TableName: s.GetTableName(),
	}, model)
}

func (s *Scope) FirstOrCreate(ctx context.Context, attributes map[string]interface{}, values map[string]interface{}, obj interface{}) (FirstOrCreateResult, error) {
	res := FirstOrCreateResult{}
	err := s.Where(attributes).First(ctx, obj)
	all := make(map[string]interface{})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			for k, v := range attributes {
				all[k] = v
			}
			for k, v := range values {
				all[k] = v
			}
			err = map2Interface(all, obj)
			err = s.Create(ctx, obj)
			if err != nil {
				return FirstOrCreateResult{}, err
			} else {
				res.Created = true
			}

		} else {
			return FirstOrCreateResult{}, err
		}
	}
	return res, nil
}

func (s *Scope) FirstOrUpdate(ctx context.Context, attributes map[string]interface{}, values map[string]interface{}, obj interface{}) (FirstOrCreateResult, error) {
	res := FirstOrCreateResult{}
	err := s.Where(attributes).First(ctx, obj)
	all := make(map[string]interface{})
	if err != nil {
		return FirstOrCreateResult{}, err
	}
	for k, v := range attributes {
		all[k] = v
	}
	for k, v := range values {
		all[k] = v
	}
	_, err = s.Update(ctx, all)
	if err != nil {
		return FirstOrCreateResult{}, err
	}
	err = map2Interface(all, obj)
	if err != nil {
		return FirstOrCreateResult{}, err
	}
	res.Created = true

	return res, nil
}

func (s *Scope) Save(ctx context.Context, dest interface{}) error {
	return s.m.proxy.Save(ctx, &WhereReq{
		TableName: s.GetTableName(),
	}, dest)
}

func map2Interface(m map[string]interface{}, i interface{}) error {
	if nm, ok := i.(*map[string]interface{}); ok {
		p := *nm
		for k, v := range m {
			p[k] = v
		}
	} else {
		b, err := json.Marshal(m)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		err = json.Unmarshal(b, i)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

func (s *Scope) WhereIn(fieldName string, list interface{}) *Scope {
	vo := utils.EnsureIsSliceOrArray(list)
	if vo.Len() == 0 {
		s.cond.where(false)
		return s
	}
	s.cond.where(fieldName, "IN", utils.UniqueSlice(vo.Interface()))
	return s
}
func (s *Scope) WhereNotIn(fieldName string, list interface{}) *Scope {
	vo := utils.EnsureIsSliceOrArray(list)
	if vo.Len() == 0 {
		return s
	}
	s.cond.where(fieldName, "NOT IN", utils.UniqueSlice(vo.Interface()))
	return s
}

//func (p *Scope) WhereRaw(cond string) *Scope {
//	p.cond.whereRaw(cond)
//	return p
//}
//
//// mysql 是包含 min, max, 相当于 field >= min && field <= max
//func (p *Scope) WhereBetween(fieldName string, min, max interface{}) *Scope {
//	p.cond.whereRaw(fmt.Sprintf("%s BETWEEN ? AND ?", quoteFieldName(fieldName)), min, max)
//	return p
//}
//func (p *Scope) WhereToday(fieldName string) *Scope {
//	today := time.Now()
//	p.WhereDate(fieldName, today)
//	return p
//}
//func (p *Scope) WhereDate(fieldName string, t time.Time) *Scope {
//	p.cond.whereRaw(fmt.Sprintf("%s BETWEEN ? AND ?", quoteFieldName(fieldName)), utils.BeginTimeStampOfDate(t), utils.EndTimeStampOfDate(t))
//	return p
//}
//
//// 相当于 field < min || field > max
//func (p *Scope) WhereNotBetween(fieldName string, min, max interface{}) *Scope {
//	p.cond.whereRaw(fmt.Sprintf("%s NOT BETWEEN ? AND ?", quoteFieldName(fieldName)), min, max)
//	return p
//}
//func (p *Scope) WhereLike(fieldName string, str string) *Scope {
//	p.cond.conds = append(p.cond.conds, fmt.Sprintf("(%s LIKE %s)", quoteFieldName(fieldName), quoteStr(utils.EscapeMysqlLikeWildcardIgnore2End(str))))
//	return p
//}
//func (p *Scope) WhereNotLike(fieldName string, str string) *Scope {
//	p.cond.conds = append(p.cond.conds, fmt.Sprintf("(%s NOT LIKE %s)", quoteFieldName(fieldName), quoteStr(utils.EscapeMysqlLikeWildcardIgnore2End(str))))
//	return p
//}
