// Package ecs is the data-oriented Entity Component System used by rendergraph-go.
//
// It is a verbatim copy of the freecs-go library, renamed to "ecs" so the
// engine ships as a single self-contained module. See
// https://github.com/matthewjberger/freecs-go for the upstream docs.
//
// Entities are generational handles. Components are plain Go structs.
// Storage is one archetype table per unique component set, with each
// component held as a contiguous typed column. Queries walk archetypes
// whose mask satisfies the query and iterate the relevant columns
// directly. Hot-path iteration via [Iter1] through [Iter4] materializes
// typed []T views over column memory once per archetype and walks them
// with no per-element reflection.
//
// # Quick start
//
//	world := ecs.New()
//	POSITION := ecs.Register[Position](world)
//	VELOCITY := ecs.Register[Velocity](world)
//
//	entity := world.Spawn(POSITION | VELOCITY)
//	ecs.Set(world, entity, Position{X: 0, Y: 0})
//	ecs.Set(world, entity, Velocity{X: 5, Y: 0})
//
//	ecs.Iter2[Position, Velocity](world, 0, 0, func(_ ecs.Entity, position *Position, velocity *Velocity) {
//	    position.X += velocity.X
//	    position.Y += velocity.Y
//	})
//
// # API shape
//
// Non-generic operations on a [World] are methods: [World.Spawn],
// [World.Despawn], [World.AddComponents], [World.Query], [World.ForEach],
// [World.ApplyCommands], and so on. Operations parameterized over a
// component type are top-level generic functions: [Get], [Set], [Add],
// [Remove], [Has], [GetMut], [Changed], [MarkChanged], [Iter1] through
// [Iter4], [IterChanged1] through [IterChanged4], [ParallelIter1] through
// [ParallelIter4], [Column], [Send], [ReadEvents], [DrainEvents],
// [AddTag], [HasTag], [QueryTag], [SetResource], [Resource]. The split is
// forced because Go forbids type parameters on methods.
//
// # Concurrency
//
// A [World] is not safe for concurrent use; wrap it in a mutex if
// multiple goroutines need to touch it. The [ParallelIter1] family fans
// out one goroutine per matching archetype within a single call and is
// the supported way to do row-parallel work; the callback must not
// mutate world topology and must access components only through the
// supplied typed pointers.
package ecs
