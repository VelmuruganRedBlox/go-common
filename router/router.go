package router

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers/auth"
	fastRouter "github.com/fasthttp/router"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
	"go.elastic.co/apm"
	"os"
	"strconv"
	"strings"
	"time"
)

type HttpRouter struct {
	realRouter   *fastRouter.Router
	executor     *CommandExecutor
	hostname     string
	restCommands map[string]*RestCommand
	isProd       bool
	authWrapper  auth.IAuthWrapper
	ready        bool
}

func NewRouter(rpcEndpointPath string, wrapper auth.IAuthWrapper, healthContext context.Context) *HttpRouter {
	h := &HttpRouter{
		realRouter:   fastRouter.New(),
		executor:     NewCommandExecutor(),
		authWrapper:  wrapper,
		restCommands: map[string]*RestCommand{},
		ready:        false,
	}

	if hostname, _ := os.Hostname(); len(hostname) > 0 {
		h.hostname = hostname
	}

	if boilerplate.GetCurrentEnvironment() == boilerplate.Prod {
		h.isProd = true
	}

	h.prepareRpcEndpoint(rpcEndpointPath)

	h.registerHttpHealthCheck(healthContext)
	h.registerHttpReadinessCheck()

	return h
}

func (r *HttpRouter) setCors(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	ctx.Response.Header.SetBytesV("Access-Control-Allow-Origin", ctx.Request.Header.Peek("Origin"))
	ctx.Response.Header.Set("Access-Control-Allow-Headers", "*")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}

func (r *HttpRouter) RegisterRpcCommand(command *Command) error {
	return r.executor.AddCommand(command)
}

func (r *HttpRouter) RegisterRestCmd(targetCmd *RestCommand) error {
	key := fmt.Sprintf("%v_%v", targetCmd.method, targetCmd.path)

	if _, ok := r.restCommands[key]; ok {
		return errors.New(fmt.Sprintf("rest command [%v] already registered", key))
	}

	r.restCommands[key] = targetCmd

	r.realRouter.Handle(targetCmd.method, targetCmd.path, func(ctx *fasthttp.RequestCtx) {
		apmTransaction := apm_helper.StartNewApmTransaction(fmt.Sprintf("[%v] [%v]", targetCmd.method,
			targetCmd.path), "rpc", nil, nil)
		defer apmTransaction.End()

		requestBody := ctx.PostBody()

		rpcRequest := rpc.RpcRequest{
			Method:  targetCmd.path,
			Params:  requestBody,
			Id:      "1",
			JsonRpc: "2.0",
		}

		apm_helper.AddApmLabel(apmTransaction, "full_url", string(ctx.URI().FullURI()))

		rpcResponse, shouldLog := r.executeAction(rpcRequest, targetCmd, ctx, apmTransaction, targetCmd.forceLog,
			func(key string) interface{} {
				if v := ctx.UserValue(key); v != nil {
					return v
				}

				if ctx.QueryArgs() != nil {
					if v := ctx.QueryArgs().Peek(key); len(v) > 0 {
						return string(v)
					}
				}

				return nil
			})

		var responseBody []byte

		defer func() {
			if !shouldLog {
				return
			}

			r.logRequestBody(requestBody, apmTransaction)
			r.logResponseBody(responseBody, apmTransaction)
		}()

		finalStatusCode := int(error_codes.None)

		var err error

		var restResponse genericRestResponse

		restResponse.Success = true
		restResponse.ExecutionTimingMs = rpcResponse.ExecutionTimingMs
		restResponse.Hostname = rpcResponse.Hostname

		if rpcResponse.Result != nil {
			restResponse.Data = rpcResponse.Result
		}
		if rpcResponse.Error != nil {
			originalCode := int(rpcResponse.Error.Code)
			restResponse.Success = false
			restResponse.Error = rpcResponse.Error.Message
			restResponse.Stack = rpcResponse.Error.Stack

			if originalCode > 0 {
				finalStatusCode = originalCode
			} else {
				switch rpcResponse.Error.Code {
				case error_codes.GenericMappingError:
					finalStatusCode = int(error_codes.GenericValidationError)
				case error_codes.CommandNotFoundError:
					finalStatusCode = int(error_codes.GenericNotFoundError)
				default:
					finalStatusCode = int(error_codes.GenericServerError)
				}
			}
		}

		if responseBody, err = json.Marshal(restResponse); err != nil {
			log.Err(err).Send()
		}

		ctx.Response.SetBodyRaw(responseBody)
		ctx.Response.SetStatusCode(finalStatusCode)
	})

	return nil
}

