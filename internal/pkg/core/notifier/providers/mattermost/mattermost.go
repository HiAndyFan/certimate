package mattermost

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/nikoksr/notify/service/mattermost"
	"github.com/usual2970/certimate/internal/pkg/core/notifier"
)

type NotifierConfig struct {
	// Mattermost 服务地址。
	ServerUrl string `json:"serverUrl"`
	// 频道ID
	ChannelId string `json:"channelId"`
	// 用户名
	Username string `json:"username"`
	// 密码
	Password string `json:"password"`
}

type NotifierProvider struct {
	config *NotifierConfig
	logger *slog.Logger
}

var _ notifier.Notifier = (*NotifierProvider)(nil)

func NewNotifier(config *NotifierConfig) (*NotifierProvider, error) {
	if config == nil {
		panic("config is nil")
	}

	return &NotifierProvider{
		config: config,
	}, nil
}

func (n *NotifierProvider) WithLogger(logger *slog.Logger) notifier.Notifier {
	if logger == nil {
		n.logger = slog.Default()
	} else {
		n.logger = logger
	}
	return n
}

func (n *NotifierProvider) Notify(ctx context.Context, subject string, message string) (res *notifier.NotifyResult, err error) {
	srv := mattermost.New(n.config.ServerUrl)

	if err := srv.LoginWithCredentials(ctx, n.config.Username, n.config.Password); err != nil {
		return nil, err
	}

	srv.AddReceivers(n.config.ChannelId)

	// 复写消息样式
	srv.PreSend(func(req *http.Request) error {
		m := map[string]interface{}{
			"channel_id": n.config.ChannelId,
			"props": map[string]interface{}{
				"attachments": []map[string]interface{}{
					{
						"title": subject,
						"text":  message,
					},
				},
			},
		}

		if body, err := json.Marshal(m); err != nil {
			return err
		} else {
			req.ContentLength = int64(len(body))
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		return nil
	})

	if err = srv.Send(ctx, subject, message); err != nil {
		return nil, err
	}

	return &notifier.NotifyResult{}, nil
}
