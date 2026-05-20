//go:build !js

// Command rendergraph-go is the engine demo entry point: a grid of spinning
// triangles driven by a dual-world ECS through the engine's render
// graph, viewed through a pan-orbit camera.
package main

import (
	"errors"
	"log"
	"runtime"
	"time"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/cogentcore/webgpu/wgpuglfw"
	"github.com/go-gl/glfw/v3.3/glfw"

	"rendergraph-go/ecs"
	"rendergraph-go/render"
	"rendergraph-go/transform"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	setupLogging()

	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(1280, 720, "rendergraph-go", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Destroy()

	instance := wgpu.CreateInstance(nil)
	defer instance.Release()

	surface := instance.CreateSurface(wgpuglfw.GetSurfaceDescriptor(window))

	width, height := window.GetSize()
	renderer, err := render.NewRenderer(instance, surface, uint32(width), uint32(height))
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Release()

	worlds, demo := buildWorlds(renderer)
	installInputCallbacks(window, worlds.Engine)

	window.SetSizeCallback(func(_ *glfw.Window, w, h int) {
		if w > 0 && h > 0 {
			if err := renderer.Resize(uint32(w), uint32(h)); err != nil {
				log.Printf("resize error: %v", err)
			}
		}
	})

	last := time.Now()
	for !window.ShouldClose() {
		glfw.PollEvents()

		now := time.Now()
		delta := float32(now.Sub(last).Seconds())
		last = now

		tickFrame(worlds, demo, delta)

		switch err := renderer.RenderFrame(worlds.Engine); {
		case err == nil:
		case errors.Is(err, render.ErrSurfaceLost):
			renderer.Reconfigure()
		default:
			log.Fatal(err)
		}
	}
}

// installInputCallbacks wires GLFW's cursor / mouse-button / scroll
// callbacks to accumulate frame deltas into the engine world's Input
// resource. tickFrame zeroes the deltas at the end of each frame.
func installInputCallbacks(window *glfw.Window, engine *ecs.World) {
	var previousMouse transform.Vec2
	var haveMouse bool

	window.SetCursorPosCallback(func(_ *glfw.Window, x, y float64) {
		input := ecs.Resource[render.Input](engine)
		current := transform.Vec2{float32(x), float32(y)}
		if haveMouse {
			input.MouseDelta = input.MouseDelta.Add(current.Sub(previousMouse))
		}
		input.MousePosition = current
		previousMouse = current
		haveMouse = true
	})

	window.SetMouseButtonCallback(func(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, _ glfw.ModifierKey) {
		input := ecs.Resource[render.Input](engine)
		pressed := action == glfw.Press
		switch button {
		case glfw.MouseButtonLeft:
			input.LeftDown = pressed
		case glfw.MouseButtonRight:
			input.RightDown = pressed
		case glfw.MouseButtonMiddle:
			input.MiddleDown = pressed
		}
	})

	window.SetScrollCallback(func(_ *glfw.Window, _, yOffset float64) {
		input := ecs.Resource[render.Input](engine)
		input.Wheel += float32(yOffset)
	})

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, _ int, action glfw.Action, _ glfw.ModifierKey) {
		if key == glfw.KeyEscape && action == glfw.Press {
			w.SetShouldClose(true)
		}
	})
}
