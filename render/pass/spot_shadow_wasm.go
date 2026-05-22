//go:build js

package pass

import (
	"indigo/render"
)

// spotShadowExecute is a no-op on wasm: the cogentcore/webgpu JS
// binding doesn't expose SetViewport / SetScissorRect, which the
// atlas approach relies on. Spot shadows are disabled in the
// browser until either the binding exposes the viewport API or
// the pass is rewritten to use one-texture-per-spot.
func spotShadowExecute(_ any, _ *render.PassContext) error {
	return nil
}
