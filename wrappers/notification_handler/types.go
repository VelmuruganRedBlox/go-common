package notification_handler

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/wrappers"
	"time"
)

type INotificationHandlerWrapper interface {
	EnqueueNotificationWithTemplate(templateName string, userId int64,
		renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	EnqueueNotificationWithCustomTemplate(title, body, headline string, userId int64, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	GetNotificationsReadCount(notificationIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]int64]
	DisableUnregisteredTokens(tokens []string, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[[]string]
}

//goland:noinspection GoNameStartsWithPackageName
type NotificationHandlerWrapper struct {
	baseWrapper     *wrappers.BaseWrapper
	defaultTimeout  time.Duration
	apiUrl          string
	serviceName     string
	publisher       eventsourcing.IEventPublisher
	customPublisher eventsourcing.IEventPublisher
}

type EnqueueMessageResult struct {
	Id    string `json:"id"`
	Error error  `json:"error"`
}
type SendNotificationWithTemplate struct {
	Id                 string                 `json:"id"`
	TemplateName       string                 `json:"template_name"`
	RenderingVariables map[string]string      `json:"rendering_variables"`
	CustomData         map[string]interface{} `json:"custom_data"`
	UserId             int64                  `json:"user_id"`
}

func (s SendNotificationWithTemplate) GetPublishKey() string {
	return fmt.Sprint(s.UserId)
}

type SendNotificationWithCustomTemplate struct {
	Id         string                 `json:"id"`
	UserId     int64                  `json:"user_id"`
	Title      string                 `json:"title"`
	Body       string                 `json:"body"`
	Headline   string                 `json:"headline"`
	CustomData map[string]interface{} `json:"custom_data"`
}

func (s SendNotificationWithCustomTemplate) GetPublishKey() string {
	return fmt.Sprint(s.UserId)
}

type NotificationChannel int

const (
	NotificationChannelPush  = NotificationChannel(1)
	NotificationChannelEmail = NotificationChannel(2)
)

type GetNotificationsReadCountRequest struct {
	NotificationIds []int64 `json:"notification_ids"`
}

type DisableUnregisteredTokensRequest struct {
	Tokens []string `json:"tokens"`
}
