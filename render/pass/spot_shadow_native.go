//go:build !js

package pass

import (
	"github.com/cogentcore/webgpu/wgpu"

	"indigo/ecs"
	"indigo/render"
	"indigo/render/asset"
)

func spotShadowExecute(s any, context *render.PassContext) error {
	state := s.(*spotShadowPassState)
	shadow := state.shadow
	if shadow.ActiveCount == 0 {
		return nil
	}

	meshState, ok := findMeshPassState(context.World)
	if !ok {
		return nil
	}
	assets := ecs.MustResource[asset.MeshAssetsResource](context.World).Assets
	lightMask := ecs.MustMaskOf[render.Light](context.World)

	for index := uint32(0); index < shadow.ActiveCount; index++ {
		slotX := (index % SpotShadowSlotsPerRow) * SpotShadowSlotSize
		slotY := (index / SpotShadowSlotsPerRow) * SpotShadowSlotSize
		var loadOp wgpu.LoadOp
		if index == 0 {
			loadOp = wgpu.LoadOpClear
		} else {
			loadOp = wgpu.LoadOpLoad
		}
		passEnc := context.Encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
			Label: "spot shadow slot",
			DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
				View:            shadow.AtlasView,
				DepthLoadOp:     loadOp,
				DepthStoreOp:    wgpu.StoreOpStore,
				DepthClearValue: 1.0,
			},
		})
		passEnc.SetPipeline(state.pipeline)
		passEnc.SetViewport(float32(slotX), float32(slotY), float32(SpotShadowSlotSize), float32(SpotShadowSlotSize), 0, 1)
		passEnc.SetScissorRect(slotX, slotY, SpotShadowSlotSize, SpotShadowSlotSize)
		passEnc.SetBindGroup(0, state.slotBgs[index], nil)
		for _, handle := range meshState.sortedHandles {
			bucket := meshState.perHandle[handle]
			entry, ok := assets.Lookup(handle)
			if !ok {
				continue
			}
			shadowBg, err := ensureShadowHandleBindGroup(bucket, context.Device, state.handleBgLayout)
			if err != nil {
				passEnc.End()
				passEnc.Release()
				return err
			}
			passEnc.SetBindGroup(1, shadowBg, nil)
			passEnc.SetVertexBuffer(0, entry.Vertices, 0, wgpu.WholeSize)
			drawNonLightInstances(passEnc, bucket, entry.VertexCount, lightMask, context.World)
		}
		passEnc.End()
		passEnc.Release()
	}
	return nil
}
