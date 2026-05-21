package render

import (
	"math"

	"indigo/ecs"
	"indigo/transform"
)

// UpdateBoundingVolumeLines emits one wireframe AABB per entity that
// has a RenderMesh + GlobalTransform when [GraphicsSettings.ShowBounds]
// is true. Reads the mesh's local-space AABB from MeshAssets and
// pushes 12 world-space line segments into the [Lines] resource for
// the next [LinesPass] frame.
func UpdateBoundingVolumeLines(world *ecs.World) {
	settings, ok := ecs.Resource[GraphicsSettings](world)
	if !ok || !settings.ShowBounds {
		return
	}
	linesRes, ok := ecs.Resource[LinesResource](world)
	if !ok {
		return
	}
	assetsRes, ok := ecs.Resource[MeshAssetsResource](world)
	if !ok {
		return
	}
	assets := assetsRes.Assets
	lines := linesRes.Lines

	color := [4]float32{0.4, 0.95, 0.55, 0.9}
	meshMask := ecs.MustMaskOf[RenderMesh](world)
	world.ForEach(meshMask, 0, func(entity ecs.Entity, _ *ecs.Archetype, _ int) {
		mesh, ok := ecs.Get[RenderMesh](world, entity)
		if !ok {
			return
		}
		global, ok := ecs.Get[transform.GlobalTransform](world, entity)
		if !ok {
			return
		}
		bounds := assets.Bounds(mesh.Mesh)
		if bounds == (BoundingVolume{}) {
			return
		}
		lines.AddBox(bounds, &global.Matrix, color)
	})
}

// BoundingVolume is an axis-aligned bounding box in mesh-local
// coordinates. Stored alongside each registered mesh so render passes
// and editor systems can ask "what's the local-space extent of this
// mesh" without re-walking the vertex data every frame.
type BoundingVolume struct {
	Min [3]float32
	Max [3]float32
}

// ComputeBounds returns the AABB enclosing every vertex position in
// vertices. An empty slice produces a zero-extent box at the origin.
func ComputeBounds(vertices []MeshVertex) BoundingVolume {
	if len(vertices) == 0 {
		return BoundingVolume{}
	}
	min := [3]float32{
		float32(math.Inf(1)), float32(math.Inf(1)), float32(math.Inf(1)),
	}
	max := [3]float32{
		float32(math.Inf(-1)), float32(math.Inf(-1)), float32(math.Inf(-1)),
	}
	for i := range vertices {
		p := vertices[i].Position
		for axis := 0; axis < 3; axis++ {
			if p[axis] < min[axis] {
				min[axis] = p[axis]
			}
			if p[axis] > max[axis] {
				max[axis] = p[axis]
			}
		}
	}
	return BoundingVolume{Min: min, Max: max}
}

// Corners returns the eight corner points of the AABB.
func (b BoundingVolume) Corners() [8][3]float32 {
	return [8][3]float32{
		{b.Min[0], b.Min[1], b.Min[2]},
		{b.Max[0], b.Min[1], b.Min[2]},
		{b.Max[0], b.Max[1], b.Min[2]},
		{b.Min[0], b.Max[1], b.Min[2]},
		{b.Min[0], b.Min[1], b.Max[2]},
		{b.Max[0], b.Min[1], b.Max[2]},
		{b.Max[0], b.Max[1], b.Max[2]},
		{b.Min[0], b.Max[1], b.Max[2]},
	}
}

// EdgeIndices returns the 24 corner-index pairs (12 edges) that make
// up the box's wireframe. Caller pairs these with [Corners].
var BoundingBoxEdges = [24]uint8{
	0, 1, 1, 2, 2, 3, 3, 0,
	4, 5, 5, 6, 6, 7, 7, 4,
	0, 4, 1, 5, 2, 6, 3, 7,
}
