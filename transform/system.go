package transform

import "rendergraph-go/ecs"

// MaxHierarchyDepth is the maximum parent-chain depth the propagation
// system will walk.
const MaxHierarchyDepth = 256

// PropagationState is the per-frame scratch the propagation system
// reuses across frames to avoid allocating fresh entity slices and
// visited maps. Stored as a resource on the world; the system resets
// the slices to length zero (preserving capacity) and clears the map.
//
// Mirrors nightshade's `transform_state.dirty_entities` Vec which is
// `mem::take`'d at the top of `update_global_transforms_system`.
type PropagationState struct {
	Dirty   []ecs.Entity
	Visited map[ecs.Entity]struct{}
}

// NewPropagationState returns an empty propagation scratch.
func NewPropagationState() PropagationState {
	return PropagationState{
		Visited: make(map[ecs.Entity]struct{}, 16),
	}
}

// UpdateGlobalTransforms is the per-frame transform propagation
// system. Walks every entity marked [LocalTransformDirty] and rewrites
// its [GlobalTransform] from the parent chain, then clears the dirty
// marker.
//
// Reads [PropagationState] off the world for scratch. Iterates the
// dirty archetype with [ecs.World.ForEach] for direct table access.
func UpdateGlobalTransforms(world *ecs.World) {
	scratch := ecs.Resource[PropagationState](world)
	scratch.Dirty = scratch.Dirty[:0]

	localMask := ecs.MaskOf[LocalTransform](world)
	globalMask := ecs.MaskOf[GlobalTransform](world)
	dirtyMask := ecs.MaskOf[LocalTransformDirty](world)

	world.ForEach(localMask|globalMask|dirtyMask, 0, func(entity ecs.Entity, _ *ecs.Archetype, _ int) {
		scratch.Dirty = append(scratch.Dirty, entity)
	})

	for _, entity := range scratch.Dirty {
		matrix := computeGlobalMatrix(world, entity, scratch.Visited)
		if global, ok := ecs.GetMut[GlobalTransform](world, entity); ok {
			global.Matrix = matrix
		}
		ecs.Remove[LocalTransformDirty](world, entity)
	}
}

func computeGlobalMatrix(world *ecs.World, entity ecs.Entity, visited map[ecs.Entity]struct{}) Mat4 {
	clear(visited)
	return walkParents(world, entity, visited)
}

func walkParents(world *ecs.World, entity ecs.Entity, visited map[ecs.Entity]struct{}) Mat4 {
	if _, cycle := visited[entity]; cycle {
		return Mat4Identity()
	}
	visited[entity] = struct{}{}

	if len(visited) > MaxHierarchyDepth {
		return Mat4Identity()
	}

	local, ok := ecs.Get[LocalTransform](world, entity)
	if !ok {
		return Mat4Identity()
	}

	localMatrix := AsMatrix(local)

	parent, hasParent := ecs.Get[Parent](world, entity)
	if !hasParent || parent.IsRoot {
		return localMatrix
	}

	parentMatrix := walkParents(world, parent.Entity, visited)
	return parentMatrix.Mul4(localMatrix)
}
