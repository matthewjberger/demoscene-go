// Cubemap mip-chain generator. Samples mip N-1 at half the
// resolution to fill mip N. Bilinear filtering on the sampler
// produces the standard 2x2 average; running this once per mip
// level builds the full chain that the GGX prefilter pass needs
// for higher-roughness samples.
//
// Direct port of nightshade's cubemap_mipgen.wgsl.

@group(0) @binding(0)
var src_texture: texture_2d_array<f32>;

@group(0) @binding(1)
var src_sampler: sampler;

@group(0) @binding(2)
var dst_texture: texture_storage_2d_array<rgba16float, write>;

struct MipParams {
    dst_size: u32,
    _pad0:    u32,
    _pad1:    u32,
    _pad2:    u32,
}

@group(0) @binding(3)
var<uniform> params: MipParams;

@compute @workgroup_size(8, 8, 1)
fn main(@builtin(global_invocation_id) global_id: vec3<u32>) {
    let coords = global_id.xy;
    let face = global_id.z;

    if (coords.x >= params.dst_size || coords.y >= params.dst_size || face >= 6u) {
        return;
    }

    let texel_size = 1.0 / f32(params.dst_size);
    let uv = (vec2<f32>(coords) + 0.5) * texel_size;

    let color = textureSampleLevel(src_texture, src_sampler, uv, i32(face), 0.0);

    textureStore(dst_texture, vec2<i32>(coords), i32(face), color);
}
