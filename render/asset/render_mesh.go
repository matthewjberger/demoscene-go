package asset

// RenderMesh is an ECS component pointing at a mesh in the engine's
// [MeshAssets] registry. Entities with [transform.GlobalTransform] +
// RenderMesh are drawn by the mesh pass; the handle picks which
// mesh. Just a handle — no per-entity material or mesh data on
// the component.
type RenderMesh struct {
	Mesh MeshHandle
}
