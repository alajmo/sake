package run

import (
	"io"
	"os"
	"sync"
)

type Client interface {
	Connect(bool, string, *sync.Mutex, SSHDialFunc) *ErrConnect
	Run([]string, string, string, string) error
	Wait() error
	Close() error
	Prefix() string
	Write(p []byte) (n int, err error)
	WriteClose() error
	Stdin() io.WriteCloser
	Stderr() io.Reader
	Stdout() io.Reader
	Signal(os.Signal) error
	GetName() string
}

type ErrConnect struct {
	Name   string
	Host   string
	User   string
	Port   uint16
	Reason string
}

func (e *ErrConnect) Error() string {
	return ""
}
