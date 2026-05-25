package pass

import (
	"regexp"
	"strings"
	"testing"

	"github.com/matthewjberger/indigo/render/asset"
)

// TestMaterialWGSLLayoutMatchesGPU computes the std430 size of the WGSL Material
// struct (the single source in material_struct.wgsl) and asserts it equals the
// Go MaterialGPU size. This catches drift between the shader layout and the Go
// struct that the GPU registry is written from — the failure mode that produced
// the black-transmission bug.
func TestMaterialWGSLLayoutMatchesGPU(t *testing.T) {
	type sa struct{ size, align int }
	types := map[string]sa{
		"f32":              {4, 4},
		"u32":              {4, 4},
		"i32":              {4, 4},
		"vec2<f32>":        {8, 8},
		"vec3<f32>":        {12, 16},
		"vec4<f32>":        {16, 16},
		"TextureTransform": {32, 16},
	}

	start := strings.Index(shaderMaterialStruct, "struct Material {")
	if start < 0 {
		t.Fatal("Material struct not found in material_struct.wgsl")
	}
	body := shaderMaterialStruct[start:]
	body = body[strings.Index(body, "{")+1 : strings.Index(body, "};")]

	fieldRe := regexp.MustCompile(`(?m)^\s*\w+:\s*([\w<>]+)\s*,`)
	matches := fieldRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		t.Fatal("no Material fields parsed")
	}

	alignUp := func(x, a int) int { return (x + a - 1) / a * a }
	offset, maxAlign := 0, 1
	for _, m := range matches {
		info, ok := types[m[1]]
		if !ok {
			t.Fatalf("unhandled WGSL field type %q", m[1])
		}
		offset = alignUp(offset, info.align)
		offset += info.size
		if info.align > maxAlign {
			maxAlign = info.align
		}
	}
	total := alignUp(offset, maxAlign)

	if uint64(total) != asset.MaterialGPUSize {
		t.Fatalf("WGSL Material std430 size = %d, want MaterialGPUSize = %d (layout drifted)", total, asset.MaterialGPUSize)
	}
}
