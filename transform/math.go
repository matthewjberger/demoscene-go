package transform

import "github.com/go-gl/mathgl/mgl32"

// Aliases over mgl32 so callers (and the engine) stay uniform on one
// math library. These are package-level so component fields can use
// them directly without mgl32 imports.
type (
	Vec2 = mgl32.Vec2
	Vec3 = mgl32.Vec3
	Mat4 = mgl32.Mat4
	Quat = mgl32.Quat
)

// QuatIdentity returns the identity rotation.
func QuatIdentity() Quat { return mgl32.QuatIdent() }

// QuatFromAxisAngle returns a unit quaternion rotating by angle
// (radians) around the given axis.
func QuatFromAxisAngle(angle float32, axis Vec3) Quat {
	return mgl32.QuatRotate(angle, axis)
}

// QuatNormalize returns a unit-length copy of q.
func QuatNormalize(q Quat) Quat { return q.Normalize() }

// QuatToMat4 converts a quaternion to a 4x4 rotation matrix.
func QuatToMat4(q Quat) Mat4 { return q.Mat4() }

// Mat4Identity returns the identity matrix.
func Mat4Identity() Mat4 { return mgl32.Ident4() }
