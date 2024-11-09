package rpc

import (
	"context"
	"errors"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/remote"
)

func ClientErrorHandler(ctx context.Context, err error) error {
	// if you want get other rpc info, you can get rpcinfo first, like `ri := rpcinfo.GetRPCInfo(ctx)`
	// for example, get remote address: `remoteAddr := rpcinfo.GetRPCInfo(ctx).To().Address()`
	var transError *remote.TransError
	if errors.As(err, &transError) {
		// TypeID is error code
		return errors.Unwrap(err)
	}
	return CreateErrorWithMsg(-1, err.Error())
}

func ServerErrorHandler(ctx context.Context, err error) error {
	if errors.Is(err, kerrors.ErrBiz) {
		return errors.Unwrap(err)
	}
	return CreateErrorWithMsg(-1, err.Error())
}
