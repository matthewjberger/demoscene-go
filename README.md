# indigo

An early game engine written in Go. Data-oriented from the ground up:
archetype ECS, a `wgpu`-backed render graph with declarative passes,
dual-world simulation and renderer separation, free-function systems
wired through a named schedule. Runs natively (GLFW) and on the web
(WebAssembly + canvas).

Architecture notes live in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Apps in this repo

- `cmd/editor` — indigo's editor (currently a stand-in: a grid of
  spinning primitives over a procedural sky and ground grid, lit by
  a directional sun, antialiased with FXAA, viewed through a
  pan-orbit camera). Will grow into a real scene editor.
- `cmd/breakout` — a small breakout clone built on the engine. A/D
  moves the paddle, space launches, R restarts.

## Quickstart

```
just run            # editor
just run breakout   # breakout
```

`just --list` for the rest.

## Provenance

The engine borrows from a handful of repos and credits them in the
relevant places:

- [`go-template`](https://github.com/matthewjberger/go-template). Module layout, justfile, dual MIT/Apache license.
- [`freecs-go`](https://github.com/matthewjberger/freecs-go). Vendored inline as `ecs/`.
- [`wgpu-example-go`](https://github.com/matthewjberger/wgpu-example-go). GLFW + cogentcore/webgpu surface setup, platform glue.
- [`nightshade`](https://github.com/matthewjberger/nightshade). Several of the data-oriented design choices (dual-world ECS, render graph shape, transform propagation, pan-orbit camera math) come from this Rust engine.

## License

Dual-licensed under [MIT](LICENSE-MIT) or [Apache-2.0](LICENSE-APACHE) at your option.
