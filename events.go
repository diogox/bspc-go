package bspc

type EventType string

const (
	// TODO: Uncomment this when "report" events are supported.
	// EventTypeAll EventType = "all".

	// Monitor
	// "Please note that bspwm initializes monitors before
	// it reads messages on its socket, therefore the initial
	// monitor events can’t be received".
	EventTypeMonitorAdd      EventType = "monitor_add"
	EventTypeMonitorRename   EventType = "monitor_rename"
	EventTypeMonitorRemove   EventType = "monitor_remove"
	EventTypeMonitorSwap     EventType = "monitor_swap"
	EventTypeMonitorFocus    EventType = "monitor_focus"
	EventTypeMonitorGeometry EventType = "monitor_geometry"

	// Desktop.
	EventTypeDesktopAdd      EventType = "desktop_add"
	EventTypeDesktopRename   EventType = "desktop_rename"
	EventTypeDesktopRemove   EventType = "desktop_remove"
	EventTypeDesktopSwap     EventType = "desktop_swap"
	EventTypeDesktopTransfer EventType = "desktop_transfer"
	EventTypeDesktopFocus    EventType = "desktop_focus"
	EventTypeDesktopActivate EventType = "desktop_activate"
	EventTypeDesktopLayout   EventType = "desktop_layout"

	// Node.
	EventTypeNodeAdd       EventType = "node_add"
	EventTypeNodeRemove    EventType = "node_remove"
	EventTypeNodeSwap      EventType = "node_swap"
	EventTypeNodeTransfer  EventType = "node_transfer"
	EventTypeNodeFocus     EventType = "node_focus"
	EventTypeNodeActivate  EventType = "node_activate"
	EventTypeNodePreselect EventType = "node_presel"
	EventTypeNodeStack     EventType = "node_stack"
	EventTypeNodeGeometry  EventType = "node_geometry"
	EventTypeNodeState     EventType = "node_state"
	EventTypeNodeFlag      EventType = "node_flag"
	EventTypeNodeLayer     EventType = "node_layer"

	// Pointer.
	EventTypePointerAction EventType = "pointer_action"
)

type (
	// Monitor.
	EventMonitorAdd struct {
		MonitorID       ID
		MonitorName     string
		MonitorGeometry rectangle
	}
	EventMonitorRename struct {
		MonitorID      ID
		MonitorOldName string
		MonitorNewName string
	}
	EventMonitorRemove struct {
		MonitorID ID
	}
	EventMonitorSwap struct {
		SourceMonitorID      ID
		DestinationMonitorID ID
	}
	EventMonitorFocus struct {
		MonitorID ID
	}
	EventMonitorGeometry struct {
		MonitorID       ID
		MonitorGeometry rectangle
	}

	// Desktop.
	EventDesktopAdd struct {
		MonitorID   ID
		DesktopID   ID
		DesktopName string
	}
	EventDesktopRename struct {
		MonitorID      ID
		DesktopID      ID
		DesktopOldName string
		DesktopNewName string
	}
	EventDesktopRemove struct {
		MonitorID ID
		DesktopID ID
	}
	EventDesktopSwap struct {
		SourceMonitorID      ID
		SourceDesktopID      ID
		DestinationMonitorID ID
		DestinationDesktopID ID
	}
	EventDesktopTransfer struct {
		SourceMonitorID      ID
		SourceDesktopID      ID
		DestinationMonitorID ID
		DestinationDesktopID ID
	}
	EventDesktopFocus struct {
		MonitorID ID
		DesktopID ID
	}
	EventDesktopActivate struct {
		MonitorID ID
		DesktopID ID
	}
	EventDesktopLayout struct {
		MonitorID     ID
		DesktopID     ID
		DesktopLayout LayoutType
	}

	// Node.
	EventNodeAdd struct {
		MonitorID ID
		DesktopID ID
		IPID      ID // TODO: What is this?
		NodeID    ID
	}
	EventNodeRemove struct {
		MonitorID ID
		DesktopID ID
		NodeID    ID
	}
	EventNodeSwap struct {
		SourceMonitorID      ID
		SourceDesktopID      ID
		SourceNodeID         ID
		DestinationMonitorID ID
		DestinationDesktopID ID
		DestinationNodeID    ID
	}
	EventNodeTransfer struct {
		SourceMonitorID      ID
		SourceDesktopID      ID
		SourceNodeID         ID
		DestinationMonitorID ID
		DestinationDesktopID ID
		DestinationNodeID    ID
	}
	EventNodeFocus struct {
		MonitorID ID
		DesktopID ID
		NodeID    ID
	}
	EventNodeActivate struct {
		MonitorID ID
		DesktopID ID
		NodeID    ID
	}
	EventNodePreselect struct {
		MonitorID ID
		DesktopID ID
		NodeID    ID

		// Only one of the below will be available.
		SplitDirection *SplitType
		SplitRatio     *float64
		IsCancel       *bool
	}
	EventNodeStack struct {
		Node1ID          ID
		RelativePosition RelativePositionType
		Node2ID          ID
	}
	EventNodeGeometry struct {
		MonitorID    ID
		DesktopID    ID
		NodeID       ID
		NodeGeometry rectangle
	}
	EventNodeState struct {
		MonitorID  ID
		DesktopID  ID
		NodeID     ID
		State      StateType
		WasEnabled bool
	}
	EventNodeFlag struct {
		MonitorID ID
		DesktopID ID
		NodeID    ID
		Flag      FlagType

		// WasEnabled will be true, if the flag was enabled.
		// If the flag was disabled, it will be false.
		WasEnabled bool
	}
	EventNodeLayer struct {
		MonitorID ID
		DesktopID ID
		NodeID    ID
		Layer     LayerType
	}

	// Pointer.
	EventPointerAction struct {
		MonitorID          ID
		DesktopID          ID
		NodeID             ID
		PointerAction      PointerActionType
		PointerActionState PointerActionStateType
	}
)
