package types

import "ns-tracking-go/pkg/errorx"

type GetTrackingReq struct {
	OrderNumber string `path:"order_number"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewResponse(code int, message string, data interface{}) Response {
	return Response{Code: code, Message: message, Data: data}
}

func Success(data interface{}) Response {
	return Response{Code: int(errorx.Success), Message: "success", Data: data}
}

func Error(err *errorx.CodeError) Response {
	if err == nil {
		return Response{Code: 500, Message: "unknown error"}
	}
	return Response{Code: err.GetCode(), Message: err.Error()}
}

func ErrorWithData(err *errorx.CodeError, data interface{}) Response {
	if err == nil {
		return Response{Code: 500, Message: "unknown error", Data: data}
	}
	return Response{Code: err.GetCode(), Message: err.Error(), Data: data}
}
