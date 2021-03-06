// +build !windows

package terminal

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/drone-runners/drone-runner-docker/conslogging"
	"github.com/drone-runners/drone-runner-docker/debugger/common"

	"github.com/creack/pty"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

func handlePtyData(data []byte) error {
	_, err := os.Stdout.Write(data)
	if err != nil {
		return errors.Wrap(err, "failed to write data to stdout")
	}
	return nil
}

func getWindowSizePayload() ([]byte, error) {
	size, err := pty.GetsizeFull(os.Stdin)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(size)
	if err != nil {
		return nil, err
	}
	return common.SerializeDataPacket(common.WinSizeData, b)
}

// ConnectTerm presents a terminal to the shell repeater
func ConnectTerm(ctx context.Context, addr string, console conslogging.ConsoleLogger) error {
	var d net.Dialer

	console.VerbosePrintf("connecting to shellrepeater on %v\n", addr)
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte{common.TermID})
	if err != nil {
		return errors.Wrap(err, "failed to write TermID connection")
	}

	sigs := make(chan os.Signal, 10)
	signal.Notify(sigs, syscall.SIGWINCH)

	writeCh := make(chan []byte, 10)

	ctx, cancel := context.WithCancel(ctx)

	ts := &termState{}
	go func() {
	outer:
		for {
			connDataType, data, err := common.ReadDataPacket(conn)
			if err != nil {
				console.Warnf("ReadDataPacket failed: %s\n", err.Error())
				break
			}
			switch connDataType {
			case common.StartShellSession:
				console.VerbosePrintf("starting new interactive shell pseudo terminal\n")
				err := ts.makeRaw()
				if err != nil {
					console.Warnf("makeRaw failed: %s\n", err.Error())
					break outer
				}
				sigs <- syscall.SIGWINCH
			case common.EndShellSession:
				err := ts.restore()
				if err != nil {
					console.Warnf("restore failed: %s\n", err.Error())
					break outer
				}
			case common.PtyData:
				err := handlePtyData(data)
				if err != nil {
					console.Warnf("handlePtyData failed: %s\n", err.Error())
					break outer
				}
			default:
				console.Warnf("unhandled terminal data type: %q\n", connDataType)
				break outer
			}
		}
		cancel()
	}()

	go func() {
		for range sigs {
			if len(sigs) > 0 {
				continue
			}
			data, err := getWindowSizePayload()
			if err != nil {
				console.Warnf("failed to get window size payload: %s\n", err.Error())
				break
			}
			writeCh <- data
		}
		cancel()
	}()

	go func() {
		for {
			buf := <-writeCh
			_, err := conn.Write(buf)
			if err != nil {
				console.Warnf("failed to send term data to shell: %s\n", err.Error())
				break
			}
		}
		cancel()
	}()
	go func() {
		for {
			buf := make([]byte, 100)
			n, err := os.Stdin.Read(buf)
			if err != nil {
				console.Warnf("failed to read from stdin: %s\n", err.Error())
				break
			}
			buf = buf[:n]
			buf2, err := common.SerializeDataPacket(common.PtyData, buf)
			if err != nil {
				console.Warnf("failed to serialize data: %s\n", err.Error())
				break
			}

			writeCh <- buf2
		}
		cancel()
	}()

	<-ctx.Done()

	console.VerbosePrintf("exiting interactive debugger shell\n")
	err = ts.restore()
	if err != nil {
		return err
	}
	return nil
}

type termState struct {
	oldState *terminal.State
	mu       sync.Mutex
}

func (ts *termState) makeRaw() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.oldState == nil {
		var err error
		ts.oldState, err = terminal.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return errors.Wrap(err, "failed to initialize terminal in raw mode")
		}
	}
	return nil
}

func (ts *termState) restore() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if ts.oldState != nil {
		err := terminal.Restore(int(os.Stdin.Fd()), ts.oldState)
		if err != nil {
			return errors.Wrap(err, "failed to restore terminal mode")
		}
		ts.oldState = nil
	}
	return nil
}
