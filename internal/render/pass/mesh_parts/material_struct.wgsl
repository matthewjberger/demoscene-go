// Single source of truth for the GPU material layout (std430, 880 bytes).
// Composed into common.wgsl (forward/OIT) and the depth prepass so the layout
// is defined exactly once. It must stay byte-identical to MaterialGPU in
// render/asset/material.go (asserted at MaterialGPUSize) and is validated by
// TestMaterialWGSLLayoutMatchesGPU.

struct TextureTransform {
    row0: vec4<f32>,
    row1: vec4<f32>,
};

struct Material {
    base_color:      vec4<f32>,
    emissive_factor: vec3<f32>,
    alpha_mode:      u32,

    base_layer:               u32,
    emissive_layer:           u32,
    normal_layer:             u32,
    metallic_roughness_layer: u32,

    occlusion_layer:    u32,
    normal_scale:       f32,
    occlusion_strength: f32,
    metallic_factor:    f32,

    roughness_factor: f32,
    alpha_cutoff:     f32,
    unlit:            u32,
    ior:              f32,

    emissive_strength: f32,
    double_sided:      u32,
    _pad1b:            f32,
    _pad1c:            f32,

    normal_map_flags:     u32,
    specular_factor:      f32,
    specular_layer:       u32,
    specular_color_layer: u32,

    specular_color_factor: vec3<f32>,
    transmission_factor:   f32,

    transmission_layer:   u32,
    thickness:            f32,
    thickness_layer:      u32,
    attenuation_distance: f32,

    attenuation_color: vec3<f32>,
    dispersion:        f32,

    anisotropy_strength:     f32,
    anisotropy_rotation_cos: f32,
    anisotropy_rotation_sin: f32,
    anisotropy_layer:        u32,

    clearcoat_factor:          f32,
    clearcoat_roughness_factor: f32,
    clearcoat_normal_scale:    f32,
    clearcoat_layer:           u32,

    clearcoat_roughness_layer: u32,
    clearcoat_normal_layer:    u32,
    sheen_color_layer:         u32,
    sheen_roughness_layer:     u32,

    sheen_color_factor:     vec3<f32>,
    sheen_roughness_factor: f32,

    iridescence_factor:        f32,
    iridescence_ior:           f32,
    iridescence_thickness_min: f32,
    iridescence_thickness_max: f32,

    iridescence_layer:           u32,
    iridescence_thickness_layer: u32,
    diffuse_transmission_factor: f32,
    diffuse_transmission_color_layer: u32,

    diffuse_transmission_color_factor: vec3<f32>,
    blend_opaque_alpha_threshold:      f32,

    base_transform:               TextureTransform,
    normal_transform:             TextureTransform,
    metallic_roughness_transform: TextureTransform,
    occlusion_transform:          TextureTransform,
    emissive_transform:           TextureTransform,

    transmission_transform:              TextureTransform,
    thickness_transform:                 TextureTransform,
    specular_transform:                  TextureTransform,
    specular_color_transform:            TextureTransform,
    clearcoat_transform:                 TextureTransform,
    clearcoat_roughness_transform:       TextureTransform,
    clearcoat_normal_transform:          TextureTransform,
    sheen_color_transform:               TextureTransform,
    sheen_roughness_transform:           TextureTransform,
    iridescence_transform:               TextureTransform,
    iridescence_thickness_transform:     TextureTransform,
    anisotropy_transform:                TextureTransform,
    diffuse_transmission_transform:      TextureTransform,
    diffuse_transmission_color_transform: TextureTransform,
};
