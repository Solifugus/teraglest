#version 330 core

// Input from vertex shader
in vec3 fragPos;         // World space position
in vec3 fragNormal;      // World space normal
in vec2 fragTexCoord;    // Texture coordinates
in float fragSurfaceType; // Surface type index

// Uniforms
uniform sampler2D uSurface1;     // Grass texture
uniform sampler2D uSurface2;     // Secondary grass texture
uniform sampler2D uSurface3;     // Road texture
uniform sampler2D uSurface4;     // Stone texture
uniform sampler2D uSurface5;     // Ground texture

// Lighting uniforms
uniform vec3 uLightDirection;    // Directional light direction (normalized)
uniform vec3 uLightColor;        // Light color
uniform vec3 uAmbientColor;      // Ambient light color

// Fog uniforms
uniform bool uUseFog;
uniform float uFogDensity;
uniform vec3 uFogColor;
uniform float uFogStart;
uniform float uFogEnd;

// Output
out vec4 FragColor;

void main() {
    // Sample appropriate texture based on surface type
    vec3 textureColor;
    int surfaceIndex = int(fragSurfaceType);

    switch (surfaceIndex) {
        case 1: // Grass
            textureColor = texture(uSurface1, fragTexCoord).rgb;
            break;
        case 2: // Secondary grass
            textureColor = texture(uSurface2, fragTexCoord).rgb;
            break;
        case 3: // Road
            textureColor = texture(uSurface3, fragTexCoord).rgb;
            break;
        case 4: // Stone
            textureColor = texture(uSurface4, fragTexCoord).rgb;
            break;
        case 5: // Ground
            textureColor = texture(uSurface5, fragTexCoord).rgb;
            break;
        default: // Default to grass
            textureColor = texture(uSurface1, fragTexCoord).rgb;
            break;
    }

    // Normalize the normal vector
    vec3 normal = normalize(fragNormal);

    // Ambient lighting
    vec3 ambient = uAmbientColor * textureColor;

    // Diffuse lighting (Lambert)
    float diffuseStrength = max(dot(normal, -uLightDirection), 0.0);
    vec3 diffuse = diffuseStrength * uLightColor * textureColor;

    // Combine lighting components
    vec3 result = ambient + diffuse;

    // Apply fog if enabled
    if (uUseFog) {
        float distance = length(fragPos);
        float fogFactor = clamp((uFogEnd - distance) / (uFogEnd - uFogStart), 0.0, 1.0);
        result = mix(uFogColor, result, fogFactor);
    }

    // Apply gamma correction
    result = pow(result, vec3(1.0 / 2.2));

    // Output final color
    FragColor = vec4(result, 1.0);
}