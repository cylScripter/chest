package handler

import (
	"context"
	"errors"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/cylScripter/chest/log"
	"github.com/cylScripter/chest/rpc"
	"reflect"
	"strings"
)

func ClientErrorHandler(ctx context.Context, err error) error {
	// if you want get other rpc info, you can get rpcinfo first, like `ri := rpcinfo.GetRPCInfo(ctx)`
	// for example, get remote address: `remoteAddr := rpcinfo.GetRPCInfo(ctx).To().Address()`
	var transError *remote.TransError
	if errors.As(err, &transError) {
		// TypeID is error code
		return errors.Unwrap(err)
	}
	return rpc.CreateErrorWithMsg(-1, err.Error())
}

func ServerErrorHandler(ctx context.Context, err error) error {
	if errors.Is(err, kerrors.ErrBiz) {
		return errors.Unwrap(err)
	}
	return rpc.CreateErrorWithMsg(-1, err.Error())
}

// ValidateRequest request body 校验
func ValidateRequest(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request, response interface{}) error {
		if arg, ok := request.(utils.KitexArgs); ok {
			requestType := reflect.TypeOf(arg.GetFirstArgument())
			// 获取User结构体的反射值
			requestValue := reflect.ValueOf(arg.GetFirstArgument())
			if requestType.Kind() != reflect.Ptr {
				return rpc.CreateErrorWithMsg(rpc.KErrRequestBodyReadFail, "request body is not a pointer")
			}
			if requestType.Elem().Kind() != reflect.Struct {
				return rpc.CreateErrorWithMsg(rpc.KErrRequestBodyReadFail, "request body is not a struct")
			}
			for i := 0; i < requestType.Elem().NumField(); i++ {
				field := requestType.Elem().Field(i)
				fieldValue := requestValue.Elem().Field(i)
				jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
				if field.Tag.Get("validate") != "" && fieldValue.IsZero() {
					return rpc.InvalidArg("request body field " + jsonTag + " is required")
				}
			}
		}
		err := next(ctx, request, response)
		if err != nil {
			log.Infof("err %v", err)
			return err
		}
		return nil
	}
}
