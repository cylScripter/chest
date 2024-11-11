package dbx

import (
	"fmt"
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
}
