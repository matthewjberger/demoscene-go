package render

import "github.com/go-gl/mathgl/mgl32"

// ndcZTo01 remaps OpenGL clip-space z [-w, w] to wgpu clip-space z [0, w].
var ndcZTo01 = mgl32.Mat4{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 0.5, 0,
	0, 0, 0.5, 1,
}

// PerspectiveZO returns mgl32's right-handed perspective composed with a
// depth-range remap from OpenGL's [-1, 1] to wgpu's [0, 1].
func PerspectiveZO(fovY, aspect, near, far float32) mgl32.Mat4 {
	return ndcZTo01.Mul4(mgl32.Perspective(fovY, aspect, near, far))
}
