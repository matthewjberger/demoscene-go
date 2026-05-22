package pass

import (
	"math"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

const gizmoAxisEpsilon = 1e-5

func vec3Approx(a, b mgl32.Vec3, eps float32) bool {
	d := a.Sub(b).Len()
	return d < eps
}

func TestLocalAxesIdentity(t *testing.T) {
	axes := LocalAxes(mgl32.Ident4())
	want := [3]mgl32.Vec3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for i := 0; i < 3; i++ {
		if !vec3Approx(axes[i], want[i], gizmoAxisEpsilon) {
			t.Errorf("axis %d = %v, want %v", i, axes[i], want[i])
		}
	}
}

func TestLocalAxesYaw90(t *testing.T) {
	// 90 deg yaw around +Y: X axis rotates to -Z, Z axis rotates to +X.
	yaw := mgl32.HomogRotate3DY(float32(math.Pi / 2))
	axes := LocalAxes(yaw)
	want := [3]mgl32.Vec3{{0, 0, -1}, {0, 1, 0}, {1, 0, 0}}
	for i := 0; i < 3; i++ {
		if !vec3Approx(axes[i], want[i], gizmoAxisEpsilon) {
			t.Errorf("axis %d = %v, want %v", i, axes[i], want[i])
		}
	}
}

func TestLocalAxesNormalizesScale(t *testing.T) {
	// 5x uniform scale + 90 deg yaw: rotation columns have length 5,
	// LocalAxes should renormalize.
	matrix := mgl32.HomogRotate3DY(float32(math.Pi / 2)).Mul4(mgl32.Scale3D(5, 5, 5))
	axes := LocalAxes(matrix)
	for i := 0; i < 3; i++ {
		if math.Abs(float64(axes[i].Len()-1)) > 1e-5 {
			t.Errorf("axis %d length = %f, want 1", i, axes[i].Len())
		}
	}
}

func TestLocalAxesDegenerateColumnFallsBackToWorld(t *testing.T) {
	matrix := mgl32.Mat4{} // all-zero matrix
	axes := LocalAxes(matrix)
	want := [3]mgl32.Vec3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for i := 0; i < 3; i++ {
		if !vec3Approx(axes[i], want[i], gizmoAxisEpsilon) {
			t.Errorf("degenerate axis %d = %v, want fallback %v", i, axes[i], want[i])
		}
	}
}
