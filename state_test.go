package bspc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/diogox/bspc-go"
)

func TestState_OrderedFocusHistory(t *testing.T) {
	t.Run("should return the inverted focus history", func(t *testing.T) {
		someID := bspc.ID(999)

		s := bspc.State{
			FocusHistory: []bspc.StateFocusHistoryEntry{
				{MonitorID: someID, DesktopID: someID, NodeID: bspc.ID(2)},
				{MonitorID: someID, DesktopID: someID, NodeID: bspc.ID(1)},
				{MonitorID: someID, DesktopID: someID, NodeID: bspc.ID(0)},
			},
		}

		ordered := s.OrderedFocusHistory()
		for want, got := range ordered {
			assert.Equal(t, want, got.NodeID)
		}
	})
}