func (r *HttpRouter) executeAction(rpcRequest rpc.RpcRequest, cmd ICommand, ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction, forceLog bool, getUserValue func(key string) interface{}) (rpcResponse rpc.RpcResponse, shouldLog bool) {
	totalTiming := time.Now()

	newCtx, cancel := context.WithCancel(ctx)
	newCtx = apm.ContextWithTransaction(newCtx, apmTransaction)

	defer cancel()

	r.logRequestHeaders(ctx, apmTransaction) // in future filter for specific routes
	r.logUserValues(ctx, apmTransaction)

	var panicErr error

	var executionMs int64

	rpcResponse = rpc.RpcResponse{
		JsonRpc: "2.0",
	}

	defer func() {
		ctx.Response.Header.SetContentType("application/json")
		rpcResponse.ExecutionTimingMs = executionMs
		rpcResponse.TotalTimingMs = time.Since(totalTiming).Milliseconds()
		rpcResponse.Hostname = r.hostname
	}()

	defer func() {
		if rec := recover(); rec != nil {
			shouldLog = true

			switch val := rec.(type) {
			case error:
				panicErr = errors.Wrap(val, fmt.Sprintf("panic! %v", val))
			default:
				panicErr = errors.New(fmt.Sprintf("panic! : %v", val))
			}

			if panicErr == nil {
				panicErr = errors.New("panic! and that is really bad")
			}

			rpcResponse.Result = nil
			rpcResponse.Error = &rpc.RpcError{
				Code:     error_codes.GenericPanicError,
				Message:  panicErr.Error(),
				Data:     nil,
				Hostname: r.hostname,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", panicErr)
			}
		}
	}()

	rpcResponse.Id = rpcRequest.Id

	shouldLog = forceLog

	userId := int64(0)

	if externalAuthValue := ctx.Request.Header.Peek("X-Ext-Authz-Check-Result"); strings.EqualFold(string(externalAuthValue), "allowed") {
		if userIdHead := ctx.Request.Header.Peek("User-Id"); len(userIdHead) > 0 {
			if userIdParsed, err := strconv.ParseInt(string(userIdHead), 10, 64); err != nil {
				apm_helper.CaptureApmError(err, apmTransaction)
			} else {
				userId = userIdParsed
			}
		}
	}

	if userId == 0 {
		if authHeaderValue := ctx.Request.Header.Peek("Authorization"); len(authHeaderValue) > 0 {
			jwtStr := string(authHeaderValue)

			if len(jwtStr) > 0 {
				jwtStr = strings.TrimSpace(strings.ReplaceAll(jwtStr, "Bearer", ""))
			}

			resp := <-r.authWrapper.ParseToken(jwtStr, false, apmTransaction, true)

			if resp.Error != nil {
				rpcResponse.Error = resp.Error
				return
			}

			userId = resp.Resp.UserId
		}
	}

	if userId > 0 {
		if apmTransaction != nil {
			apmTransaction.Context.SetUserID(fmt.Sprint(userId))
		}
	}

	if userId == 0 && (cmd.RequireIdentityValidation() || cmd.AccessLevel() > common.AccessLevelPublic) {
		err := errors.New("missing jwt token for auth")

		rpcResponse.Error = &rpc.RpcError{
			Code:     error_codes.MissingJwtToken,
			Message:  "missing jwt token for auth",
			Hostname: r.hostname,
		}

		if !r.isProd {
			rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
		}
		return
	}

	executionTiming := time.Now()

	if resp, err := cmd.GetFn()(rpcRequest.Params, MethodExecutionData{
		ApmTransaction: apmTransaction,
		Context:        newCtx,
		UserId:         userId,
		getUserValueFn: getUserValue,
	}); err != nil {
		rpcResponse.Error = &rpc.RpcError{
			Code:     err.GetCode(),
			Message:  err.GetMessage(),
			Data:     nil,
			Hostname: r.hostname,
		}

		shouldLog = true

		if !r.isProd {
			rpcResponse.Error.Stack = err.GetStack()
		}
	} else {
		if resp == nil {
			resp = "ok"
		}

		rpcResponse.Result = resp
	}

	executionMs = time.Since(executionTiming).Milliseconds()
	return
}

