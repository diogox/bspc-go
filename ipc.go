package bspc

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"syscall"
)

var (
	errInvalidUnixSocket = errors.New("invalid unix socket")
	errClosedUnixSocket  = errors.New("unix socket has been closed")
)

type ipcCommand string

// intoMessage adds NULL to the end of every word in the command.
// This is necessary because bspwm's C code expects it.
func (ic ipcCommand) intoMessage() string {
	var msg string

	words := strings.Split(string(ic), " ")
	for _, w := range words {
		msg += w + "\x00"
	}

	return msg
}

type ipcConn struct {
	socketAddr *net.UnixAddr
	socketConn *net.UnixConn
}

func newIPCConn(addressPath string) (ipcConn, error) {
	// TODO: Try using monkey-patching to facilitate unit testing for this: var resolveAddr = func() {//...} and then replacing that in the test file.
	addr, err := net.ResolveUnixAddr("unixgram", addressPath)
	if err != nil {
		return ipcConn{}, fmt.Errorf("failed to resolve unix address: %v", err)
	}

	// TODO: For this line too
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return ipcConn{}, fmt.Errorf("%w: %v", errInvalidUnixSocket, err)
	}

	return ipcConn{
		socketAddr: addr,
		socketConn: conn,
	}, nil
}

func (ipc ipcConn) Send(cmd ipcCommand) error {
	// TODO: For this line too
	if _, err := ipc.socketConn.Write([]byte(cmd.intoMessage())); err != nil {
		if errors.Is(err, syscall.EPIPE) {
			// bspwm has closed the socket connection
			return errClosedUnixSocket
		}

		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

func (ipc ipcConn) Receive() ([]byte, error) {
	const maxBufferSize = 512

	var msg []byte
	for buffer := make([]byte, maxBufferSize); ; buffer = make([]byte, maxBufferSize) {
		// TODO: For this line too
		_, _, err := ipc.socketConn.ReadFromUnix(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("failed to receive response: %v", err)
		}

		msg = append(msg, buffer...)
	}

	if len(msg) == 0 {
		return nil, errors.New("response was empty")
	}

	return bytes.Trim(msg, "\x00"), nil
}

func (ipc ipcConn) ReceiveAsync() (chan []byte, chan error) {
	var (
		resCh = make(chan []byte)
		errCh = make(chan error, 1)
	)

	const maxBufferSize = 512

	go func(resCh chan []byte, errCh chan error) {
		for buffer := make([]byte, maxBufferSize); ; buffer = make([]byte, maxBufferSize) {
			_, _, err := ipc.socketConn.ReadFromUnix(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				errCh <- fmt.Errorf("failed to receive response: %v", err)
				break
			}

			if len(buffer) == 0 {
				errCh <- errors.New("response was empty")
				break
			}

			resCh <- bytes.Trim(buffer, "\x00")
		}
	}(resCh, errCh)

	return resCh, errCh
}

func (ipc ipcConn) Close() error {
	return ipc.socketConn.Close()
}
