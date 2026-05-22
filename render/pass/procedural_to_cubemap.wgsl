// Procedural sky -> cubemap capture. Renders indigo's sky shader
// into a 6-face cubemap so the filter_envmap pass can pre-integrate
// it into irradiance + prefiltered specular maps for IBL. Re-uses
// the exact sky_color function from sky.wgsl so the captured
// cubemap matches what gets drawn behind the scene.
//
// Direct port of nightshade's procedural_to_cubemap.wgsl, scoped
// to the single atmosphere type indigo currently supports.

struct ProceduralUniform {
    output_size: u32,
    time:        f32,
    _pad0:       u32,
    _pad1:       u32,
};

@group(0) @binding(0)
var<uniform> params: ProceduralUniform;

@group(0) @binding(1)
var output_texture: texture_storage_2d_array<rgba16float, write>;

fn cube_to_world(face: u32, uv: vec2<f32>) -> vec3<f32> {
    var dir: vec3<f32>;
    let x = 2.0 * uv.x - 1.0;
    let y = 2.0 * uv.y - 1.0;

    switch face {
        case 0u: {
            dir = vec3<f32>(1.0, -y, -x);
        }
        case 1u: {
            dir = vec3<f32>(-1.0, -y, x);
        }
        case 2u: {
            dir = vec3<f32>(x, 1.0, y);
        }
        case 3u: {
            dir = vec3<f32>(x, -1.0, -y);
        }
        case 4u: {
            dir = vec3<f32>(x, -y, 1.0);
        }
        default: {
            dir = vec3<f32>(-x, -y, -1.0);
        }
    }
    return normalize(dir);
}

// sky_color matches sky.wgsl's fs_sky body so the captured
// cubemap is the same gradient + sun-disk the on-screen sky pass
// renders. Kept inline here (not deduped) because the sky pass
// has its own additional per-frame uniforms; the capture only
// needs the world-space direction.
fn sky_color(dir: vec3<f32>) -> vec3<f32> {
    let sky_top_color = vec3<f32>(0.385, 0.454, 0.55);
    let sky_horizon_color = vec3<f32>(0.646, 0.656, 0.67);
    let ground_horizon_color = vec3<f32>(0.646, 0.656, 0.67);
    let ground_bottom_color = vec3<f32>(0.2, 0.169, 0.133);

    let height = dir.y;

    let sky_curve = 0.15;
    let ground_curve = 0.02;

    var color: vec3<f32>;
    if (height > 0.0) {
        let t = 1.0 - pow(1.0 - height, 1.0 / sky_curve);
        color = mix(sky_horizon_color, sky_top_color, clamp(t, 0.0, 1.0));
    } else {
        let t = 1.0 - pow(1.0 + height, 1.0 / ground_curve);
        color = mix(ground_horizon_color, ground_bottom_color, clamp(t, 0.0, 1.0));
    }

    color = color * 1.3;

    let sun_direction = normalize(vec3<f32>(0.0, 0.5, -1.0));
    let sun_angle = acos(clamp(dot(dir, sun_direction), -1.0, 1.0));
    let sun_disk = 1.0 - smoothstep(0.0, 0.02, sun_angle);
    let sun_color = vec3<f32>(1.0, 0.95, 0.8);
    color = mix(color, sun_color, sun_disk * 0.5);

    return color;
}

@compute @workgroup_size(8, 8, 1)
fn main(@builtin(global_invocation_id) global_id: vec3<u32>) {
    let coords = global_id.xy;
    let face = global_id.z;

    if (coords.x >= params.output_size || coords.y >= params.output_size || face >= 6u) {
        return;
    }

    let uv = vec2<f32>(
        (f32(coords.x) + 0.5) / f32(params.output_size),
        (f32(coords.y) + 0.5) / f32(params.output_size)
    );

    let direction = cube_to_world(face, uv);
    let color = sky_color(direction);

    textureStore(output_texture, coords, face, vec4<f32>(color, 1.0));
}
