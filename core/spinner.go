// Singleton module
package core

import (
	"io"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/theckman/yacspin"
)

var (
	buildMode = "prod"
)

var lock = &sync.Mutex{}

type Loader struct {
	cfg      *yacspin.Config
	spinner  *yacspin.Spinner
	disabled bool
}

var spinner *Loader

func GetSpinner() *Loader {
	if spinner == nil {
		lock.Lock()
		defer lock.Unlock()
		if spinner == nil {
			// NOTE: Don't print the spinner in tests since it causes
			// golden files to produce different results.
			var cfg yacspin.Config
			if buildMode == "TEST" {
				cfg = yacspin.Config{
					Frequency:       100 * time.Millisecond,
					CharSet:         yacspin.CharSets[9],
					SuffixAutoColon: false,
					Writer:          io.Discard,
				}
			} else {
				cfg = yacspin.Config{
					Frequency:       100 * time.Millisecond,
					CharSet:         yacspin.CharSets[9],
					SuffixAutoColon: false,
					ShowCursor:      true,
				}
			}

			spn, err := yacspin.New(cfg)
			if err != nil {
				return nil
			}

			spinner = &Loader{
				cfg:      &cfg,
				spinner:  spn,
				disabled: false,
			}
		}
	}

	return spinner
}

func (s *Loader) Start(msg string, delay time.Duration) {
	var err error

	s.Enable()
	go func() {
		time.Sleep(delay * time.Millisecond)
		if !s.disabled {
			s.spinner.Message(msg)
			err = s.spinner.Start()
			if err != nil {
				return
			}
		}
	}()

	// In-case user interrupts, make sure spinner is stopped
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan

		if spinner != nil && err == nil {
			_ = s.spinner.Stop()
		}
		os.Exit(0)
	}()
}

func (s *Loader) Stop() {
	err := s.spinner.Stop()
	if err != nil {
		return
	}
}

func (s *Loader) Enable() {
	s.disabled = false
}
