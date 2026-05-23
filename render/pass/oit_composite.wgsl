// Weighted-blended OIT resolve. Reads the accumulation buffer
// (sum of weight * color, alpha = sum of weight) and the reveal
// buffer (product of (1 - alpha) for every fragment, kept on the
// hardware via Zero+OneMinusSrcAlpha blend). Composites the
// resulting transparent color back into scene_color via a
// standard SrcAlpha / OneMinusSrcAlpha blend so the opaque pass's
// result becomes the background.

@group(0) @binding(0) var accum_texture: texture_2d<f32>;
@group(0) @binding(1) var accum_sampler: sampler;
@group(0) @binding(2) var reveal_texture: texture_2d<f32>;
@group(0) @binding(3) var reveal_sampler: sampler;

struct VertexOutput {
    @builtin(position) clip_position: vec4<f32>,
    @location(0) uv: vec2<f32>,
};

@vertex
fn vertex_main(@builtin(vertex_index) vertex_index: u32) -> VertexOutput {
    let x = f32((vertex_index << 1u) & 2u);
    let y = f32(vertex_index & 2u);
    var out: VertexOutput;
    out.clip_position = vec4<f32>(x * 2.0 - 1.0, y * 2.0 - 1.0, 0.0, 1.0);
    out.uv = vec2<f32>(x, 1.0 - y);
    return out;
}

@fragment
fn fragment_main(in: VertexOutput) -> @location(0) vec4<f32> {
    let accum = textureSample(accum_texture, accum_sampler, in.uv);
    let reveal = textureSample(reveal_texture, reveal_sampler, in.uv).r;
    // accum.a holds the sum of weights; divide rgb by it to get
    // weighted-average color. Out-alpha is 1 - reveal_product.
    let avg = accum.rgb / max(accum.a, 1e-5);
    let out_alpha = clamp(1.0 - reveal, 0.0, 1.0);
    return vec4<f32>(avg, out_alpha);
}
