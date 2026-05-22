package ecs

// EcsCommand is the sealed interface for deferred ECS-world
// mutations: despawns, structural component changes, anything that
// can't safely happen inside an `IterChanged` / query closure.
// Commands queued via [QueueEcsCommand] drain at the
// [ProcessEcsCommands] call site in the frame schedule, where no
// iteration is in flight.
type EcsCommand interface {
	apply(world *World)
}

// DespawnEntity removes the entity from the world. No-op if the
// entity has already been despawned.
type DespawnEntity struct{ Entity Entity }

func (c DespawnEntity) apply(world *World) {
	world.Despawn(c.Entity)
}

// EcsCommandQueueResource owns the per-frame command queue. One
// instance is registered on every world that needs deferred
// mutations.
type EcsCommandQueueResource struct {
	commands []EcsCommand
}

// QueueEcsCommand appends a command to the world's queue. The
// command applies the next time [ProcessEcsCommands] runs.
// Resource must already exist on the world (registered at engine
// setup).
func QueueEcsCommand(world *World, command EcsCommand) {
	queue := MustResource[EcsCommandQueueResource](world)
	queue.commands = append(queue.commands, command)
}

// ProcessEcsCommands drains the world's queue, applying each
// command in order. Safe to call once per frame as a system —
// commands queued during this drain are deferred to the next call
// because the loop only iterates the slice captured at entry.
func ProcessEcsCommands(world *World) {
	queue := MustResource[EcsCommandQueueResource](world)
	pending := queue.commands
	queue.commands = queue.commands[:0]
	for _, cmd := range pending {
		cmd.apply(world)
	}
}
