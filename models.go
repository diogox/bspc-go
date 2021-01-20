package bspc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (
	// ID represents any bspwm ID type.
	ID uint

	// State contains the structure for the whole state of a bspwm instance.
	State struct {
		FocusedMonitorID ID        `json:"focusedMonitorId"`
		PrimaryMonitorID ID        `json:"primaryMonitorId"`
		ClientsCount     int       `json:"clientsCount"`
		Monitors         []Monitor `json:"monitors"`
		FocusHistory     []struct {
			MonitorID ID `json:"monitorId"`
			DesktopID ID `json:"desktopId"`
			NodeID    ID `json:"nodeId"`
		} `json:"focusHistory"`
		StackedNodesList []ID `json:"stackingList"`
	}

	// Monitor contains all the info regarding a given monitor.
	Monitor struct {
		// RandRID is the monitor's ID in the RandR tool.
		RandRID          int       `json:"randrId"`
		ID               ID        `json:"id"`
		Name             string    `json:"name"`
		Wired            bool      `json:"wired"`
		StickyCount      int       `json:"stickyCount"`
		WindowGap        int       `json:"windowGap"`
		BorderWidth      int       `json:"borderWidth"`
		FocusedDesktopID ID        `json:"focusedDesktopId"`
		Padding          padding   `json:"padding"`
		Rectangle        rectangle `json:"rectangle"`
		Desktops         []Desktop `json:"desktops"`
	}

	// Desktop contains all the info regarding a given desktop.
	Desktop struct {
		Name          string     `json:"name"`
		ID            ID         `json:"id"`
		Layout        LayoutType `json:"layout"`
		UserLayout    LayoutType `json:"userLayout"`
		WindowGap     int        `json:"windowGap"`
		BorderWidth   int        `json:"borderWidth"`
		FocusedNodeID ID         `json:"focusedNodeId"`
		Padding       padding    `json:"padding"`
		Root          Node       `json:"root"`
	}

	// Node contains all the info regarding a given node.
	Node struct {
		ID         ID        `json:"id"`
		SplitType  SplitType `json:"SplitType"`
		SplitRatio float64   `json:"splitRatio"`
		Vacant     bool      `json:"vacant"`
		Hidden     bool      `json:"hidden"`
		Sticky     bool      `json:"sticky"`
		Private    bool      `json:"private"`
		Locked     bool      `json:"locked"`
		Marked     bool      `json:"marked"`
		Preselect  *struct {
			SplitDirection DirectionType `json:"splitDir"`
			SplitRatio     float64       `json:"splitRatio"`
		} `json:"presel"`
		Rectangle   rectangle   `json:"rectangle"`
		Constraints constraints `json:"constraints"`
		FirstChild  *Node       `json:"firstChild"`
		SecondChild *Node       `json:"secondChild"`
		Client      struct {
			ClassName         string    `json:"className"`
			InstanceName      string    `json:"instanceName"`
			BorderWidth       int       `json:"borderWidth"`
			State             string    `json:"state"`
			LastState         string    `json:"lastState"`
			Layer             string    `json:"layer"`
			LastLayer         string    `json:"lastLayer"`
			Urgent            bool      `json:"urgent"`
			Shown             bool      `json:"shown"`
			TiledRectangle    rectangle `json:"tiledRectangle"`
			FloatingRectangle rectangle `json:"floatingRectangle"`
		} `json:"client"`
	}

	// Event holds the event type and a payload that can be type-cast into the correct event-type model.
	Event struct {
		Type EventType

		// Payload needs to be type-cast into an event struct, according to the event type above.
		Payload interface{}
	}

	padding struct {
		Top    int `json:"top"`
		Right  int `json:"right"`
		Bottom int `json:"bottom"`
		Left   int `json:"left"`
	}

	rectangle struct {
		X      int `json:"x"`
		Y      int `json:"Y"`
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	constraints struct {
		MinWidth  int `json:"min_width"`
		MinHeight int `json:"min_height"`
	}
)

func hexToID(hex string) (ID, error) {
	id, err := strconv.ParseUint(strings.Replace(hex, "x0", "", 1), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("failed to parse hex to ID: %v", err)
	}

	return ID(id), nil
}

func geometryToRectangle(geometry string) (rectangle, error) {
	geometryParts := strings.Split(geometry, "+")
	if len(geometryParts) != 3 {
		return rectangle{}, errors.New("not enough fields for monitor geometry")
	}

	geometryResolution := strings.Split(geometryParts[0], "x")
	if len(geometryParts) != 2 {
		return rectangle{}, errors.New("not enough fields for monitor geometry resolution")
	}

	geometryX, err := strconv.Atoi(geometryParts[1])
	if err != nil {
		return rectangle{}, fmt.Errorf("monitor geometry X not a number: %v", err)
	}

	geometryY, err := strconv.Atoi(geometryParts[2])
	if err != nil {
		return rectangle{}, fmt.Errorf("monitor geometry Y not a number: %v", err)
	}

	geometryWidth, err := strconv.Atoi(geometryResolution[0])
	if err != nil {
		return rectangle{}, fmt.Errorf("monitor geometry width not a number: %v", err)
	}

	geometryHeight, err := strconv.Atoi(geometryResolution[1])
	if err != nil {
		return rectangle{}, fmt.Errorf("monitor geometry height not a number: %v", err)
	}

	return rectangle{
		X:      geometryX,
		Y:      geometryY,
		Width:  geometryWidth,
		Height: geometryHeight,
	}, nil
}
