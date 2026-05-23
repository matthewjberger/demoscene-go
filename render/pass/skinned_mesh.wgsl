// Skinned mesh shader. Vertex stage blends position + normal by
// up to 4 joint matrices per vertex; fragment stage does a basic
// directional + ambient lit shading (the full PBR + shadow path
// is a follow-up). Output lands in scene_color so the postprocess
// chain (SSAO, bloom, tonemap) still applies.

struct ViewProj {
    view_proj: mat4x4<f32>,
};

@group(0) @binding(0) var<uniform> view_proj_uniform: ViewProj;

struct SkinnedUniforms {
    light_direction: vec4<f32>,
    light_color:     vec4<f32>,
    ambient_color:   vec4<f32>,
};

@group(1) @binding(0) var<uniform>       skinned_uniforms: SkinnedUniforms;
@group(2) @binding(0) var<storage, read> models:           array<mat4x4<f32>>;
@group(2) @binding(1) var<storage, read> joint_matrices:   array<mat4x4<f32>>;
@group(2) @binding(2) var<storage, read> entity_ids:       array<u32>;

struct VertexInput {
    @location(0) position:      vec4<f32>,
    @location(1) normal:        vec4<f32>,
    @location(2) tangent:       vec4<f32>,
    @location(3) uv:            vec4<f32>,
    @location(4) color:         vec4<f32>,
    @location(5) joint_indices: vec4<u32>,
    @location(6) joint_weights: vec4<f32>,
};

struct VertexOutput {
    @builtin(position) clip_position: vec4<f32>,
    @location(0) world_normal: vec3<f32>,
    @location(1) color: vec4<f32>,
    @location(2) @interpolate(flat) entity_id: u32,
};

struct FragmentOutput {
    @location(0) color:       vec4<f32>,
    @location(1) entity_id:   u32,
    @location(2) view_normal: vec4<f32>,
};

@vertex
fn vertex_main(input: VertexInput, @builtin(instance_index) instance_index: u32) -> VertexOutput {
    let model = models[instance_index];

    // The skin matrix is the per-vertex weighted blend of up to
    // four joint matrices. Each joint matrix is the joint's world
    // transform composed with its inverse-bind matrix, so applying
    // it to the bind-pose vertex gives the animated world position.
    let skin =
        input.joint_weights.x * joint_matrices[input.joint_indices.x] +
        input.joint_weights.y * joint_matrices[input.joint_indices.y] +
        input.joint_weights.z * joint_matrices[input.joint_indices.z] +
        input.joint_weights.w * joint_matrices[input.joint_indices.w];

    let world = model * skin * vec4<f32>(input.position.xyz, 1.0);
    let world_normal = (model * skin * vec4<f32>(input.normal.xyz, 0.0)).xyz;

    var out: VertexOutput;
    out.clip_position = view_proj_uniform.view_proj * world;
    out.world_normal = world_normal;
    out.color = input.color;
    out.entity_id = entity_ids[instance_index];
    return out;
}

@fragment
fn fragment_main(in: VertexOutput) -> FragmentOutput {
    let normal = normalize(in.world_normal);
    let light_dir = -normalize(skinned_uniforms.light_direction.xyz);
    let lambert = max(dot(normal, light_dir), 0.0);
    let lit = skinned_uniforms.ambient_color.rgb + skinned_uniforms.light_color.rgb * lambert;
    var out: FragmentOutput;
    out.color = vec4<f32>(in.color.rgb * lit, in.color.a);
    out.entity_id = in.entity_id;
    out.view_normal = vec4<f32>(normal, 1.0);
    return out;
}
