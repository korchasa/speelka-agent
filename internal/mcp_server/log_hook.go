package mcp_server

import (
	"context"

	"github.com/sirupsen/logrus"
)

type LogHook struct {
	server *MCPServer
	ctx    context.Context
}

func (h *LogHook) Fire(entry *logrus.Entry) error {
	err := h.server.SendNotificationToClient(h.ctx, "notifications/message", entry.Data)
	if err != nil {
		return err
	}
	return nil
}

func (h *LogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
