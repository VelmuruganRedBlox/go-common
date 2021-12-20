package follow

import (
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/boilerplate"
	"github.com/digitalmonsters/go-common/common"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/rpc"
	"github.com/digitalmonsters/go-common/wrappers"
	"go.elastic.co/apm"
	"time"
)

type IFollowWrapper interface {
	GetFollowContentUserByContentAuthorIdsInternal(contentAuthorIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan FollowContentUserByContentAuthorIdsResponseChan
}

//goland:noinspection GoNameStartsWithPackageName
type FollowWrapper struct {
	baseWrapper    *wrappers.BaseWrapper
	defaultTimeout time.Duration
	apiUrl         string
	serviceName    string
}

func NewFollowWrapper(config boilerplate.WrapperConfig) IFollowWrapper {
	timeout := 5 * time.Second

	if config.TimeoutSec > 0 {
		timeout = time.Duration(config.TimeoutSec) * time.Second
	}

	return &FollowWrapper{
		baseWrapper:    wrappers.GetBaseWrapper(),
		defaultTimeout: timeout,
		apiUrl:         fmt.Sprintf("%v/rpc", common.StripSlashFromUrl(config.ApiUrl)),
		serviceName:    "follows",
	}
}

func (w *FollowWrapper) GetFollowContentUserByContentAuthorIdsInternal(contentAuthorIds []int64, userId int64, apmTransaction *apm.Transaction, forceLog bool) chan FollowContentUserByContentAuthorIdsResponseChan {
	respCh := make(chan FollowContentUserByContentAuthorIdsResponseChan, 2)

	respChan := w.baseWrapper.SendRpcRequest(w.apiUrl, "GetFollowContentUserByContentAuthorIdsInternal", FollowContentUserByContentAuthorIdsRequest{
		UserId:     userId,
		ContentAuthorIds: contentAuthorIds,
	}, w.defaultTimeout, apmTransaction, w.serviceName, forceLog)

	w.baseWrapper.GetPool().Submit(func() {
		defer func() {
			close(respCh)
		}()

		resp := <-respChan

		result := FollowContentUserByContentAuthorIdsResponseChan{
			Error: resp.Error,
		}

		if len(resp.Result) > 0 {
			data := map[int64]bool{}

			if err := json.Unmarshal(resp.Result, &data); err != nil {
				result.Error = &rpc.RpcError{
					Code:        error_codes.GenericMappingError,
					Message:     err.Error(),
					Data:        nil,
					Hostname:    w.baseWrapper.GetHostName(),
					ServiceName: w.serviceName,
				}
			} else {
				result.Data = data
			}
		}

		respCh <- result
	})

	return respCh
}
