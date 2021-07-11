package engine2

import (
	"context"
	"github.com/drone-runners/drone-runner-docker/engine"
	"github.com/drone/runner-go/pipeline/runtime"
	"io"
)

type Earthly struct {
}

// New returns a new engine.
func New(opts engine.Opts) *Earthly {
	return &Earthly{}
}

// NewEnv returns a new Engine from the environment.
func NewEnv(opts engine.Opts) (*Earthly, error) {
	return New(opts), nil
}

// Ping pings the Earthly daemon.
func (e *Earthly) Ping(ctx context.Context) error {
	return nil
}

// Setup the pipeline environment.
func (e *Earthly) Setup(ctx context.Context, specv runtime.Spec) error {
	//cmd := exec.Command("./earthly", "+build")
	//cmd.Stdin = strings.NewReader("some input")
	//var out bytes.Buffer
	//cmd.Stdout = &out
	//cmd.Run()
	return nil
}

// Run runs the pipeline step.
func (e *Earthly) Run(ctx context.Context, specv runtime.Spec, stepv runtime.Step, output io.Writer) (*runtime.State, error) {
	//spec := specv.(*Spec)
	//step := stepv.(*Step)

	//// create the container
	//err := e.create(ctx, spec, step, output)
	//if err != nil {
	//	return nil, errors.TrimExtraInfo(err)
	//}
	//// start the container
	//err = e.start(ctx, step.ID)
	//if err != nil {
	//	return nil, errors.TrimExtraInfo(err)
	//}
	//// tail the container
	//err = e.tail(ctx, step.ID, output)
	//if err != nil {
	//	return nil, errors.TrimExtraInfo(err)
	//}
	// wait for the response
	return &runtime.State{}, nil
}
