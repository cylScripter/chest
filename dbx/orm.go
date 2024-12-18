package dbx

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cylScripter/chest/log"
	"github.com/cylScripter/chest/utils"
	"github.com/cylScripter/openapi/base"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const DefaultLimit = 2000

type DbProxy interface {
	First(ctx context.Context, req *WhereReq, dest interface{}) error
	Find(ctx context.Context, req *WhereReq, dest interface{}) error
	Create(ctx context.Context, req *CreateReq, dest interface{}) error
	ToSql(ctx context.Context, req *WhereReq, dest interface{}) (string, error)
	FindPaginate(ctx context.Context, req *WhereReq, dest interface{}) (*base.Paginate, error)
	Count(ctx context.Context, req *WhereReq, dest interface{}) (int64, error)
	Delete(ctx context.Context, req *WhereReq, dest interface{}) (DeleteResult, error)
	AutoMigrate(dest ...interface{}) error
	Update(ctx context.Context, req *WhereReq, dest interface{}, values map[string]interface{}) (UpdateResult, error)
	Save(ctx context.Context, req *WhereReq, dest interface{}) error
}

type DbConfig struct {
	DbName       string
	DbType       string
	User         string
	Password     string
	Ip           string
	Port         int
	MaxIdleCoins int // 最大空闲连接数
}

type Db struct {
	config DbConfig
	db     *gorm.DB
}

func (p *Db) GetModel(tableName string, dest interface{}) *gorm.DB {
	query := p.db
	if tableName != "" {
		query = query.Table(tableName)
	} else {
		modelType := strings.ReplaceAll(fmt.Sprintf("%T", dest), "[]", "")
		query = query.Table(utils.CamelToSnake(modelType))
	}
	return query
}

func NewDb(cfg DbConfig) (*Db, error) {
	dns := fmt.Sprintf("%s:%s@tcp(%v:%v)/%s?charset=utf8&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Ip, cfg.Port, cfg.DbName)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dns,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{})
	if err != nil {
		log.Errorf("NewDb failed, err:%v", err)
		return nil, err
	}
	return &Db{
		config: cfg,
		db:     db,
	}, nil
}

type WhereReq struct {
	Limit     uint32
	Offset    uint32
	Selects   []string
	Groups    []string
	Orders    []string
	Cond      []string
	needGroup bool
	Unscoped  bool
	TableName string
}

type CreateReq struct {
	TableName string
	Selects   []string
	Omit      []string
}

type DeleteResult struct {
	RowsAffected uint64
}
type UpdateResult struct {
	RowsAffected uint64
	Sql          string
}
type SelectResult struct {
	Total      uint32
	NextOffset uint32
}
type UpdateOrCreateResult struct {
	Created      bool
	RowsAffected uint64
}
type FirstOrCreateResult struct {
	Created bool
}

func (p *Db) Find(ctx context.Context, req *WhereReq, dest interface{}) error {
	query := p.GetModel(req.TableName, dest)
	if !req.Unscoped {
		query = query.Scopes(ScopeGetIsDel())
	}
	if req.Limit > 0 {
		query = query.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		query = query.Offset(int(req.Offset))
	}
	if len(req.Selects) > 0 {
		query = query.Select(req.Selects)
	}
	// where
	for _, cond := range req.Cond {
		query = query.Where(cond)
	}
	// group
	if req.needGroup {
		log.Infof("groups:%v", req.Groups)
		for _, group := range req.Groups {
			query = query.Group(group)
		}
	}
	for _, order := range req.Orders {
		query = query.Order(order)
	}

	return query.Find(dest).Error
}

func (p *Db) ToSql(ctx context.Context, req *WhereReq, dest interface{}) (string, error) {
	sql := p.db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		query := tx
		if req.TableName != "" {
			query = query.Table(req.TableName)
		} else {
			modelType := strings.ReplaceAll(fmt.Sprintf("%T", dest), "[]", "")
			query = query.Table(utils.CamelToSnake(modelType))
		}

		if !req.Unscoped {
			query = query.Scopes(ScopeGetIsDel())
		}
		for _, cond := range req.Cond {
			query = query.Where(cond)
		}
		for _, order := range req.Orders {
			query = query.Order(order)
		}
		// group
		if req.needGroup {
			log.Infof("groups:%v", req.Groups)
			for _, group := range req.Groups {
				query = query.Group(group)
			}
		}
		if len(req.Selects) > 0 {
			query = query.Select(req.Selects)
		}
		return query.Find(dest)
	})
	return sql, nil
}

func (p *Db) Create(ctx context.Context, req *CreateReq, dest interface{}) error {
	query := p.GetModel(req.TableName, dest)
	if len(req.Omit) > 0 {
		query = query.Omit(req.Omit...)
	}
	if len(req.Selects) > 0 {
		query = query.Select(req.Selects)
	}
	data, err := serializeStructToMap(dest)
	if err != nil {
		return err
	}
	return query.Create(data).Error
}

