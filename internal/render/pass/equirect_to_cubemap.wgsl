struct EquirectUniform {
    output_size: u32,
    _pad0:       u32,
    _pad1:       u32,
    _pad2:       u32,
};

@group(0) @binding(0)
var<uniform> params: EquirectUniform;

@group(0) @binding(1)
var equirect_texture: texture_2d<f32>;

@group(0) @binding(2)
var equirect_sampler: sampler;

@group(0) @binding(3)
var output_texture: texture_storage_2d_array<rgba16float, write>;

const PI: f32 = 3.141592653589793;

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

fn world_to_equirect(dir: vec3<f32>) -> vec2<f32> {
    let phi = atan2(dir.z, dir.x);
    let theta = asin(dir.y);

    var uv = vec2<f32>(phi / (2.0 * PI), theta / PI);
    uv.x = uv.x + 0.5;
    uv.y = 0.5 - uv.y;
    return uv;
}

@compute @workgroup_size(16, 16, 1)
fn main(@builtin(global_invocation_id) global_id: vec3<u32>) {
    let coords = global_id.xy;
    let face = global_id.z;

    if (coords.x >= params.output_size || coords.y >= params.output_size || face >= 6u) {
        return;
    }

    let uv = (vec2<f32>(coords) + 0.5) / f32(params.output_size);
    let dir = cube_to_world(face, uv);
    let equirect_uv = world_to_equirect(dir);
    let color = textureSampleLevel(equirect_texture, equirect_sampler, equirect_uv, 0.0);

    textureStore(output_texture, coords, face, color);
}
