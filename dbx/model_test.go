package dbx

import (
	"fmt"
	"github.com/cylScripter/chest/utils"
	"testing"
)

type User struct {
	Id   int64
	Name string
}

func TestNewModel(t *testing.T) {
	cond := &Cond{
		isTopLevel: true,
	}
	cond.OrWhere("id", 1).OrWhere("name", "aaa")
	fmt.Println(cond.ToString())

	model := NewModel(&ModelConfig{
		Db:              "default",
		Type:            User{},
		NotFoundErrCode: 1000,
	})

	fmt.Println(fmt.Sprintf("%T", model.Type))

	fmt.Println(model.OrWhere("id = ? OR name = ?", 1, "aaa").OrWhere("name", "aaa").GetCondString())

	input := "dbx.ModelUserList"
	output := utils.CamelToSnake(input)
	fmt.Println(output) // 输出: modle_user_list

}
