package notification_handler

import (
	"context"
)

type NotificationHandlerWrapperMock struct {
	EnqueueNotificationWithTemplateFn func(templateName string, userId int64,
		renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
	EnqueueNotificationWithCustomTemplateFn func(title, body, headline string, userId int64, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult
}

func (m *NotificationHandlerWrapperMock) EnqueueNotificationWithTemplate(templateName string, userId int64,
	renderingVars map[string]string, customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult {
	return m.EnqueueNotificationWithTemplateFn(templateName, userId, renderingVars, customData, ctx)
}

func GetMock() INotificationHandlerWrapper { // for compiler errors
	return &NotificationHandlerWrapperMock{}
}

func (m *NotificationHandlerWrapperMock) EnqueueNotificationWithCustomTemplate(title, body, headline string, userId int64,
	customData map[string]interface{}, ctx context.Context) chan EnqueueMessageResult {
	return m.EnqueueNotificationWithCustomTemplateFn(title, body, headline, userId, customData, ctx)
}
