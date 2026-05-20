package render

import (
	"fmt"

	"github.com/cogentcore/webgpu/wgpu"
)

// MeshHandle is an opaque index into a [MeshAssets] registry. Zero is
// not a special value; the renderer hands the first registered mesh
// out as handle 0.
type MeshHandle uint32

// MeshVertex is the input layout the engine's stock mesh shader
// expects: position and color, each padded to vec4 stride. Custom
// passes that bypass MeshAssets can use their own layout.
type MeshVertex struct {
	Position [4]float32
	Color    [4]float32
}

// meshEntry is the renderer's per-mesh record: GPU vertex buffer plus
// the vertex count to feed to Draw.
type meshEntry struct {
	Name        string
	Vertices    *wgpu.Buffer
	VertexCount uint32
}

// MeshAssets is the engine's per-renderer mesh registry. Stored on the
// engine ECS world via [MeshAssetsResource] so passes that want to draw
// a mesh can look up its GPU buffer by handle. Mirrors the role of
// nightshade's mesh asset cache, scaled down to "list of vertex buffers
// indexed by handle."
type MeshAssets struct {
	entries []meshEntry
}

// MeshAssetsResource wraps a *MeshAssets so it can be installed on an
// ECS world via Go-type-keyed resources. Mutations through the wrapped
// pointer persist; freecs-go keeps a stable pointer to the wrapper.
type MeshAssetsResource struct {
	Assets *MeshAssets
}

// NewMeshAssets returns an empty registry. Use [MeshAssets.Register]
// to add meshes.
func NewMeshAssets() *MeshAssets { return &MeshAssets{} }

// Register uploads vertices to a new GPU buffer and returns the handle
// callers should attach to entities via [RenderMesh].
func (assets *MeshAssets) Register(device *wgpu.Device, name string, vertices []MeshVertex) (MeshHandle, error) {
	buffer, err := device.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Label:    name + " vertex buffer",
		Contents: wgpu.ToBytes(vertices),
		Usage:    wgpu.BufferUsageVertex,
	})
	if err != nil {
		return 0, fmt.Errorf("mesh assets: %s vertex buffer: %w", name, err)
	}
	handle := MeshHandle(len(assets.entries))
	assets.entries = append(assets.entries, meshEntry{
		Name:        name,
		Vertices:    buffer,
		VertexCount: uint32(len(vertices)),
	})
	return handle, nil
}

// Lookup returns the per-mesh entry for handle. Callers that pass an
// out-of-range handle get a false result and should skip the entity.
func (assets *MeshAssets) Lookup(handle MeshHandle) (*meshEntry, bool) {
	if int(handle) >= len(assets.entries) {
		return nil, false
	}
	return &assets.entries[handle], true
}

// Count returns the number of registered meshes.
func (assets *MeshAssets) Count() int { return len(assets.entries) }

// Release frees every GPU buffer owned by the registry.
func (assets *MeshAssets) Release() {
	for index := range assets.entries {
		if assets.entries[index].Vertices != nil {
			assets.entries[index].Vertices.Release()
			assets.entries[index].Vertices = nil
		}
	}
	assets.entries = nil
}

// UnitTriangleVertices is the built-in mesh registered by NewRenderer
// as the engine's stock primitive. Same geometry as the prior
// hardcoded triangle.
var UnitTriangleVertices = []MeshVertex{
	{Position: [4]float32{0.5, -0.5, 0.0, 1.0}, Color: [4]float32{1.0, 0.0, 0.0, 1.0}},
	{Position: [4]float32{-0.5, -0.5, 0.0, 1.0}, Color: [4]float32{0.0, 1.0, 0.0, 1.0}},
	{Position: [4]float32{0.0, 0.5, 0.0, 1.0}, Color: [4]float32{0.0, 0.0, 1.0, 1.0}},
}
