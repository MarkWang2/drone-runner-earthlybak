package main

import (
	"context"

	"github.com/drone-runners/drone-runner-docker/debugger/server"
	"github.com/drone-runners/drone-runner-docker/slog"

	"github.com/sirupsen/logrus"
)

const addr = "0.0.0.0:8373"

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	log := slog.GetLogger(ctx).With("app", "shellrepeater")

	x := server.NewServer(addr, log)
	x.Start()
}
