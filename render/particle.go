package render

import "github.com/matthewjberger/indigo/transform"

type ParticleShape uint32

const (
	ParticleShapePoint ParticleShape = iota
	ParticleShapeSphere
	ParticleShapeCone
	ParticleShapeBox
)

type ParticleEmitterType uint32

const (
	ParticleFirework ParticleEmitterType = iota
	ParticleFire
	ParticleSmoke
	ParticleSpark
	ParticleGlow
)

// ParticleEmitter spawns GPU-simulated billboard particles from the entity's
// world position each frame. Accumulator is internal spawn-rate carry state.
type ParticleEmitter struct {
	Direction          transform.Vec3
	SpawnRate          float32
	VelocityMin        float32
	VelocityMax        float32
	Spread             float32
	LifetimeMin        float32
	LifetimeMax        float32
	SizeStart          float32
	SizeEnd            float32
	Gravity            float32
	Drag               float32
	TurbulenceStrength float32
	TurbulenceFreq     float32
	Shape              ParticleShape
	ShapeParams        [4]float32
	EmitterType        ParticleEmitterType
	ColorStart         [4]float32
	ColorEnd           [4]float32
	EmissiveStrength   float32
	Accumulator        float32
}
