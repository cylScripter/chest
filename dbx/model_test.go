package dbx

import (
	"context"
	"github.com/cylScripter/chest/log"
	"testing"
)

type ModelUser struct {
	Id          int32  `thrift:"id,1" frugal:"1,default,i32" gorm:"column:id" json:"id"`
	CreatedAt   int32  `thrift:"created_at,2" frugal:"2,default,i32" json:"created_at"`
	UpdatedAt   int32  `thrift:"updated_at,3" frugal:"3,default,i32" json:"updated_at"`
	DeletedAt   int32  `thrift:"deleted_at,4" frugal:"4,default,i32" json:"deleted_at"`
	UserId      string `thrift:"user_id,5" frugal:"5,default,string" json:"user_id"`
	Password    string `thrift:"password,6" frugal:"6,default,string" json:"-"`
	Mobile      string `thrift:"mobile,7" frugal:"7,default,string" json:"mobile"`
	Email       string `thrift:"email,8" frugal:"8,default,string" json:"email"`
	Nickname    string `thrift:"nickname,9" frugal:"9,default,string" json:"nickname"`
	Avatar      string `thrift:"avatar,10" frugal:"10,default,string" json:"avatar"`
	Status      bool   `thrift:"status,11" frugal:"11,default,bool" json:"status"`
	LastLoginAt int32  `thrift:"last_login_at,12" frugal:"12,default,i32" json:"last_login_at"`
}

type TUser struct {
	*Model
}

var orm *Db

func init() {
	var err error
	orm, err = NewDb(DbConfig{
		DbName:   "test",
		DbType:   "mysql",
		Ip:       "106.53.50.226",
		Port:     19457,
		User:     "root",
		Password: "Jy18300015530@",
	})
	if err != nil {
		panic(err)
	}

	User = &TUser{
		Model: NewModel(&ModelConfig{
			&ModelUser{},
			5000,
			"user",
		}, orm),
	}
}

var User *TUser

func TestNewModel(t *testing.T) {

	var userList []*ModelUser
	err := User.NewScope().Where("deleted_at", 0).Find(context.Background(), &userList)
	if err != nil {
		log.Infof("err:%v", err)
		return
	}

	sql, err := User.NewScope().Select("id").Where("deleted_at", 0).OrderAsc("id").ToSql(context.Background(), &userList)
	log.Infof("sql:%v", sql)

	log.Infof("userList:%v", userList[0].UserId)

}
