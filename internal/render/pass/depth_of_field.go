package pass

import (
	_ "embed"
	"fmt"
	"unsafe"

	"github.com/cogentcore/webgpu/wgpu"
	"github.com/go-gl/mathgl/mgl32"

	"github.com/matthewjberger/indigo/ecs"
	"github.com/matthewjberger/indigo/render"
)

//go:embed depth_of_field.wgsl
var depthOfFieldShader string

type dofParams struct {
	FocusDistance       float32
	FocusRange          float32
	MaxBlurRadius       float32
	BokehThreshold      float32
	BokehIntensity      float32
	NearPlane           float32
	FarPlane            float32
	SampleCount         uint32
	TextureSize         mgl32.Vec2
	TiltShiftEnabled    uint32
	TiltShiftAngle      float32
	TiltShiftCenter     float32
	TiltShiftBlurAmount float32
	VisualizeTiltShift  uint32
	Enabled             uint32
}

type dofPassState struct {
	pipeline     *wgpu.RenderPipeline
	bgLayout     *wgpu.BindGroupLayout
	sampler      *wgpu.Sampler
	paramsBuffer *wgpu.Buffer
	bindGroup    *wgpu.BindGroup
}

func AddDoFPass(renderer *render.Renderer, aspect func() float32) (*render.Pass, render.ResourceID, error) {
	_ = aspect
	outputID := renderer.Graph.AddColorTexture(render.ResourceDescriptor{
		Name: "dof_color",
		Kind: render.ResourceKindTransientColor,
		Texture: render.TextureDescriptor{
			Format: render.HdrFormat,
			Width:  renderer.Config.Width,
			Height: renderer.Config.Height,
			Usage:  wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageTextureBinding,
		},
	})

	state, err := newDoFState(renderer.Device)
	if err != nil {
		return nil, 0, err
	}
	pass := &render.Pass{
		Name:                 "dof",
		Reads:                []string{"input", "depth"},
		Writes:               []string{"output"},
		Prepare:              func(c *render.PassContext) error { return dofPrepare(state, c) },
		Execute:              func(c *render.PassContext) error { return dofExecute(state, c) },
		InvalidateBindGroups: func() { dofInvalidate(state) },
		Release:              func() { dofRelease(state) },
	}
	if err := renderer.Graph.AddPass(pass, []render.SlotBinding{
		{Slot: "input", ResourceID: renderer.SceneColorID},
		{Slot: "depth", ResourceID: renderer.DepthID},
		{Slot: "output", ResourceID: outputID},
	}); err != nil {
		return nil, 0, err
	}
	return pass, outputID, nil
}

func newDoFState(device *wgpu.Device) (*dofPassState, error) {
	module, err := device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label:          "dof shader",
		WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{Code: depthOfFieldShader},
	})
	if err != nil {
		return nil, fmt.Errorf("dof: shader: %w", err)
	}
	defer module.Release()

	bgLayout, err := device.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "dof bg layout",
		Entries: []wgpu.BindGroupLayoutEntry{
			{Binding: 0, Visibility: wgpu.ShaderStageFragment, Texture: wgpu.TextureBindingLayout{SampleType: wgpu.TextureSampleTypeFloat, ViewDimension: wgpu.TextureViewDimension2D}},
			{Binding: 1, Visibility: wgpu.ShaderStageFragment, Texture: wgpu.TextureBindingLayout{SampleType: wgpu.TextureSampleTypeDepth, ViewDimension: wgpu.TextureViewDimension2D}},
			{Binding: 2, Visibility: wgpu.ShaderStageFragment, Sampler: wgpu.SamplerBindingLayout{Type: wgpu.SamplerBindingTypeFiltering}},
			{Binding: 3, Visibility: wgpu.ShaderStageFragment, Buffer: wgpu.BufferBindingLayout{Type: wgpu.BufferBindingTypeUniform}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dof: bg layout: %w", err)
	}

	pipelineLayout, err := device.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "dof pipeline layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{bgLayout},
	})
	if err != nil {
		return nil, fmt.Errorf("dof: pipeline layout: %w", err)
	}
	defer pipelineLayout.Release()

	pipeline, err := device.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:       "dof pipeline",
		Layout:      pipelineLayout,
		Vertex:      wgpu.VertexState{Module: module, EntryPoint: "vertex_main"},
		Primitive:   wgpu.PrimitiveState{Topology: wgpu.PrimitiveTopologyTriangleList, CullMode: wgpu.CullModeNone},
		Multisample: wgpu.MultisampleState{Count: 1, Mask: 0xFFFFFFFF},
		Fragment: &wgpu.FragmentState{
			Module:     module,
			EntryPoint: "fragment_main",
			Targets:    []wgpu.ColorTargetState{{Format: render.HdrFormat, WriteMask: wgpu.ColorWriteMaskAll}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("dof: pipeline: %w", err)
	}

	sampler, err := device.CreateSampler(&wgpu.SamplerDescriptor{
		Label:         "dof sampler",
		AddressModeU:  wgpu.AddressModeClampToEdge,
		AddressModeV:  wgpu.AddressModeClampToEdge,
		AddressModeW:  wgpu.AddressModeClampToEdge,
		MagFilter:     wgpu.FilterModeLinear,
		MinFilter:     wgpu.FilterModeLinear,
		MipmapFilter:  wgpu.MipmapFilterModeNearest,
		MaxAnisotropy: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("dof: sampler: %w", err)
	}

	paramsBuffer, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "dof params",
		Size:  uint64(unsafe.Sizeof(dofParams{})),
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return nil, fmt.Errorf("dof: params buffer: %w", err)
	}

	return &dofPassState{
		pipeline:     pipeline,
		bgLayout:     bgLayout,
		sampler:      sampler,
		paramsBuffer: paramsBuffer,
	}, nil
}

