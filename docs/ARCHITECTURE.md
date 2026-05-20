# Architecture

rendergraph-go is a Go port of the data-oriented design used in
[nightshade](https://github.com/matthewjberger/nightshade), built on
[freecs-go](https://github.com/matthewjberger/freecs-go) (vendored inline as
`ecs/`).

## Principles

- **No OOP.** Data and functions, never objects with methods that own
  business logic. Where Rust nightshade uses traits (`State`, `PassNode`),
  rendergraph-go uses structs whose fields are function values.
- **One ECS world.** `*ecs.World` holds entities, components, and
  type-keyed resources. The renderer is a resource on the world.
- **Render graph for rendering.** Passes declare which slot names they
  read and write; the graph wires slots to resources and decides
  clear-vs-load based on first-write.

## Module layout

Inherits from `go-template`:

- `cmd/<binary>/` for binaries.
- Sibling top-level packages for libraries (`ecs`, `render`, `app`).
- `justfile` for build / test / format / audit recipes.
- Dual MIT/Apache license, `docs/ARCHITECTURE.md` (this file).

## ECS world

The `ecs` package is freecs-go renamed. Conventions:

- Components are plain structs registered with
  `bit := ecs.Register[MyComponent](world)`. Up to 64 per world.
- Resources are keyed by Go type. Define named types (`type DeltaTime float32`)
  so each scalar gets its own slot.
- Schedules order named systems; `world.Step()` advances change-detection
  ticks and rolls event queues at the end of each frame.

## Renderer-as-resource

The renderer is stored on the world via `ecs.SetResource(world, RendererResource{...})`
and recovered the same way. Systems that need GPU access look up the
resource; nothing else holds the renderer directly. This mirrors how
nightshade hangs its `WgpuRenderer` off `world.resources`.

## Render graph

The Go render graph keeps the fundamentals of nightshade's:

- `ResourceID` handles into a flat descriptor / handle / version table.
- Resource kinds: external / transient × color / depth. External
  resources are refreshed by the caller each frame (the swapchain);
  transients are owned by the graph and reallocated on resize.
- `Pass` is a struct of function-value fields (`Prepare`, `Execute`,
  `InvalidateBindGroups`, `Release`) plus declared `Reads`/`Writes` slot
  lists and opaque `State`. The graph never dispatches through an
  interface.
- `Compile` walks passes in insertion order and records the first-write
  set: the (pass, resource) pairs that clear-on-load instead of loading.
- `Execute` runs every pass in order, allocating a fresh
  `PassContext` per pass; the context holds the device, queue, encoder,
  resources, slot bindings, and the ECS world (the data-oriented
  equivalent of nightshade's `configs: &World` field on
  `PassExecutionContext`).
- Resource versioning: `Resources.Versions[id]` increments every time the
  underlying handle changes (external view replacement, transient
  reallocation on resize). Each pass keeps a per-slot version snapshot;
  on mismatch the graph calls `Pass.InvalidateBindGroups` so cached bind
  groups referencing stale views get rebuilt. Mirrors nightshade's
  `versions` map + `PassNode::invalidate_bind_groups` pair.

### Default graph layout

`render.NewRenderer` registers three resources by default, matching the
nightshade default:

- `scene_color` -- transient color, surface format, RENDER_ATTACHMENT |
  TEXTURE_BINDING | COPY_SRC. Color passes (today: the triangle) write
  here.
- `depth` -- transient depth.
- `swapchain` -- external color. The present pass blits `scene_color`
  into it.

A single application call registers two passes against this layout:

```go
triangle, _ := render.NewTrianglePass(device, format, aspectFn)
renderer.Graph.AddPass(triangle, []render.SlotBinding{
    {Slot: "color", ResourceID: renderer.SceneColorID},
    {Slot: "depth", ResourceID: renderer.DepthID},
})

present, _ := render.NewPresentPass(device, format)
renderer.Graph.AddPass(present, []render.SlotBinding{
    {Slot: "input",  ResourceID: renderer.SceneColorID},
    {Slot: "output", ResourceID: renderer.SwapchainID},
})
```

Future passes (bloom, SSAO, tonemap) plug in between the triangle and the
present pass, reading and writing `scene_color` or new transients without
touching anything in the graph internals.

Extension points kept deliberately separate so future passes (bloom,
SSAO, OIT, etc.) plug in without changes to the graph internals:

- Add a pass: build a `*render.Pass`, call `graph.AddPass(pass, bindings)`,
  then `graph.Compile()`.
- Add an intermediate target: `graph.AddColorTexture` / `AddDepthTexture`
  with `ResourceKindTransientColor` / `ResourceKindTransientDepth`.

What's intentionally left out for now (and easy to add later):

- Topological sort of passes by read/write dependencies. Today execution
  follows insertion order; the data the sort needs is already collected.
- Dead-pass culling: a pass whose writes are never read can be skipped.
- Transient aliasing pool. Each transient is its own texture today; the
  nightshade pool would reuse textures with compatible descriptors.
- Subgraphs and per-pass enable/disable toggles.

## Lifecycle

```
main()
├── glfw window, wgpu instance + surface
├── render.NewRenderer(...)            → also builds the default graph (swapchain + depth)
├── ecs.New()                          → world
├── ecs.SetResource(world, DeltaTime, RendererResource)
├── app.Initialize(world)              → user scene setup
├── app.ConfigureRenderGraph(world, r) → user passes (e.g. NewTrianglePass)
├── renderer.Graph.Compile()
└── for each frame:
    ├── poll + delta
    ├── app.RunSystems(world)          → advance triangle, etc.
    ├── app.PreRender(world)
    ├── world.ApplyCommands(); world.Step()
    └── renderer.RenderFrame()         → executes the graph, presents
```

## Why a struct of functions instead of a `PassNode` interface

In Rust nightshade, `PassNode` is a trait with `&mut self` methods. In Go
the equivalent would be an interface with methods on a concrete struct
that owns the pass's GPU state. That works, but it conflates two things:
the pass's *data* (pipeline, buffers, bind groups) and its *behavior*
(prepare + execute funcs). The data-oriented split lifts behavior to
function values and keeps state purely as fields. Passes compose by
swapping function pointers, not by overriding methods.