func (p *Db) First(ctx context.Context, req *WhereReq, dest interface{}) error {
	query := p.GetModel(req.TableName, dest)
	if !req.Unscoped {
		query = query.Scopes(ScopeGetIsDel())
	}
	if len(req.Selects) > 0 {
		query = query.Select(req.Selects)
	}
	// where
	for _, cond := range req.Cond {
		query = query.Where(cond)
	}
	// group
	if req.needGroup {
		for _, group := range req.Groups {
			query = query.Group(group)
		}
	}
	for _, order := range req.Orders {
		query = query.Order(order)
	}
	return query.First(dest).Error
}

func (p *Db) AutoMigrate(dest ...interface{}) error {
	for _, v := range dest {
		modelType := fmt.Sprintf("%T", v)
		err := p.db.Table(utils.CamelToSnake(modelType)).AutoMigrate(v)
		if err != nil {
			log.Errorf("AutoMigrate failed, err:%v", err)
			return err
		}
	}
	return nil
}

func (p *Db) FindPaginate(ctx context.Context, req *WhereReq, dest interface{}) (*base.Paginate, error) {
	result, err := p.FindWithResult(ctx, req, dest)
	return &base.Paginate{
		Total:  int32(result.Total),
		Offset: int32(req.Offset),
		Limit:  int32(req.Limit),
	}, err
}

func (p *Db) FindWithResult(ctx context.Context, req *WhereReq, dest interface{}) (SelectResult, error) {
	var res SelectResult
	var total int64
	query := p.GetModel(req.TableName, dest)
	var limit int
	switch {
	case req.Limit < 0:
		limit = 10
	case req.Limit > DefaultLimit:
		limit = DefaultLimit
	default:
		limit = int(req.Limit)
	}
	if !req.Unscoped {
		query = query.Scopes(ScopeGetIsDel())
	}
	query.Count(&total)
	if limit > 0 {
		query = query.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		query = query.Offset(int(req.Offset))
	}
	if len(req.Selects) > 0 {
		query = query.Select(req.Selects)
	}
	for _, cond := range req.Cond {
		query = query.Where(cond)
	}
	for _, order := range req.Orders {
		query = query.Order(order)
	}
	// group
	if req.needGroup {
		for _, group := range req.Groups {
			query = query.Group(group)
		}
	}
	result := query.Find(dest)
	res.Total = uint32(total)
	return res, result.Error
}

func (p *Db) Count(ctx context.Context, req *WhereReq, dest interface{}) (int64, error) {
	query := p.GetModel(req.TableName, dest).Select("count(id) as count")
	if !req.Unscoped {
		query = query.Scopes(ScopeGetIsDel())
	}
	type res struct {
		Count int64 `json:"count"`
	}
	var result res
	if req.Limit > 0 {
		query = query.Limit(int(req.Limit))
	}

	if req.Offset > 0 {
		query = query.Offset(int(req.Offset))
	}

	for _, cond := range req.Cond {
		query = query.Where(cond)
	}
	if req.needGroup {
		for _, group := range req.Groups {
			query = query.Group(group)
		}
	}
	err := query.First(&result).Error
	return result.Count, err
}

func ScopeGetIsDel() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted_at = 0")
	}
}

func (p *Db) Delete(ctx context.Context, req *WhereReq, dest interface{}) (DeleteResult, error) {
	var res DeleteResult
	query := p.GetModel(req.TableName, dest)
	if !req.Unscoped {
		query = query.Scopes(ScopeGetIsDel())
	}
	for _, cond := range req.Cond {
		query = query.Where(cond)
	}
	result := query.Update("deleted_at", time.Now().Unix())
	res.RowsAffected = uint64(result.RowsAffected)
	return res, result.Error
}

func (p *Db) Update(ctx context.Context, req *WhereReq, dest interface{}, values map[string]interface{}) (UpdateResult, error) {
	res := UpdateResult{}
	query := p.GetModel(req.TableName, dest)
	if !req.Unscoped {
		query = query.Scopes(ScopeGetIsDel())
	}
	for _, cond := range req.Cond {
		query = query.Where(cond)
	}
	result := query.Updates(values)
	res.Sql = result.Statement.SQL.String()
	res.RowsAffected = uint64(result.RowsAffected)
	return res, nil
}

func (p *Db) Save(ctx context.Context, req *WhereReq, dest interface{}) error {
	query := p.GetModel(req.TableName, dest)
	return query.Save(dest).Error
}

func serializeStructToMap(input interface{}) (map[string]string, error) {
	result := make(map[string]string)
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a struct or pointer to struct")
	}

	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		// Get the JSON key or fallback to field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			nestedMap, err := serializeStructToMap(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			// Serialize nested map as JSON
			nestedJSON, err := json.Marshal(nestedMap)
			if err != nil {
				return nil, err
			}
			result[jsonTag] = string(nestedJSON)
			continue
		}

		// Convert values to strings
		switch fieldValue.Kind() {
		case reflect.String:
			result[jsonTag] = fieldValue.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			result[jsonTag] = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			result[jsonTag] = strconv.FormatUint(fieldValue.Uint(), 10)
		case reflect.Float32, reflect.Float64:
			result[jsonTag] = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.Bool:
			result[jsonTag] = strconv.FormatBool(fieldValue.Bool())
		default:
			// Unsupported types are ignored
		}
	}

	return result, nil
}
