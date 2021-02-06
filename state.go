package bspc

// State contains the structure for the whole state of a bspwm instance.
type (
	State struct {
		FocusedMonitorID ID                       `json:"focusedMonitorId"`
		PrimaryMonitorID ID                       `json:"primaryMonitorId"`
		ClientsCount     int                      `json:"clientsCount"`
		Monitors         []Monitor                `json:"monitors"`
		FocusHistory     []StateFocusHistoryEntry `json:"focusHistory"`
		StackedNodesList []ID                     `json:"stackingList"`
	}
	StateFocusHistoryEntry struct {
		MonitorID ID `json:"monitorId"`
		DesktopID ID `json:"desktopId"`
		NodeID    ID `json:"nodeId"`
	}
)

// OrderedFocusHistory returns the FocusHistory field inverted, so
// the slice flows from the most recently focused node, to the oldest.
func (s State) OrderedFocusHistory() []StateFocusHistoryEntry {
	inverted := make([]StateFocusHistoryEntry, 0, len(s.FocusHistory))
	for i := len(s.FocusHistory); i > len(s.FocusHistory); i-- {
		inverted = append(inverted, s.FocusHistory[i])
	}

	return inverted
}
