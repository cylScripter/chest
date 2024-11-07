package rpc

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type ErrMsg struct {
	ErrCode int32  `json:"err_code,omitempty"`
	ErrMsg  string `json:"err_msg,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

const (
	KSystemError = -1
	// KErrRequestBodyReadFail 服务端读取请求数据异常
	KErrRequestBodyReadFail = -2002
	// KErrResponseMarshalFail 服务返回数据序列化失败
	KErrResponseMarshalFail = -2003
	// KProcessPanic 业务处理异常
	KProcessPanic       = -2004
	KExceedMaxCallDepth = -2005
)

func CreateErrorWithMsg(errCode int32, errMsg string) *ErrMsg {
	return &ErrMsg{ErrCode: errCode, ErrMsg: errMsg}
}
func (err *ErrMsg) Error() string {
	return fmt.Sprintf("errcode %d, errmsg %s", err.ErrCode, err.ErrMsg)
}

// FromError 反序列化
func FromError(err error) *ErrMsg {
	var errMsg *ErrMsg
	ok := errors.As(err, &errMsg)
	if ok {
		return errMsg
	}
	// 使用正则表达式解析错误消息
	re := regexp.MustCompile(`errcode (\d+), errmsg (.+)`)
	matches := re.FindStringSubmatch(err.Error())
	if len(matches) == 3 {
		errCode, _ := strconv.ParseInt(matches[1], 10, 32)
		return &ErrMsg{
			ErrCode: int32(errCode),
			ErrMsg:  matches[2],
		}
	}
	// 解析失败，返回一个系统错误的 ErrMsg 实例
	return &ErrMsg{
		ErrCode: KSystemError,
		ErrMsg:  err.Error(),
	}
}
