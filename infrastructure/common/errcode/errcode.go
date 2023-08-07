package errcode

import (
	"fmt"
	"reflect"
)

type ErrCode struct {
	Code   int32  `json:"code"`
	Msg    string `json:"msg"`
	Detail string `json:"-"` // 错误详情
}

func (e *ErrCode) Error() string {
	return e.Msg
}

func (e *ErrCode) String() string {
	return e.Msg
}

func (e *ErrCode) WithMsg(msg string) *ErrCode {
	return &ErrCode{
		Code:   e.Code,
		Msg:    msg,
		Detail: e.Detail,
	}
}

func (e *ErrCode) WithMsgRemark(msgRemark string) *ErrCode {
	msg := fmt.Sprintf("%s(%s)", e.Msg, msgRemark)
	return &ErrCode{
		Code:   e.Code,
		Msg:    msg,
		Detail: e.Detail,
	}
}

func (e *ErrCode) WithDetail(detail string) *ErrCode {
	return &ErrCode{
		Code:   e.Code,
		Msg:    e.Msg,
		Detail: detail,
	}
}

func (e *ErrCode) WithErr(err error) *ErrCode {
	if err != nil {
		return &ErrCode{
			Code:   e.Code,
			Msg:    e.Msg,
			Detail: err.Error(),
		}
	}
	return e
}

func (e *ErrCode) As(target interface{}) bool {
	targetVal := reflect.ValueOf(target)
	return reflect.TypeOf(e).AssignableTo(targetVal.Type())
}
