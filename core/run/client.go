package run

import (
	"io"
	"os"
	"sync"
)

type Client interface {
	Connect(SSHDialFunc, bool, string, uint, *sync.Mutex) *ErrConnect
	Run(int, []string, string, string, string) error
	Wait(int) error
	Close(int) error
	Write(int, []byte) (n int, err error)
	WriteClose(int) error
	Stdin(int) io.WriteCloser
	Stderr(int) io.Reader
	Stdout(int) io.Reader
	Signal(int, os.Signal) error
	GetName() string
	Prefix() (string, string, string, uint16)
	Connected() bool
}

type ErrConnect struct {
	Name   string
	Host   string
	User   string
	Port   uint16
	Reason string
}

func (e *ErrConnect) Error() string {
	return e.Reason
}
