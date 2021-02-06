package bspctest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/diogox/bspc-go"
)

// This is a helper for when working with gomock.
// Example usage:
//
// mockClient.EXPECT().
// 	Query("wm --dump-state", bspctest.QueryResponse(t, bspc.State{PrimaryMonitorID: bspc.ID(2)} )).
// 	Return(nil)
//
// Whatever you pass into the second argument, will be the value the mock will use to populate the pointer in the code.

type Matcher struct {
	t   *testing.T
	res interface{}
}

// QueryResponse sets the mocked response to be returned in the Query method's passed in resolver.
func QueryResponse(t *testing.T, res interface{}) *Matcher {
	return &Matcher{
		t:   t,
		res: res,
	}
}

func (m *Matcher) String() string {
	bb, err := json.Marshal(m.res)
	require.NoError(m.t, err)

	return string(bb)
}

func (m *Matcher) Matches(x interface{}) bool {
	resolver, ok := x.(bspc.QueryResponseResolver)
	if !ok {
		return false
	}

	bb, err := json.Marshal(m.res)
	require.NoError(m.t, err)

	// Running this populates the variable passed into it by reference.
	err = resolver(bb)
	require.NoError(m.t, err)

	return true
}
