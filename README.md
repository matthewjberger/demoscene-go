# rendergraph-go

A small, data-oriented Go engine: ECS-as-storage, a frame render graph
on top of wgpu, dual-world separation between simulation and renderer,
and free-function systems instead of objects with methods.

The ECS layer is [`freecs-go`](https://github.com/matthewjberger/freecs-go),
vendored inline as `ecs/` so the engine and its storage can evolve
together. The graphics boilerplate -- GLFW + `cogentcore/webgpu` surface
setup, the platform glue -- is lifted from
[`wgpu-example-go`](https://github.com/matthewjberger/wgpu-example-go).
The render graph (slot-based passes, transient/external resource
tracking, version-driven bind-group invalidation) and the engine-level
patterns are based on the design from
[`nightshade`](https://github.com/matthewjberger/nightshade).

Architecture notes live in [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md).

## Quickstart

```
just run
```

`just --list` for the rest.

## Provenance

- [`go-template`](https://github.com/matthewjberger/go-template) -- module layout, justfile, dual MIT/Apache license.
- [`freecs-go`](https://github.com/matthewjberger/freecs-go) -- vendored inline as `ecs/`.
- [`wgpu-example-go`](https://github.com/matthewjberger/wgpu-example-go) -- GLFW + cogentcore/webgpu surface setup, platform glue.
- [`nightshade`](https://github.com/matthewjberger/nightshade) -- render graph and engine-level patterns.

## License

Dual-licensed under [MIT](LICENSE-MIT) or [Apache-2.0](LICENSE-APACHE) at your option.