func (r *HttpRouter) prepareRpcEndpoint(rpcEndpointPath string) {
	r.realRouter.OPTIONS(rpcEndpointPath, func(ctx *fasthttp.RequestCtx) {
		r.setCors(ctx)
	})

	r.realRouter.POST(rpcEndpointPath, func(ctx *fasthttp.RequestCtx) {
		var rpcRequest rpc.RpcRequest
		var rpcResponse rpc.RpcResponse
		var shouldLog bool
		var requestBody []byte
		var apmTransaction *apm.Transaction

		defer func() {
			r.setCors(ctx)
		}()
		
		defer func() {
			var responseBody []byte

			if rpcResponse.Result != nil || rpcResponse.Error != nil {
				if respBody, err := json.Marshal(rpcResponse); err != nil {
					shouldLog = true
					rpcResponse.Result = nil
					rpcResponse.Error = &rpc.RpcError{
						Code:     error_codes.GenericMappingError,
						Message:  errors.Wrap(err, "error during response serialization").Error(),
						Data:     nil,
						Hostname: r.hostname,
					}
					if !r.isProd {
						rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
					}

					if respBody, err1 := json.Marshal(rpcResponse); err1 != nil {
						responseBody = []byte(fmt.Sprintf("that`s really not good || %v", err1.Error()))
					} else {
						responseBody = respBody
					}
				} else {
					responseBody = respBody
				}

				ctx.Response.SetBodyRaw(responseBody)
			}

			if rpcResponse.Error != nil {
				shouldLog = true
			}

			if shouldLog {
				r.logRequestBody(requestBody, apmTransaction)
				r.logResponseBody(responseBody, apmTransaction)
			}
		}()

		requestBody = ctx.PostBody()

		if err := json.Unmarshal(requestBody, &rpcRequest); err != nil {
			rpcResponse.Error = &rpc.RpcError{
				Code:     error_codes.GenericMappingError,
				Message:  err.Error(),
				Data:     nil,
				Hostname: r.hostname,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
			}

			return
		}

		apmTransaction = apm_helper.StartNewApmTransaction(rpcRequest.Method, "rpc", nil, nil)
		defer apmTransaction.End()

		cmd, err := r.executor.GetCommand(rpcRequest.Method)

		if err != nil {
			rpcResponse.Error = &rpc.RpcError{
				Code:     error_codes.CommandNotFoundError,
				Message:  err.Error(),
				Data:     nil,
				Hostname: r.hostname,
			}

			if !r.isProd {
				rpcResponse.Error.Stack = fmt.Sprintf("%+v", err)
			}

			return
		}

		rpcResponse, shouldLog = r.executeAction(rpcRequest, cmd, ctx, apmTransaction, cmd.forceLog, nil)
	})
}

func (r *HttpRouter) GET(path string, handler fasthttp.RequestHandler) {
	r.realRouter.GET(path, handler)
}

func (r *HttpRouter) logRequestBody(body []byte, apmTransaction *apm.Transaction) {
	if apmTransaction == nil {
		return
	}

	if len(body) > 0 {
		apm_helper.AddApmData(apmTransaction, "request_body", body)
	}
}

func (r *HttpRouter) logResponseBody(responseBody []byte,
	apmTransaction *apm.Transaction) {
	if body := responseBody; len(body) > 0 {
		apm_helper.AddApmData(apmTransaction, "response_body", body)
	}
}

func (r *HttpRouter) logUserValues(ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction) string {
	var realMethodName string

	ctx.VisitUserValues(func(key []byte, i interface{}) {
		keyStr := string(key)
		valueStr, ok := i.(string)

		if !ok { // not supported cast
			return
		}

		apm_helper.AddApmLabel(apmTransaction, keyStr, valueStr)
	})

	return realMethodName
}

func (r *HttpRouter) logRequestHeaders(ctx *fasthttp.RequestCtx,
	apmTransaction *apm.Transaction) {
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		keyStr := strings.ToLower(string(key))

		if keyStr == "cookies" || keyStr == "authorization" || keyStr == "x-forwarded-client-cert" {
			return
		}

		valueStr := string(value)

		apm_helper.AddApmLabel(apmTransaction, keyStr, valueStr)
	})
}

func (r *HttpRouter) Router() *fastRouter.Router {
	return r.realRouter
}

func (r *HttpRouter) Handler() func(ctx *fasthttp.RequestCtx) {
	return fasthttp.CompressHandlerBrotliLevel(r.realRouter.Handler, 11, 9)
}

func (r *HttpRouter) GetRpcRegisteredCommands() []Command {
	var commands []Command

	if r.executor.commands != nil {
		for _, c := range r.executor.commands {
			commands = append(commands, *c)
		}
	}

	return commands
}

func (r *HttpRouter) GetRestRegisteredCommands() []RestCommand {
	var commands []RestCommand

	if r.restCommands != nil {
		for _, c := range r.restCommands {
			commands = append(commands, *c)
		}
	}

	return commands
}

func (r *HttpRouter) registerHttpHealthCheck(healthContext context.Context) {
	r.GET("/health", func(ctx *fasthttp.RequestCtx) {
		if healthContext.Err() == nil {
			ctx.Response.SetStatusCode(200)
		} else {
			ctx.Response.SetStatusCode(500)
		}
	})
}

func (r *HttpRouter) registerHttpReadinessCheck() {
	r.GET("/readiness", func(ctx *fasthttp.RequestCtx) {
		if r.ready {
			ctx.Response.SetStatusCode(200)
		} else {
			ctx.Response.SetStatusCode(500)
		}
	})
}

func (r *HttpRouter) Ready() {
	r.ready = true
}

func (r *HttpRouter) NotReady() {
	r.ready = false
}
