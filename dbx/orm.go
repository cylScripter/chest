package dbx

import (
	"context"
	"fmt"
	"github.com/cylScripter/chest/log"
	"github.com/cylScripter/chest/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

type DbProxy interface {
	Find(ctx context.Context, req *WhereReq, dest interface{}) error
	ToSql(ctx context.Context, req *WhereReq, dest interface{}) (string, error)
	AutoMigrate(dest ...interface{}) error
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
	TableName string
}

type FindResp struct {
}

func (p *Db) Find(ctx context.Context, req *WhereReq, dest interface{}) error {
	query := p.db
	if req.TableName != "" {
		query = query.Table(req.TableName)
	} else {
		modelType := strings.ReplaceAll(fmt.Sprintf("%T", dest), "[]", "")
		query = query.Table(utils.CamelToSnake(modelType))
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

func (p *Db) Create(ctx context.Context, req *WhereReq, dest interface{}) error {
	query := p.db
	if req.TableName != "" {
		query = query.Table(req.TableName)
	} else {
		modelType := fmt.Sprintf("%T", dest)
		query = query.Table(utils.CamelToSnake(modelType))
	}
	return query.Create(dest).Error
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
