package bspc

type (
	LayoutType             string
	SplitType              string
	DirectionType          string
	StateType              string
	FlagType               string
	RelativePositionType   string
	LayerType              string
	PointerActionType      string
	PointerActionStateType string
)

const (
	LayoutTypeTiled   LayoutType = "tiled"
	LayoutTypeMonocle LayoutType = "monocle"

	SplitTypeHorizontal SplitType = "horizontal"
	SplitTypeVertical   SplitType = "vertical"

	DirectionTypeUp    DirectionType = "north"
	DirectionTypeDown  DirectionType = "south"
	DirectionTypeLeft  DirectionType = "west"
	DirectionTypeRight DirectionType = "east"

	StateTypeTiled       StateType = "tiled"
	StateTypePseudoTiled StateType = "pseudo_tiled"
	StateTypeFloating    StateType = "floating"
	StateTypeFullscreen  StateType = "fullscreen"

	FlagTypeHidden  FlagType = "hidden"
	FlagTypeSticky  FlagType = "sticky"
	FlagTypePrivate FlagType = "private"
	FlagTypeLocked  FlagType = "locked"
	FlagTypeMarked  FlagType = "marked"
	FlagTypeUrgent  FlagType = "urgent"

	RelativePositionTypeAbove RelativePositionType = "above"
	RelativePositionTypeBelow RelativePositionType = "below"

	LayerTypeAbove  LayerType = "above"
	LayerTypeNormal LayerType = "normal"
	LayerTypeBelow  LayerType = "below"

	PointerActionTypeMove         PointerActionType = "move"
	PointerActionTypeResizeCorner PointerActionType = "resize_corner"
	PointerActionTypeResizeSide   PointerActionType = "resize_side"

	PointerActionStateTypeBegin PointerActionStateType = "begin"
	PointerActionStateTypeEnd   PointerActionStateType = "end"
)

func (lt LayoutType) IsValid() bool {
	return lt == LayoutTypeTiled ||
		lt == LayoutTypeMonocle
}

func (st SplitType) IsValid() bool {
	return st == SplitTypeHorizontal ||
		st == SplitTypeVertical
}

func (dt DirectionType) IsValid() bool {
	return dt == DirectionTypeUp ||
		dt == DirectionTypeDown ||
		dt == DirectionTypeLeft ||
		dt == DirectionTypeRight
}

func (st StateType) IsValid() bool {
	return st == StateTypeTiled ||
		st == StateTypePseudoTiled ||
		st == StateTypeFloating ||
		st == StateTypeFullscreen
}

func (ft FlagType) IsValid() bool {
	return ft == FlagTypeHidden ||
		ft == FlagTypeLocked ||
		ft == FlagTypeMarked ||
		ft == FlagTypePrivate ||
		ft == FlagTypeSticky ||
		ft == FlagTypeUrgent
}

func (rpt RelativePositionType) IsValid() bool {
	return rpt == RelativePositionTypeAbove ||
		rpt == RelativePositionTypeBelow
}

func (lt LayerType) IsValid() bool {
	return lt == LayerTypeAbove ||
		lt == LayerTypeBelow ||
		lt == LayerTypeNormal
}

func (pat PointerActionType) IsValid() bool {
	return pat == PointerActionTypeMove ||
		pat == PointerActionTypeResizeCorner ||
		pat == PointerActionTypeResizeSide
}

func (past PointerActionStateType) IsValid() bool {
	return past == PointerActionStateTypeBegin ||
		past == PointerActionStateTypeEnd
}
