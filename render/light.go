package render

import "github.com/matthewjberger/indigo/transform"

type LightType uint32

const (
	LightTypeDirectional LightType = 0
	LightTypePoint       LightType = 1
	LightTypeSpot        LightType = 2
)

type Light struct {
	Type           LightType
	Color          transform.Vec3
	Intensity      float32
	Range          float32
	InnerConeAngle float32
	OuterConeAngle float32
	CastShadows    bool
	ShadowBias     float32

	LightSize float32

	// CookieEnabled projects CookieLayer (a material texture-array layer)
	// through the spotlight's view-projection to tint the light. Spotlights
	// with a shadow index only.
	CookieEnabled bool
	CookieLayer   uint32
}

const MaxLights = 8
