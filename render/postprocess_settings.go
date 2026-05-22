package render

// PostProcessSettings carries tunables the postprocess pass reads
// each frame. Stored as an ECS resource on the engine world so
// systems (and eventually a slider in the editor UI) can mutate
// the values without touching the pass internals.
//
// Defaults match a moderately bright scene: full exposure, bloom
// adds about 4% of its intensity (subtle glow on bright pixels).
type PostProcessSettings struct {
	Exposure       float32
	BloomEnabled   bool
	BloomIntensity float32
}

// DefaultPostProcessSettings returns the engine's starting values.
func DefaultPostProcessSettings() PostProcessSettings {
	return PostProcessSettings{
		Exposure:       1.0,
		BloomEnabled:   true,
		BloomIntensity: 0.04,
	}
}
