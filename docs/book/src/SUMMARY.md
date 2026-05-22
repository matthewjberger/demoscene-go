# indigo

[Introduction](introduction.md)

---

# Getting Started

- [Installation](installation.md)
- [Your First Application](first-application.md)
- [Project Structure](project-structure.md)

---

# Architecture

- [Architecture Overview](architecture-overview.md)
- [The Three Worlds](three-worlds.md)
- [Frame Lifecycle](frame-lifecycle.md)

---

# Entity Component System

- [Archetype Storage](ecs-archetype.md)
- [Entities & Generational Handles](ecs-entities.md)
- [Components & Registration](ecs-components.md)
- [Spawning & Mutation](ecs-spawn.md)
- [Queries](ecs-queries.md)
- [Resources](ecs-resources.md)
- [Schedules & Systems](ecs-schedules.md)
- [Commands](ecs-commands.md)
- [Events](ecs-events.md)
- [Tags](ecs-tags.md)
- [Change Detection](ecs-change.md)
- [Parallel Iteration](ecs-parallel.md)
- [Multi-World & Bridges](ecs-multiworld.md)

---

# Core Systems

- [Math & Coordinates](math.md)
- [Transform Hierarchy](transform-hierarchy.md)
- [App Layer](app-layer.md)
- [Window & Input](window-input.md)
- [Cross-World Sync](app-sync.md)

---

# Rendering

- [Rendering Architecture](rendering-architecture.md)
- [The Render Graph](render-graph.md)
  - [Passes & Slot Manifests](render-graph-passes.md)
  - [Resources & Versioning](render-graph-resources.md)
  - [Compile & Schedule](render-graph-compile.md)
  - [Bind Group Invalidation](render-graph-bindgroups.md)
  - [Adding a Pass](render-graph-custom.md)
- [Cameras & Projection](cameras.md)
- [Pan/Orbit Camera](camera-pan-orbit.md)
- [Lights & Clustered Lighting](lighting.md)
- [Image-Based Lighting](ibl.md)
- [The Mesh Pass](pass-mesh.md)
- [Sky & Procedural Skybox](pass-sky.md)
- [Grid Pass](pass-grid.md)
- [Lines, Normals, Bounding Volumes](pass-debug-lines.md)
- [Gizmos](pass-gizmo.md)
- [Selection Mask & Outline](pass-outline.md)
- [Picking](pass-picking.md)
- [FXAA & Postprocess](pass-postprocess.md)
- [Present](pass-present.md)
- [Graphics Settings](graphics-settings.md)

---

# Assets

- [Mesh Assets](assets-mesh.md)
- [Texture Cache](assets-textures.md)
- [Materials & Material Registry](assets-materials.md)
- [glTF Loading](assets-gltf.md)
- [Spawning Loaded Scenes](assets-spawn.md)
- [Animation Data & Player](assets-animation.md)

---

# User Interface

- [Retained UI Overview](ui-overview.md)
- [UI Components & Builder](ui-builder.md)
- [Layout](ui-layout.md)
- [Interaction & Hit Testing](ui-interaction.md)
- [Fonts & Text](ui-text.md)
- [Text Input](ui-text-input.md)
- [UI ↔ Engine Bridge](ui-bridge.md)

---

# Platform

- [Desktop (GLFW)](platform-desktop.md)
- [WebAssembly](platform-wasm.md)

---

# Examples

- [Breakout](example-breakout.md)
- [Editor](example-editor.md)

---

# Appendix

- [Glossary](appendix-glossary.md)
- [Platform Support](appendix-platforms.md)
- [Troubleshooting](appendix-troubleshooting.md)
