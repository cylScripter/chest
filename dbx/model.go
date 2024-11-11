package dbx

import "reflect"

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
	return m
}
