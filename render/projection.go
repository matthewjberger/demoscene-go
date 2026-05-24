package render

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// PerspectiveZO builds a reverse-Z perspective matrix for a 0..1 clip-space
// depth range (wgpu): the near plane maps to 1 and the far plane to 0. Reverse-Z
// distributes depth precision far more evenly, so the depth buffer clears to 0
// and depth tests use Greater/GreaterEqual. Mirrors nightshade's
// reverse_z_perspective.
func PerspectiveZO(fovY, aspect, near, far float32) mgl32.Mat4 {
	f := float32(1.0 / math.Tan(float64(fovY)/2.0))
	depth := far - near
	// Column-major: each group of four values is a column.
	return mgl32.Mat4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, near / depth, -1,
		0, 0, near * far / depth, 0,
	}
}
