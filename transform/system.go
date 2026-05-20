package transform

import (
	"sort"

	"indigo/ecs"
)

// MaxHierarchyDepth is the maximum parent-chain depth the propagation
// system will walk.
const MaxHierarchyDepth = 256

// PropagationState is the per-frame scratch the propagation system
// reuses across frames to avoid allocating fresh entity slices.
// Stored as a resource on the world; the system resets the slices to
// length zero (preserving capacity).
type PropagationState struct {
	Dirty  []ecs.Entity
	Depths []int
	Order  []int
}

// NewPropagationState returns an empty propagation scratch.
func NewPropagationState() PropagationState {
	return PropagationState{}
}

// UpdateGlobalTransforms is the per-frame transform propagation
// system. Walks every entity marked [LocalTransformDirty], orders
// them shallow-first by parent-chain depth, and rewrites each one's
// [GlobalTransform] from its parent's GlobalTransform plus its own
// LocalTransform.
//
// Because dirty entities are processed roots-down, by the time a
// child is touched its parent's GlobalTransform is either the value
// from the previous frame (parent not dirty this frame, still
// correct) or already rewritten this frame (parent dirty, processed
// at a shallower depth). One matrix multiply per dirty entity, with
// no recursive ancestor walk per entity.
func UpdateGlobalTransforms(world *ecs.World) {
	scratch := ecs.Resource[PropagationState](world)
	scratch.Dirty = scratch.Dirty[:0]
	scratch.Depths = scratch.Depths[:0]
	scratch.Order = scratch.Order[:0]

	localMask := ecs.MaskOf[LocalTransform](world)
	globalMask := ecs.MaskOf[GlobalTransform](world)
	dirtyMask := ecs.MaskOf[LocalTransformDirty](world)

	world.ForEach(localMask|globalMask|dirtyMask, 0, func(entity ecs.Entity, _ *ecs.Archetype, _ int) {
		scratch.Dirty = append(scratch.Dirty, entity)
	})

	if len(scratch.Dirty) == 0 {
		return
	}

	for _, entity := range scratch.Dirty {
		scratch.Depths = append(scratch.Depths, depthOf(world, entity))
	}

	for index := range scratch.Dirty {
		scratch.Order = append(scratch.Order, index)
	}
	sort.SliceStable(scratch.Order, func(i, j int) bool {
		return scratch.Depths[scratch.Order[i]] < scratch.Depths[scratch.Order[j]]
	})

	for _, index := range scratch.Order {
		entity := scratch.Dirty[index]
		depth := scratch.Depths[index]
		matrix := computeMatrix(world, entity, depth)
		if global, ok := ecs.GetMut[GlobalTransform](world, entity); ok {
			global.Matrix = matrix
		}
		ecs.Remove[LocalTransformDirty](world, entity)
	}
}

// depthOf returns the entity's parent-chain depth: 0 for roots, one
// plus the parent's depth otherwise. Returns [MaxHierarchyDepth]+1
// when the chain exceeds the limit (a stand-in for "this entity's
// hierarchy is too deep to resolve safely"); the propagation step
// treats those entities as if they were disconnected from their
// parent so a runaway chain or a cycle cannot accumulate unbounded
// translation.
func depthOf(world *ecs.World, entity ecs.Entity) int {
	current := entity
	for depth := 0; depth <= MaxHierarchyDepth; depth++ {
		parent, hasParent := ecs.Get[Parent](world, current)
		if !hasParent || parent.IsRoot {
			return depth
		}
		current = parent.Entity
	}
	return MaxHierarchyDepth + 1
}

// computeMatrix produces the entity's world-space matrix as
// parent_global * local, or just local if the entity is a root, its
// parent has no GlobalTransform, or its hierarchy depth exceeded
// [MaxHierarchyDepth] (in which case the depth is
// [MaxHierarchyDepth]+1 and we treat the entity as disconnected from
// its broken/too-deep chain). The parent's GlobalTransform read here
// is whatever's currently in the column: either last frame's value
// (parent not dirty) or this frame's recomputed value (parent was
// dirty too, processed at a shallower depth before this entity).
func computeMatrix(world *ecs.World, entity ecs.Entity, depth int) Mat4 {
	local, ok := ecs.Get[LocalTransform](world, entity)
	if !ok {
		return Mat4Identity()
	}
	localMatrix := AsMatrix(local)

	if depth > MaxHierarchyDepth {
		return localMatrix
	}

	parent, hasParent := ecs.Get[Parent](world, entity)
	if !hasParent || parent.IsRoot {
		return localMatrix
	}
	parentGlobal, ok := ecs.Get[GlobalTransform](world, parent.Entity)
	if !ok {
		return localMatrix
	}
	return parentGlobal.Matrix.Mul4(localMatrix)
}
