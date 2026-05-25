
struct VertexInput {
    @location(0) position: vec4<f32>,
    @location(1) normal:   vec4<f32>,
    @location(2) tangent:  vec4<f32>,
    @location(3) uv:       vec4<f32>,
    @location(4) color:    vec4<f32>,
};

struct VertexOutput {
    @builtin(position) clip_position: vec4<f32>,
    @location(0) @interpolate(flat) material_index: u32,
    @location(1) uv: vec4<f32>,
    @location(2) color: vec4<f32>,
    @location(3) @interpolate(flat) flip_winding: u32,
};


@group(0) @binding(0) var<uniform> view_proj: mat4x4<f32>;

struct MorphDisplacement {
    position: vec3<f32>,
    _pad0: f32,
    normal: vec3<f32>,
    _pad1: f32,
    tangent: vec3<f32>,
    _pad2: f32,
};

struct MorphInstance {
    weights: array<f32, 8>,
    target_count: u32,
    vertex_count: u32,
    _mpad0: u32,
    _mpad1: u32,
};

@group(1) @binding(0) var<storage, read> models:           array<mat4x4<f32>>;
@group(1) @binding(1) var<storage, read> material_indices: array<u32>;
@group(1) @binding(2) var<storage, read> entity_ids:       array<u32>;
@group(1) @binding(3) var<storage, read> visible_indices:  array<u32>;
@group(1) @binding(4) var<storage, read> morph_displacements: array<MorphDisplacement>;
@group(1) @binding(5) var<storage, read> morph_instances:     array<MorphInstance>;

@group(2) @binding(0) var<storage, read> materials: array<Material>;

@group(3) @binding(0) var material_srgb_array: texture_2d_array<f32>;
@group(3) @binding(1) var material_sampler: sampler;

const NO_LAYER: u32 = 0xFFFFFFFFu;

fn apply_wrap_axis(value: f32, mode: u32) -> f32 {
    if (mode == 0u) {
        return fract(value);
    } else if (mode == 1u) {
        let cycle = value - 2.0 * floor(value * 0.5);
        return min(cycle, 2.0 - cycle);
    }
    return clamp(value, 0.0, 1.0);
}

fn apply_wrap(uv: vec2<f32>, packed: u32) -> vec2<f32> {
    let mode_u = (packed >> 16u) & 0x3u;
    let mode_v = (packed >> 18u) & 0x3u;
    return vec2<f32>(apply_wrap_axis(uv.x, mode_u), apply_wrap_axis(uv.y, mode_v));
}

fn texture_uv(uv: vec4<f32>, transform: TextureTransform) -> vec2<f32> {
    let coords = select(uv.xy, uv.zw, u32(transform.row0.w) == 1u);
    let h = vec3<f32>(coords, 1.0);
    return vec2<f32>(dot(transform.row0.xyz, h), dot(transform.row1.xyz, h));
}

@vertex
fn vertex_main(input: VertexInput, @builtin(instance_index) instance_index: u32, @builtin(vertex_index) vertex_index: u32) -> VertexOutput {
    let slot = visible_indices[instance_index];
    let model = models[slot];
    var local_position = input.position.xyz;
    var morph = morph_instances[slot];
    if (morph.target_count > 0u) {
        for (var t = 0u; t < morph.target_count; t = t + 1u) {
            let w = morph.weights[t];
            if (abs(w) > 0.0001) {
                local_position = local_position + morph_displacements[t * morph.vertex_count + vertex_index].position * w;
            }
        }
    }
    var out: VertexOutput;
    let world = model * vec4<f32>(local_position, 1.0);
    out.clip_position = view_proj * world;
    out.material_index = material_indices[slot];
    out.uv = input.uv;
    out.color = input.color;
    let model_det = dot(model[0].xyz, cross(model[1].xyz, model[2].xyz));
    out.flip_winding = select(0u, 1u, model_det < 0.0);
    return out;
}

@fragment
fn fragment_main(in: VertexOutput, @builtin(front_facing) front_facing: bool) {
    let mat = materials[in.material_index];
    let geometric_front = front_facing != (in.flip_winding != 0u);
    if (mat.double_sided == 0u && !geometric_front) {
        discard;
    }
    if (mat.alpha_mode == 2u || mat.transmission_factor > 0.0) {
        discard;
    }
    if (mat.alpha_mode == 1u) {
        var a = mat.base_color.a * in.color.a;
        if (mat.base_layer != NO_LAYER) {
            let packed = mat.base_layer;
            let layer = i32(packed & 0xFFFFu);
            let wrapped = apply_wrap(texture_uv(in.uv, mat.base_transform), packed);
            a = a * textureSampleLevel(material_srgb_array, material_sampler, wrapped, layer, 0.0).a;
        }
        if (a < mat.alpha_cutoff) {
            discard;
        }
    }
}
