package render

// EngineConfig surfaces the compile-time GPU-buffer-shape caps the
// engine ships with: shadow grid sizes, max light counts, IBL
// resolutions, selection-mask cap. These determine the size of
// allocated storage / texture buffers and changing them at runtime
// is not safe — they are knobs for forks or build-time overrides.
//
// The shipped defaults match the values used by the editor and
// breakout apps. apps that need to override should construct an
// EngineConfig, mutate fields, and set it as an ECS resource
// before NewRenderer / NewWorlds runs.
//
// Some passes still read raw constants today (see render/pass/*.go);
// the long-term direction is to route every cap through this
// struct. New caps should be added here first.
type EngineConfig struct {
	// Shadow mapping.
	NumShadowCascades   int
	ShadowMapSize       uint32
	MaxSpotShadows      uint32
	MaxPointShadows     uint32
	SpotShadowAtlasSize uint32

	// Clustered lighting.
	MaxLightsBuffer     uint32
	MaxLightsPerCluster uint32
	ClusterGridX        uint32
	ClusterGridY        uint32
	ClusterGridZ        uint32

	// Image-based lighting.
	BrdfLutSize           uint32
	PrefilteredSize       uint32
	IrradianceSamples     uint32
	ProceduralCubemapSize uint32

	// Selection / picking.
	SelectionMaskMaxEntities uint32

	// Skinned mesh joint cap.
	MaxJointsPerSkin uint32

	// Bloom mip chain depth.
	BloomMipCount uint32
}

// DefaultEngineConfig returns the configuration the engine ships
// with: matches the per-pass constants at the time of writing.
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		NumShadowCascades:        4,
		ShadowMapSize:            2048,
		MaxSpotShadows:           4,
		MaxPointShadows:          4,
		SpotShadowAtlasSize:      2048,
		MaxLightsBuffer:          1024,
		MaxLightsPerCluster:      256,
		ClusterGridX:             16,
		ClusterGridY:             9,
		ClusterGridZ:             24,
		BrdfLutSize:              256,
		PrefilteredSize:          512,
		IrradianceSamples:        1024,
		ProceduralCubemapSize:    1024,
		SelectionMaskMaxEntities: 4096,
		MaxJointsPerSkin:         128,
		BloomMipCount:            5,
	}
}
