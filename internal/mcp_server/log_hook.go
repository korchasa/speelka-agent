package mcp_server

import (
	"context"

	"github.com/sirupsen/logrus"
)

type LogHook struct {
	server *MCPServer
}

func (h *LogHook) Fire(entry *logrus.Entry) error {
	err := h.server.SendNotificationToClient(context.Background(), "notifications/message", entry.Data)
	if err != nil {
		return err
	}
	return nil
}

func (h *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