func dofPrepare(state *dofPassState, context *render.PassContext) error {
	width, height, err := ssaoSize(context, "input")
	if err != nil {
		return err
	}

	settings := render.DefaultGraphics().DepthOfField
	if g, ok := ecs.Resource[render.Graphics](context.World); ok && g != nil {
		settings = g.DepthOfField
	}
	enabled := uint32(0)
	if settings.Enabled {
		enabled = 1
	}

	near := float32(0.1)
	far := float32(1000.0)
	if camera, ok := ecs.Resource[render.Camera](context.World); ok && camera != nil {
		near = camera.Near
		far = camera.Far
	}

	params := dofParams{
		FocusDistance:  settings.FocusDistance,
		FocusRange:     settings.FocusRange,
		MaxBlurRadius:  settings.MaxBlurRadius,
		BokehThreshold: settings.BokehThreshold,
		BokehIntensity: settings.BokehIntensity,
		NearPlane:      near,
		FarPlane:       far,
		SampleCount:    settings.SampleCount,
		TextureSize:    mgl32.Vec2{float32(width), float32(height)},
		Enabled:        enabled,
	}
	writeBuffer(context.Device, context.Queue, context.Encoder, state.paramsBuffer, 0, bytesOf(&params))

	if state.bindGroup != nil {
		return nil
	}
	colorView, err := context.TextureView("input")
	if err != nil {
		return err
	}
	depthView, err := context.TextureView("depth")
	if err != nil {
		return err
	}
	bg, err := context.Device.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "dof bg",
		Layout: state.bgLayout,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, TextureView: colorView},
			{Binding: 1, TextureView: depthView},
			{Binding: 2, Sampler: state.sampler},
			{Binding: 3, Buffer: state.paramsBuffer, Offset: 0, Size: uint64(unsafe.Sizeof(dofParams{}))},
		},
	})
	if err != nil {
		return fmt.Errorf("dof: bind group: %w", err)
	}
	state.bindGroup = bg
	return nil
}

func dofExecute(state *dofPassState, context *render.PassContext) error {
	attachment, err := context.ColorAttachment("output")
	if err != nil {
		return err
	}
	passEnc := context.Encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label:            "dof",
		ColorAttachments: []wgpu.RenderPassColorAttachment{attachment},
	})
	passEnc.SetPipeline(state.pipeline)
	passEnc.SetBindGroup(0, state.bindGroup, nil)
	passEnc.Draw(3, 1, 0, 0)
	passEnc.End()
	passEnc.Release()
	return nil
}

func dofInvalidate(state *dofPassState) {
	if state.bindGroup != nil {
		state.bindGroup.Release()
		state.bindGroup = nil
	}
}

func dofRelease(state *dofPassState) {
	if state.bindGroup != nil {
		state.bindGroup.Release()
	}
	if state.paramsBuffer != nil {
		state.paramsBuffer.Release()
	}
	if state.sampler != nil {
		state.sampler.Release()
	}
	if state.pipeline != nil {
		state.pipeline.Release()
	}
	if state.bgLayout != nil {
		state.bgLayout.Release()
	}
}
