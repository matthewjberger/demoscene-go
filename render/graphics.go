package render

import (
	"github.com/matthewjberger/indigo/ecs"
	"github.com/matthewjberger/indigo/ui"
)

// Graphics is the engine's single runtime-tunable settings resource.
// Every pass reads its toggle / parameter from this struct each
// frame; nothing else holds renderer state. Adding a new tunable
// means adding a field here, wiring its default in [DefaultGraphics],
// and reading it in the pass.
//
// Sub-structs group related fields so the type stays scannable. The
// flat layout (no nested pointers) keeps the resource cheap to copy
// and trivially comparable.
type Graphics struct {
	// Visibility toggles.
	ShowSky       bool
	ShowGrid      bool
	ShowBounds    bool
	ShowNormals   bool
	ShowSkeletons bool

	// Post-process toggles + parameters.
	Exposure    float32
	FxaaEnabled bool
	Bloom       Bloom
	Ssao        Ssao

	// GPU culling.
	Cull Cull

	// Debug overlay line styling.
	Lines DebugLines
}

// Bloom collects the bloom post-process tunables.
type Bloom struct {
	Enabled   bool
	Intensity float32
}

// Ssao collects the screen-space ambient occlusion tunables.
type Ssao struct {
	Enabled     bool
	Radius      float32
	Bias        float32
	Intensity   float32
	SampleCount int
}

// Cull holds the GPU culling toggle + minimum on-screen pixel
// diameter below which an instance is dropped.
type Cull struct {
	Enabled            bool
	MinScreenPixelSize float32
}

// DebugLines collects styling for the bounding-volume / normal /
// skeleton overlays. Sizes are in world units; colors are RGBA.
type DebugLines struct {
	NormalLength       float32
	NormalColor        [4]float32
	SkeletonJointSize  float32
	SkeletonJointColor [4]float32
	SkeletonBoneColor  [4]float32
}

// DefaultGraphics returns the engine's starting settings: sky +
// grid on, FXAA on, bloom on (subtle), SSAO on (tight radius),
// GPU culling on (1-pixel cutoff), debug overlays off.
func DefaultGraphics() Graphics {
	return Graphics{
		ShowSky:       true,
		ShowGrid:      true,
		ShowBounds:    false,
		ShowNormals:   false,
		ShowSkeletons: false,
		Exposure:      1.0,
		FxaaEnabled:   true,
		Bloom: Bloom{
			Enabled:   true,
			Intensity: 0.04,
		},
		Ssao: Ssao{
			Enabled:     true,
			Radius:      0.5,
			Bias:        0.025,
			Intensity:   1.5,
			SampleCount: 32,
		},
		Cull: Cull{
			Enabled:            true,
			MinScreenPixelSize: 1.0,
		},
		Lines: DebugLines{
			NormalLength:       0.08,
			NormalColor:        [4]float32{1.0, 0.92, 0.2, 0.95},
			SkeletonJointSize:  0.04,
			SkeletonJointColor: [4]float32{0.4, 1.0, 0.4, 1.0},
			SkeletonBoneColor:  [4]float32{1.0, 0.85, 0.2, 1.0},
		},
	}
}

// UpdateGraphicsToggles flips visibility toggles based on this
// frame's keyboard input: G grid, S sky, F FXAA, B bounds, N
// normals, K skeletons. No-ops when a text input is focused so
// keyboard text entry doesn't trip toggles.
func UpdateGraphicsToggles(world *ecs.World) {
	if ecs.HasResource[ui.WorldRef](world) {
		if ui.AnyTextInputFocused(ecs.MustResource[ui.WorldRef](world).World) {
			return
		}
	}
	input := ecs.MustResource[Input](world)
	settings := ecs.MustResource[Graphics](world)
	for _, key := range input.KeysJustDown {
		switch key {
		case 'G':
			settings.ShowGrid = !settings.ShowGrid
		case 'S':
			settings.ShowSky = !settings.ShowSky
		case 'F':
			settings.FxaaEnabled = !settings.FxaaEnabled
		case 'B':
			settings.ShowBounds = !settings.ShowBounds
		case 'N':
			settings.ShowNormals = !settings.ShowNormals
		case 'K':
			settings.ShowSkeletons = !settings.ShowSkeletons
		}
	}
}
