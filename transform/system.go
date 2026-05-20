package transform

import (
	"log"
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
	Dirty    []ecs.Entity
	Depths   []int
	Order    []int
	Children map[ecs.Entity][]ecs.Entity
	Seen     map[ecs.Entity]struct{}
	Queue    []ecs.Entity

	depthCapReported bool
}

// NewPropagationState returns an empty propagation scratch.
func NewPropagationState() PropagationState {
	return PropagationState{
		Children: make(map[ecs.Entity][]ecs.Entity, 16),
		Seen:     make(map[ecs.Entity]struct{}, 16),
	}
}

// UpdateGlobalTransforms is the per-frame transform propagation
// system. Collects every entity marked [LocalTransformDirty],
// cascades the dirty set to include every descendant via the Parent
// chain (so moving a parent recomputes its children's globals),
// orders the result shallow-first by parent-chain depth, and
// rewrites each one's [GlobalTransform] from its parent's
// GlobalTransform plus its own LocalTransform.
//
// Cascade is per-frame: we build a children index from a single
// ForEach over Parent components, then BFS-expand the initial dirty
// set. This keeps [MarkDirty] cheap (it doesn't need to know about
// children) and tolerates Parent re-parenting without a separately-
// maintained cache.
func UpdateGlobalTransforms(world *ecs.World) {
	scratch := ecs.Resource[PropagationState](world)
	scratch.Dirty = scratch.Dirty[:0]
	scratch.Depths = scratch.Depths[:0]
	scratch.Order = scratch.Order[:0]
	scratch.Queue = scratch.Queue[:0]
	for k := range scratch.Children {
		delete(scratch.Children, k)
	}
	for k := range scratch.Seen {
		delete(scratch.Seen, k)
	}

	localMask := ecs.MaskOf[LocalTransform](world)
	globalMask := ecs.MaskOf[GlobalTransform](world)
	dirtyMask := ecs.MaskOf[LocalTransformDirty](world)

	world.ForEach(localMask|globalMask|dirtyMask, 0, func(entity ecs.Entity, _ *ecs.Archetype, _ int) {
		scratch.Dirty = append(scratch.Dirty, entity)
		scratch.Seen[entity] = struct{}{}
	})

	if len(scratch.Dirty) == 0 {
		return
	}

	parentMask := ecs.MaskOf[Parent](world)
	world.ForEach(parentMask, 0, func(entity ecs.Entity, table *ecs.Archetype, index int) {
		parents, _ := ecs.Column[Parent](world, table)
		p := &parents[index]
		if p.IsRoot {
			return
		}
		scratch.Children[p.Entity] = append(scratch.Children[p.Entity], entity)
	})

	scratch.Queue = append(scratch.Queue, scratch.Dirty...)
	for len(scratch.Queue) > 0 {
		current := scratch.Queue[0]
		scratch.Queue = scratch.Queue[1:]
		for _, child := range scratch.Children[current] {
			if _, ok := scratch.Seen[child]; ok {
				continue
			}
			scratch.Seen[child] = struct{}{}
			scratch.Dirty = append(scratch.Dirty, child)
			scratch.Queue = append(scratch.Queue, child)
		}
	}

	for _, entity := range scratch.Dirty {
		depth := depthOf(world, entity)
		scratch.Depths = append(scratch.Depths, depth)
		if depth > MaxHierarchyDepth && !scratch.depthCapReported {
			log.Printf("transform: hierarchy depth exceeded MaxHierarchyDepth=%d for entity %v; treating as disconnected", MaxHierarchyDepth, entity)
			scratch.depthCapReported = true
		}
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
