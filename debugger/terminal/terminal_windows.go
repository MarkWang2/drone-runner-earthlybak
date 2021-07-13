// +build windows

package terminal

import (
	"context"

	"github.com/drone-runners/drone-runner-docker/conslogging"

	"github.com/pkg/errors"
)

func ConnectTerm(ctx context.Context, addr string, console conslogging.ConsoleLogger) error {
	return errors.New("debugger not supported on Windows yet")
}
