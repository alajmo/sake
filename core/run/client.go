package run

import (
	"io"
	"os"
	"sync"
)

type Client interface {
	Connect(bool, string, *sync.Mutex) *ErrConnect
	Run([]string, string) error
	Wait() error
	Close() error
	Prefix() string
	Write(p []byte) (n int, err error)
	WriteClose() error
	Stdin() io.WriteCloser
	Stderr() io.Reader
	Stdout() io.Reader
	Signal(os.Signal) error
	GetHost() string
}

type ErrConnect struct {
	Name   string
	Host   string
	User   string
	Port   uint8
	Reason string
}
