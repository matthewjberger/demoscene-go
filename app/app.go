// Package app is the data-oriented analogue of nightshade's State trait.
//
// Nightshade's State is an OO trait the application implements; methods
// dispatch through a vtable. rendergraph-go keeps the same lifecycle hooks but
// stores them as function-value fields on an [App] struct. Applications
// construct an [App] and hand it to the main loop; no interfaces, no
// vtables, no inheritance, just data.
//
// Conventions:
//   - Initialize runs once after the renderer is built.
//   - ConfigureRenderGraph runs once after Initialize. This is where the
//     application registers its passes against the engine's resources.
//   - RunSystems runs every frame before rendering.
//   - PreRender runs every frame after RunSystems, before [Renderer.RenderFrame].
//
// Every hook is optional; nil fields are skipped.
package app

import (
	"rendergraph-go/ecs"
	"rendergraph-go/render"
)

// App is a bundle of lifecycle hooks. Mirrors the names from
// nightshade's State trait.
type App struct {
	Initialize           func(world *ecs.World)
	ConfigureRenderGraph func(world *ecs.World, renderer *render.Renderer)
	RunSystems           func(world *ecs.World)
	PreRender            func(world *ecs.World)
}
