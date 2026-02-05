package retry

import (
	"errors"
	"time"

	"github.com/runtime-radar/runtime-radar/lib/errcommon"
)

// Config represents configuration of retrying function.
// For example, &Config{MaxAttempts: 5, ShouldRetry: func(error) { return true }, Delay: 1*time.Minute} will try
// to execute function 5 times once in a minute independently of kind of error returned.
type Config struct {
	MaxAttempts int
	ShouldRetry func(error) bool
	Delay       time.Duration
}

// NewDefaultConfig returns *Config that says to do at most one attempt before returning an error.
// ShouldRetry is configured so it always returns true. Delay is set to 100ms.
func NewDefaultConfig() *Config {
	return &Config{
		1,
		func(error) bool { return true },
		time.Millisecond * 100,
	}
}

// Do executes f and in case of an error retries it depending on configuration.
// If config is not passed it uses default one.
func Do(f func() error, c *Config) error {
	if f == nil {
		return errors.New("nil func given")
	}

	if c == nil {
		c = NewDefaultConfig()
	}

	var errs []error

	for i := 0; i < c.MaxAttempts; i++ {
		err := f()
		if err == nil {
			return nil
		}

		errs = append(errs, err)

		if !c.ShouldRetry(err) || i == c.MaxAttempts-1 { // last attempt - no need to sleep
			break
		}

		time.Sleep(c.Delay)
	}

	return errcommon.CollectErrors("retry.Do", errs)
}
