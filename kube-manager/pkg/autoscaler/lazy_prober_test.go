package autoscaler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type closableProber struct {
	*stubProber

	closed int
}

func (p *closableProber) Close() error {
	p.closed++

	return nil
}

func TestLazyProberDialsOnce(t *testing.T) {
	inner := &closableProber{stubProber: newStubProber(proberStep{Backlog: 7, Consumers: 1})}

	var dials int

	p := &LazyProber{Dial: func() (ProberCloser, error) {
		dials++

		return inner, nil
	}}

	for range 3 {
		backlog, consumers, err := p.InspectQueue(t.Context())
		require.NoError(t, err)
		require.Equal(t, 7, backlog)
		require.Equal(t, 1, consumers)
	}

	require.Equal(t, 1, dials)
	require.Equal(t, 3, inner.calls())
}

// A broker down while the service starts is retried on every sample instead of being fatal.
func TestLazyProberRetriesFailedDial(t *testing.T) {
	dialErr := errors.New("connection refused")
	inner := &closableProber{stubProber: newStubProber(proberStep{Backlog: 3, Consumers: 2})}

	var dials int

	p := &LazyProber{Dial: func() (ProberCloser, error) {
		dials++
		if dials < 3 {
			return nil, dialErr
		}

		return inner, nil
	}}

	for range 2 {
		_, _, err := p.InspectQueue(t.Context())
		require.ErrorIs(t, err, dialErr)
	}

	backlog, consumers, err := p.InspectQueue(t.Context())
	require.NoError(t, err)
	require.Equal(t, 3, backlog)
	require.Equal(t, 2, consumers)
	require.Equal(t, 3, dials)
}

func TestLazyProberClose(t *testing.T) {
	inner := &closableProber{stubProber: newStubProber(proberStep{})}

	p := &LazyProber{Dial: func() (ProberCloser, error) { return inner, nil }}

	// Never dialed: nothing to close, and closing must not dial.
	require.NoError(t, p.Close())
	require.Zero(t, inner.closed)

	_, _, err := p.InspectQueue(t.Context())
	require.NoError(t, err)

	require.NoError(t, p.Close())
	require.Equal(t, 1, inner.closed)

	require.NoError(t, p.Close())
	require.Equal(t, 1, inner.closed)
}
