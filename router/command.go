package router

import (
	"fmt"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers/auth_go"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"strconv"
	"strings"
)

type ICommand interface {
	RequireIdentityValidation() bool
	AccessLevel() common.AccessLevel
	GetMethodName() string
	GetFn() CommandFunc
	ForceLog() bool
	GetPath() string
	GetHttpMethod() string
	CanExecute(ctx *fasthttp.RequestCtx, apmTransaction *apm.Transaction, auth auth_go.IAuthGoWrapper) (int64, *rpc.RpcError)
}

type CommandFunc func(request []byte, executionData MethodExecutionData) (interface{}, *error_codes.ErrorWithCode)

type Command struct {
	methodName                string
	forceLog                  bool
	fn                        CommandFunc
	requireIdentityValidation bool
}

func (c *Command) Execute(request []byte, data MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
	return c.fn(request, data)
}

func NewCommand(methodName string, fn CommandFunc, forceLog bool, requireIdentityValidation bool) ICommand {
	return &Command{
		methodName:                strings.ToLower(methodName),
		forceLog:                  forceLog,
		fn:                        fn,
		requireIdentityValidation: requireIdentityValidation,
	}
}

func (c Command) GetMethodName() string {
	return c.methodName
}

func (c Command) GetPath() string { // for rest
	return c.GetMethodName()
}

func (c Command) AccessLevel() common.AccessLevel {
	return common.AccessLevelPublic
}

func (c Command) RequireIdentityValidation() bool {
	return c.requireIdentityValidation
}

func (c Command) GetHttpMethod() string {
	return "post"
}

func (c Command) GetFn() CommandFunc {
	return c.fn
}

func (c Command) CanExecute(ctx *fasthttp.RequestCtx, apmTransaction *apm.Transaction, auth auth_go.IAuthGoWrapper) (int64, *rpc.RpcError) {
	return publicCanExecuteLogic(ctx, c.requireIdentityValidation)
}

func publicCanExecuteLogic(ctx *fasthttp.RequestCtx, requireIdentityValidation bool) (int64, *rpc.RpcError) {
	var userId int64
	if externalAuthValue := ctx.Request.Header.Peek("X-Ext-Authz-Check-Result"); strings.EqualFold(string(externalAuthValue), "allowed") {
		if userIdHead := ctx.Request.Header.Peek("User-Id"); len(userIdHead) > 0 {
			if userIdParsed, err := strconv.ParseInt(string(userIdHead), 10, 64); err != nil {
				return 0, &rpc.RpcError{
					Code:        error_codes.InvalidJwtToken,
					Message:     fmt.Sprintf("can not parse str to int for user-id. input string %v. [%v]", userIdHead, err.Error()),
					Hostname:    hostName,
					ServiceName: hostName,
				}
			} else {
				userId = userIdParsed
			}
		}
	}

	if requireIdentityValidation && userId <= 0 {
		return 0, &rpc.RpcError{
			Code:        error_codes.MissingJwtToken,
			Message:     "public method requires identity validation",
			Hostname:    hostName,
			ServiceName: hostName,
		}
	}

	return userId, nil
}

func (c Command) ForceLog() bool {
	if c.forceLog {
		return true
	}

	if c.AccessLevel() > common.AccessLevelRead {
		return true
	}

	return false
}
