// Package window holds the engine's per-frame timing and viewport
// state as plain data. Platform glue (GLFW on desktop, DOM listeners
// on wasm) writes into a [Window] resource stored on the ECS world;
// systems read from the same resource.
//
// The platform layer owns the actual OS window and feeds this
// resource each frame via [Advance].
package window

// ViewportSize is the surface's pixel dimensions. Updated on resize.
type ViewportSize struct {
	Width  uint32
	Height uint32
}

// Timing captures per-frame timing data. Updated by [Advance] from
// the platform main loop's delta.
type Timing struct {
	DeltaSeconds  float32
	UptimeSeconds float32
	FramesPerSec  float32
	FrameCounter  uint64
}

// Window is the engine's window resource. Installed via
// [ecs.SetResource] on whichever worlds need to read viewport and
// timing.
type Window struct {
	Viewport ViewportSize
	Timing   Timing
}

// Advance updates timing by delta (seconds). Computes a smoothed FPS
// estimate via a 1/delta running snapshot. Frame-counter increments
// monotonically.
func Advance(timing *Timing, delta float32) {
	timing.DeltaSeconds = delta
	timing.UptimeSeconds += delta
	timing.FrameCounter++
	if delta > 0 {
		timing.FramesPerSec = 1.0 / delta
	}
}
