package asset

// MorphWeights holds the per-target blend weights for an entity whose mesh has
// morph targets. Animation morph-weight channels write into Weights; the mesh
// pass uploads them per instance and the vertex shader applies the weighted
// displacements.
type MorphWeights struct {
	Weights []float32
}
