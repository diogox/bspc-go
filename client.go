package bspc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type (
	Logger interface {
		Info(msg string)
		Warn(msg string)
	}

	Client interface {
		Query(rawCmd string, resResolver QueryResponseResolver) error
		// subscribe(rawEvents string) (chan Event, chan error, error) // TODO: Remove this, or make it public again
		SubscribeEvents(event EventType, events ...EventType) (chan Event, chan error, error)
	}

	// client holds the socket path, because it needs to initialize a socket connection on each method call.
	// Reusing the same connection across multiple calls is not reliable, because bspwm will sometimes close
	// connections after it responds.
	client struct {
		socketPath string
		logger     Logger
	}
)

// New returns a client instance with the first unix socket path it finds
// with a name matching: /tmp/bspwm<host_name>_<display_number>_<screen_number>-socket
// If the value passed in as a logger is nil, logging will be disabled.
func New(logger Logger) (Client, error) {
	errSocketFound := errors.New("socket has been found")

	regex, err := regexp.Compile(`^/tmp/\w+_\d+_\d+-socket$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile bspwm socket name regex: %v", err)
	}

	var socketPath string
	err = filepath.Walk(filepath.Dir("/tmp/"), func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if regex.MatchString(path) {
			socketPath = path
			return errSocketFound
		}

		return nil
	})
	if err != nil && !errors.Is(err, errSocketFound) {
		return nil, fmt.Errorf("failed to find bspwm unix socket: %v", err)
	}

	return NewWithSocketPath(socketPath, logger)
}

// NewWithSocketPath returns a client instance with the given UNIX socket path.
// If the value passed in as a logger is nil, logging will be disabled.
func NewWithSocketPath(path string, logger Logger) (Client, error) {
	if _, err := newUnixSocketAddress(path); err != nil {
		return nil, err
	}

	return client{
		socketPath: path,
		logger:     logger,
	}, nil
}

// Query takes in a "raw" string bpsc command (without the "bspc" prefix), and populates its
// response into the provided type. The models provided in this package can be used to construct
// the response type.
func (c client) Query(rawCmd string, resResolver QueryResponseResolver) error {
	// TODO: How can I return a sentinel error if the command is invalid? Need to read errors from the socket.
	c.logger.Info(fmt.Sprintf("using socket at path %s", c.socketPath)) // TODO: The logger needs to be optional, need to nil-check

	socketAddr, err := newUnixSocketAddress(c.socketPath)
	if err != nil {
		return err
	}

	ipc, err := newIPCConn(socketAddr)
	if err != nil {
		return fmt.Errorf("failed to initialize socket connection: %w", err)
	}

	if err := ipc.Send(ipcCommand(rawCmd)); err != nil {
		return err
	}

	resBytes, err := ipc.Receive()
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	if resResolver == nil {
		return nil
	}

	if err := resResolver(resBytes); err != nil {
		return fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return nil
}

// subscribe returns two channels: one for the events published by bspwm
// and that we subscribe to, and one for the errors that might occur during
// the subscription.
// It is a private method, because passing in more than one event can cause errors,
// as bspwm will sometimes jumble to events together with no clear delimiters that I could identify. TODO
// (for eg. when you enable monocle mode, `desktop_layout` and `node_remove` events will often be
// mixed in the same string, with no delimiters between the end of one event, and the beginning of another).
func (c client) subscribe(rawEvents string) (chan Event, chan error, error) {
	c.logger.Info(fmt.Sprintf("using socket at path %s", c.socketPath))

	socketAddr, err := newUnixSocketAddress(c.socketPath)
	if err != nil {
		return nil, nil, err
	}

	ipc, err := newIPCConn(socketAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize socket connection: %w", err)
	}

	const subscribeCmd = "subscribe"

	if err := ipc.Send(ipcCommand(subscribeCmd + " " + rawEvents)); err != nil {
		return nil, nil, err
	}

	resCh, errCh := ipc.ReceiveAsync()

	eventCh := make(chan Event)
	go func(resCh chan []byte) {
		for res := range resCh {
			parts := strings.Split(strings.ReplaceAll(string(res), "\n", ""), " ")
			if len(parts) < 2 {
				c.logEventWarning("unknown", "not enough fields")
				continue
			}

			ev := Event{
				Type: EventType(parts[0]),
			}

			parts = parts[1:]

			switch ev.Type {
			case EventTypeMonitorAdd:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				id, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				geometry, err := geometryToRectangle(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, err.Error())
					continue
				}

				ev.Payload = EventMonitorAdd{
					MonitorID:       id,
					MonitorName:     parts[1],
					MonitorGeometry: geometry,
				}
			case EventTypeMonitorRename:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				id, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				ev.Payload = EventMonitorRename{
					MonitorID:      id,
					MonitorOldName: parts[1],
					MonitorNewName: parts[2],
				}
			case EventTypeMonitorRemove:
				if len(parts) != 1 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				id, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				ev.Payload = EventMonitorRemove{
					MonitorID: id,
				}
			case EventTypeMonitorSwap:
				if len(parts) != 2 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				srcID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dstID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				ev.Payload = EventMonitorSwap{
					SourceMonitorID:      srcID,
					DestinationMonitorID: dstID,
				}
			case EventTypeMonitorFocus:
				if len(parts) != 1 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				id, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				ev.Payload = EventMonitorFocus{
					MonitorID: id,
				}

			case EventTypeMonitorGeometry:
				if len(parts) != 2 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				id, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				geometry, err := geometryToRectangle(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, err.Error())
					continue
				}

				ev.Payload = EventMonitorGeometry{
					MonitorID:       id,
					MonitorGeometry: geometry,
				}
			case EventTypeDesktopAdd:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				ev.Payload = EventDesktopAdd{
					MonitorID:   mID,
					DesktopID:   dID,
					DesktopName: parts[2],
				}
			case EventTypeDesktopRename:
				if len(parts) != 4 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				ev.Payload = EventDesktopRename{
					MonitorID:      mID,
					DesktopID:      dID,
					DesktopOldName: parts[2],
					DesktopNewName: parts[3],
				}
			case EventTypeDesktopRemove:
				if len(parts) != 2 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				ev.Payload = EventDesktopRemove{
					MonitorID: mID,
					DesktopID: dID,
				}
			case EventTypeDesktopSwap, EventTypeDesktopTransfer:
				if len(parts) != 4 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				srcMonitorID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				srcDesktopID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				dstMonitorID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				dstDesktopID, err := hexToID(parts[3])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[3]))
					continue
				}

				switch ev.Type {
				case EventTypeDesktopSwap:
					ev.Payload = EventDesktopSwap{
						SourceMonitorID:      srcMonitorID,
						SourceDesktopID:      srcDesktopID,
						DestinationMonitorID: dstMonitorID,
						DestinationDesktopID: dstDesktopID,
					}
				case EventTypeDesktopTransfer:
					ev.Payload = EventDesktopTransfer{
						SourceMonitorID:      srcMonitorID,
						SourceDesktopID:      srcDesktopID,
						DestinationMonitorID: dstMonitorID,
						DestinationDesktopID: dstDesktopID,
					}
				}
			case EventTypeDesktopFocus, EventTypeDesktopActivate:
				if len(parts) != 2 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				switch ev.Type {
				case EventTypeDesktopFocus:
					ev.Payload = EventDesktopFocus{
						MonitorID: mID,
						DesktopID: dID,
					}
				case EventTypeDesktopActivate:
					ev.Payload = EventDesktopActivate{
						MonitorID: mID,
						DesktopID: dID,
					}
				}
			case EventTypeDesktopLayout:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				desktopLayout := LayoutType(parts[2])
				if !desktopLayout.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid desktop layout %s", desktopLayout))
					continue
				}

				ev.Payload = EventDesktopLayout{
					MonitorID:     mID,
					DesktopID:     dID,
					DesktopLayout: desktopLayout,
				}
			case EventTypeNodeAdd:
				if len(parts) != 4 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				ipID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				nodeID, err := hexToID(parts[3])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[3]))
					continue
				}

				ev.Payload = EventNodeAdd{
					MonitorID: mID,
					DesktopID: dID,
					IPID:      ipID,
					NodeID:    nodeID,
				}
			case EventTypeNodeRemove:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nodeID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				ev.Payload = EventNodeRemove{
					MonitorID: mID,
					DesktopID: dID,
					NodeID:    nodeID,
				}
			case EventTypeNodeSwap, EventTypeNodeTransfer:
				if len(parts) != 6 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}
				srcMonitorID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				srcDesktopID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				srcNodeID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				dstMonitorID, err := hexToID(parts[3])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[3]))
					continue
				}

				dstDesktopID, err := hexToID(parts[4])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[4]))
					continue
				}

				dstNodeID, err := hexToID(parts[5])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[5]))
					continue
				}

				switch ev.Type {
				case EventTypeNodeSwap:
					ev.Payload = EventNodeSwap{
						SourceMonitorID:      srcMonitorID,
						SourceDesktopID:      srcDesktopID,
						SourceNodeID:         srcNodeID,
						DestinationMonitorID: dstMonitorID,
						DestinationDesktopID: dstDesktopID,
						DestinationNodeID:    dstNodeID,
					}
				case EventTypeNodeTransfer:
					ev.Payload = EventNodeTransfer{
						SourceMonitorID:      srcMonitorID,
						SourceDesktopID:      srcDesktopID,
						SourceNodeID:         srcNodeID,
						DestinationMonitorID: dstMonitorID,
						DestinationDesktopID: dstDesktopID,
						DestinationNodeID:    dstNodeID,
					}
				}
			case EventTypeNodeFocus, EventTypeNodeActivate:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				switch ev.Type {
				case EventTypeNodeFocus:
					ev.Payload = EventNodeFocus{
						MonitorID: mID,
						DesktopID: dID,
						NodeID:    nID,
					}
				case EventTypeNodeActivate:
					ev.Payload = EventNodeActivate{
						MonitorID: mID,
						DesktopID: dID,
						NodeID:    nID,
					}
				}
			case EventTypeNodePreselect:
				if len(parts) != 5 || len(parts) != 6 { // TODO: there is an optional field when `cancel` is present.
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				const (
					fieldCancel    = "cancel"
					fieldRatio     = "ratio"
					fieldDirection = "dir"
				)

				var (
					isCancel  *bool
					ratio     *float64
					direction *SplitType
				)

				switch parts[4] {
				case fieldCancel:
					cancel := true
					isCancel = &cancel
				case fieldRatio:
					r, err := strconv.ParseFloat(parts[5], 64)
					if err != nil {
						c.logEventWarning(ev.Type, "not enough fields")
						continue
					}
					ratio = &r
				case fieldDirection:
					d := SplitType(parts[5])
					if !d.IsValid() {
						c.logEventWarning(ev.Type, fmt.Sprintf("invalid split type: %s", d))
						continue
					}
					direction = &d
				default:
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid field '%s'", parts[4]))
					continue
				}

				ev.Payload = EventNodePreselect{
					MonitorID:      mID,
					DesktopID:      dID,
					NodeID:         nID,
					SplitDirection: direction,
					SplitRatio:     ratio,
					IsCancel:       isCancel,
				}
			case EventTypeNodeStack:
				if len(parts) != 3 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				n1ID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				relativePosition := RelativePositionType(parts[1])
				if !relativePosition.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid relative position %s", relativePosition))
					continue
				}

				n2ID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				ev.Payload = EventNodeStack{
					Node1ID:          n1ID,
					RelativePosition: relativePosition,
					Node2ID:          n2ID,
				}
			case EventTypeNodeGeometry:
				if len(parts) != 4 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				geometry, err := geometryToRectangle(parts[3])
				if err != nil {
					c.logEventWarning(ev.Type, err.Error())
					continue
				}

				ev.Payload = EventNodeGeometry{
					MonitorID:    mID,
					DesktopID:    dID,
					NodeID:       nID,
					NodeGeometry: geometry,
				}
			case EventTypeNodeState:
				if len(parts) != 5 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				state := StateType(parts[3])
				if !state.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid state %s", state))
					continue
				}

				const (
					enabledON  = "on"
					enabledOFF = "off"
				)

				var wasEnabled bool
				switch parts[4] {
				case enabledON:
					wasEnabled = true
				case enabledOFF:
					wasEnabled = false
				default:
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid field '%s'", parts[4]))
					continue
				}

				ev.Payload = EventNodeState{
					MonitorID:  mID,
					DesktopID:  dID,
					NodeID:     nID,
					State:      state,
					WasEnabled: wasEnabled,
				}
			case EventTypeNodeFlag:
				if len(parts) != 5 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				state := StateType(parts[3])
				if !state.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid state %s", state))
					continue
				}

				const (
					enabledON  = "on"
					enabledOFF = "off"
				)

				var wasEnabled bool
				switch parts[4] {
				case enabledON:
					wasEnabled = true
				case enabledOFF:
					wasEnabled = false
				default:
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid field '%s'", parts[4]))
					continue
				}

				ev.Payload = EventNodeState{
					MonitorID:  mID,
					DesktopID:  dID,
					NodeID:     nID,
					State:      state,
					WasEnabled: wasEnabled,
				}
			case EventTypeNodeLayer:
				if len(parts) != 4 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				layer := LayerType(parts[3])
				if !layer.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid layer %s", layer))
					continue
				}

				ev.Payload = EventNodeLayer{
					MonitorID: mID,
					DesktopID: dID,
					NodeID:    nID,
					Layer:     layer,
				}
			case EventTypePointerAction:
				if len(parts) != 5 {
					c.logEventWarning(ev.Type, "not enough fields")
					continue
				}

				mID, err := hexToID(parts[0])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[0]))
					continue
				}

				dID, err := hexToID(parts[1])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[1]))
					continue
				}

				nID, err := hexToID(parts[2])
				if err != nil {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid id %s", parts[2]))
					continue
				}

				pointerAction := PointerActionType(parts[3])
				if !pointerAction.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid pointer action %s", pointerAction))
					continue
				}

				pointerActionState := PointerActionStateType(parts[4])
				if !pointerActionState.IsValid() {
					c.logEventWarning(ev.Type, fmt.Sprintf("invalid pointer action state %s", pointerActionState))
					continue
				}

				ev.Payload = EventPointerAction{
					MonitorID:          mID,
					DesktopID:          dID,
					NodeID:             nID,
					PointerAction:      pointerAction,
					PointerActionState: pointerActionState,
				}
			default:
				c.logEventWarning(ev.Type, fmt.Sprintf("unsupported event: %s", res))
				continue
			}

			eventCh <- ev
		}
	}(resCh)

	return eventCh, errCh, nil
}

// SubscribeEvents takes in one or more of the available events in this package and calls Subscribe
// with the appropriate raw command. Take a look at Subscribe to know more.
// It currently uses a socket connection for each event as to avoid having different events jumbled together
// by bspwm. Haven't found a way to tell those "glued" events apart. TODO.
func (c client) SubscribeEvents(event EventType, moreEvents ...EventType) (chan Event, chan error, error) {
	// TODO: Refactor this code to be more maintainable.

	var (
		eventsChannel = make(chan Event)
		errorsChannel = make(chan error)
	)

	events := []EventType{event}
	events = append(events, moreEvents...)

	var (
		eventChs []chan Event
		errChs   []chan error
	)

	for _, ev := range events {
		eventCh, errCh, err := c.subscribe(string(ev))
		if err != nil {
			return nil, nil, err
		}

		eventChs = append(eventChs, eventCh)
		errChs = append(errChs, errCh)
	}

	for i := range eventChs {
		var (
			evCh  = eventChs[i]
			errCh = errChs[i]
		)

		go func() {
			for {
				select {
				case ev := <-evCh:
					eventsChannel <- ev
				case err := <-errCh:
					errorsChannel <- err
				}
			}
		}()
	}

	return eventsChannel, errorsChannel, nil
}

func (c client) logEventWarning(ev EventType, msg string) {
	if l := c.logger; l != nil {
		l.Warn(fmt.Sprintf(`"%s" event - %s`, ev, msg))
	}
}
