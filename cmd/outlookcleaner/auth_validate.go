package main

import (
	"context"

	"github.com/raokrutarth/golang-playspace/pkg/logger"
)

func AuthValidate(ctx context.Context) {
	log := logger.GetLoggerFromContext(ctx)
	var err error
	connections, err := NewMailAccountConnections(ctx)
	if err != nil {
		log.Error("failed to get account", "error", err)
		return
	}
	for _, c := range connections {
		if err = c.client.Logout(); err != nil {
			log.Error("failed logout", "error", err)
		}
	}
}
