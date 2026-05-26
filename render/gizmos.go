package render

import "github.com/matthewjberger/indigo/transform"

type GizmoMode uint8

const (
	GizmoNone GizmoMode = iota
	GizmoTranslate
	GizmoRotate
	GizmoScale
)

type Gizmos struct {
	Mode GizmoMode

	HoverAxis  int
	HoverPlane int

	Dragging bool
	DragAxis uint8
	DragMode GizmoMode

	StartLocal          transform.LocalTransform
	StartGlobalTrans    [3]float32
	AxisWorldDirection  [3]float32
	AxisWorldLengthDrag float32
	InitialT            float32
	StartWorldVector    [3]float32

	DraggingPlane     bool
	DragPlaneAxis     uint8
	PlaneInitialTrans [3]float32
	PlaneNormalWorld  [3]float32
	PlaneInitialHit   [3]float32

	PrevLeftDown bool
}

func NewGizmos() *Gizmos {
	return &Gizmos{Mode: GizmoTranslate, HoverAxis: -1, HoverPlane: -1}
}
