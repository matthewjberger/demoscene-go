package pass

import (
	"fmt"
	"log"

	"github.com/matthewjberger/indigo/ecs"
	"github.com/matthewjberger/indigo/render"
)

type RebuildIbl struct{}

func (RebuildIbl) Apply(world *ecs.World, renderer *render.Renderer) {
	resource, ok := ecs.Resource[IBLResource](world)
	if !ok {
		log.Printf("rebuild_ibl: no IBLResource on world")
		return
	}
	if err := resource.IBL.Rebake(renderer.Device, renderer.Queue); err != nil {
		log.Printf("rebuild_ibl: %v", fmt.Errorf("rebake: %w", err))
	}
}

// LoadHdrSkybox bakes a decoded equirectangular HDRI into the IBL environment
// on the render thread, then switches the sky pass to it.
type LoadHdrSkybox struct {
	DisplayName string
	Width       uint32
	Height      uint32
	Pixels      []byte
}

func (c LoadHdrSkybox) Apply(world *ecs.World, renderer *render.Renderer) {
	resource, ok := ecs.Resource[IBLResource](world)
	if !ok {
		log.Printf("load_hdr_skybox: no IBLResource on world")
		return
	}
	if err := resource.IBL.LoadEquirect(renderer.Device, renderer.Queue, c.Pixels, c.Width, c.Height); err != nil {
		log.Printf("load_hdr_skybox: %v", err)
		return
	}
	settings := ecs.MustResource[render.Graphics](world)
	settings.HdriLoaded = true
	settings.ShowSky = true
	log.Printf("hdri loaded: %s (%dx%d)", c.DisplayName, c.Width, c.Height)
}
