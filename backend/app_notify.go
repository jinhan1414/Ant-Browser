package backend

import (
	"context"
	"fmt"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const systemNotificationEventName = "app:system-notify"

type systemNotificationPayload struct {
	AppName string `json:"appName"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

type runtimeEventEmitter func(ctx context.Context, eventName string, data ...interface{})

type appEventNotifier struct {
	ctx     context.Context
	appName string
	emit    runtimeEventEmitter
}

func newAppEventNotifier(app *App) *appEventNotifier {
	if app == nil {
		return &appEventNotifier{}
	}
	return &appEventNotifier{
		ctx:     app.ctx,
		appName: app.appName(),
		emit:    wailsruntime.EventsEmit,
	}
}

func (n *appEventNotifier) Notify(title string, body string) error {
	if n == nil || n.ctx == nil || n.emit == nil {
		return fmt.Errorf("system notification emitter not initialized")
	}
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		title = strings.TrimSpace(n.appName)
	}
	n.emit(n.ctx, systemNotificationEventName, systemNotificationPayload{
		AppName: strings.TrimSpace(n.appName),
		Title:   title,
		Body:    body,
	})
	return nil
}
