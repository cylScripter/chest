package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

		// Handle arrays and slices
		if fieldValue.Kind() == reflect.Array || fieldValue.Kind() == reflect.Slice {
			arrayLength := fieldValue.Len()
			var arrayElements []string
			for j := 0; j < arrayLength; j++ {
				elementValue := fieldValue.Index(j)
				switch elementValue.Kind() {
				case reflect.String:
					arrayElements = append(arrayElements, elementValue.String())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					arrayElements = append(arrayElements, strconv.FormatInt(elementValue.Int(), 10))
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					arrayElements = append(arrayElements, strconv.FormatUint(elementValue.Uint(), 10))
				case reflect.Float32, reflect.Float64:
					arrayElements = append(arrayElements, strconv.FormatFloat(elementValue.Float(), 'f', -1, 64))
				case reflect.Bool:
					arrayElements = append(arrayElements, strconv.FormatBool(elementValue.Bool()))
				default:
					// Unsupported types are ignored
				}
			}
			result[jsonTag] = "[" + strings.Join(arrayElements, ",") + "]"
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

type ModelFile struct {
	Id         int32   `json:"id"`
	CreatedAt  int32   `json:"created_at"`
	UpdatedAt  int32   `json:"updated_at"`
	FileName   string  `json:"file_name"`
	FilePath   string  `json:"file_path"`
	FileType   int32   `json:"file_type"`
	Status     int32   `json:"status"`
	UploadId   string  `json:"upload_id"`
	Meta       []int32 `json:"meta"`
	Suffix     string  `json:"suffix"`
	StrFileId  string  `json:"str_file_id"`
	BucketName string  `json:"bucket_name"`
	DeleteAt   int32   `json:"delete_at"`
}

type ModelFile_Meta struct {
	Size    int32  `json:"size"`
	Author  string `json:"author"`
	Version string `json:"version"`
}

func main() {
	//meta := ModelFile_Meta{
	//	Size:    12345,
	//	Author:  "John Doe",
	//	Version: "1.0.0",
	//}

	model := ModelFile{
		Id:         1,
		CreatedAt:  1623445567,
		UpdatedAt:  1623445567,
		FileName:   "example.txt",
		FilePath:   "/files/example.txt",
		FileType:   1,
		Status:     0,
		UploadId:   "upload123",
		Meta:       []int32{1, 2, 35},
		Suffix:     ".txt",
		StrFileId:  "hash123",
		BucketName: "my-bucket",
		DeleteAt:   0,
	}

	serialized, err := serializeStructToMap(model)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Serialized Map: %v\n", serialized["meta"])
}
