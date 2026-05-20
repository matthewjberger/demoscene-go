package render

import "rendergraph-go/transform"

// Input is the engine's per-frame input snapshot. Stored as a typed
// resource on the engine world; platform glue (GLFW callbacks on
// desktop, DOM listeners on wasm) writes into it once per frame.
//
// The fields are the minimum a camera controller needs. Keep this
// flat and value-typed so the platform layer never has to think about
// allocation.
type Input struct {
	MousePosition transform.Vec2
	MouseDelta    transform.Vec2
	Wheel         float32

	LeftDown   bool
	RightDown  bool
	MiddleDown bool
}

// BeginFrame zeroes the per-frame deltas (mouse movement, wheel
// scroll) so the input state at the start of each frame is "no input
// this frame yet." Held buttons stay held.
func (i *Input) BeginFrame() {
	i.MouseDelta = transform.Vec2{0, 0}
	i.Wheel = 0
}
