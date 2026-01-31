#version 330 core

// Maximum number of lights supported
#define MAX_LIGHTS 8

// Light types
#define DIRECTIONAL_LIGHT 0
#define POINT_LIGHT 1
#define SPOT_LIGHT 2

// Input from vertex shader
in vec3 fragPos;        // World space position
in vec3 fragNormal;     // World space normal
in vec2 fragTexCoord;   // Texture coordinates
in vec3 viewPos;        // Camera position in world space

// Tangent space vectors for normal mapping
in vec3 tangentLightPos[8];  // Light positions in tangent space
in vec3 tangentViewPos;      // Camera position in tangent space
in vec3 tangentFragPos;      // Fragment position in tangent space

// Advanced material properties
struct Material {
    vec3 diffuse;         // Diffuse color
    vec3 specular;        // Specular color
    vec3 emissive;        // Emissive color
    float shininess;      // Specular shininess
    float metallic;       // Metallic factor (0-1)
    float roughness;      // Surface roughness (0-1)
    float opacity;        // Material opacity
    float normalStrength; // Normal map intensity
    float emissiveStrength; // Emissive intensity
    vec2 textureScale;    // UV scale
    vec2 textureOffset;   // UV offset
};

uniform Material material;

// Texture uniforms
uniform sampler2D uDiffuseTexture;
uniform sampler2D uNormalTexture;
uniform sampler2D uSpecularTexture;
uniform sampler2D uEmissiveTexture;

// Texture usage flags
uniform bool uUseDiffuseTexture;
uniform bool uUseNormalTexture;
uniform bool uUseSpecularTexture;
uniform bool uUseEmissiveTexture;

// Light structure
struct Light {
    int type;          // Light type (0=directional, 1=point, 2=spot)
    vec3 position;     // Position (for point and spot lights)
    vec3 direction;    // Direction (for directional and spot lights)
    vec3 color;        // Light color * intensity

    // Attenuation (for point and spot lights)
    float constant;
    float linear;
    float quadratic;

    // Spot light properties
    float innerCone;   // Inner cone cosine
    float outerCone;   // Outer cone cosine
};

// Lighting uniforms
uniform int uNumLights;              // Number of active lights
uniform Light uLights[MAX_LIGHTS];   // Array of lights
uniform vec3 uAmbientColor;          // Global ambient lighting

// Output
out vec4 FragColor;

// Function to calculate directional light contribution with normal mapping
vec3 calculateDirectionalLightNM(Light light, vec3 normal, vec3 viewDir, vec3 baseColor, vec3 specularColor, float shininess) {
    vec3 lightDir = normalize(-light.direction);

    // Diffuse lighting
    float diff = max(dot(normal, lightDir), 0.0);
    vec3 diffuse = diff * light.color * baseColor;

    // Specular lighting (Blinn-Phong)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), shininess);
    vec3 specular = spec * light.color * specularColor;

    return diffuse + specular;
}

// Function to calculate point light contribution with normal mapping
vec3 calculatePointLightNM(Light light, vec3 normal, vec3 viewDir, vec3 baseColor, vec3 specularColor, float shininess, vec3 fragPos) {
    vec3 lightDir = normalize(light.position - fragPos);
    float distance = length(light.position - fragPos);

    // Attenuation
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));

    // Diffuse lighting
    float diff = max(dot(normal, lightDir), 0.0);
    vec3 diffuse = diff * light.color * baseColor;

    // Specular lighting (Blinn-Phong)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), shininess);
    vec3 specular = spec * light.color * specularColor;

    // Apply attenuation
    diffuse *= attenuation;
    specular *= attenuation;

    return diffuse + specular;
}

void main() {
    // Sample textures
    vec3 baseColor = material.diffuse;
    if (uUseDiffuseTexture) {
        vec4 diffuseSample = texture(uDiffuseTexture, fragTexCoord);
        baseColor = diffuseSample.rgb * material.diffuse;
        // Handle texture alpha for transparency
        if (diffuseSample.a < 0.1) {
            discard;
        }
    }

    // Sample normal map and calculate normal in world space
    vec3 normal;
    if (uUseNormalTexture) {
        // Sample normal map
        vec3 normalMap = texture(uNormalTexture, fragTexCoord).rgb;
        normalMap = normalize(normalMap * 2.0 - 1.0); // Convert from [0,1] to [-1,1]

        // Apply normal strength
        normalMap.xy *= material.normalStrength;
        normalMap = normalize(normalMap);

        // Use the normal map in tangent space
        normal = normalMap;
    } else {
        // Use vertex normal (in tangent space, this would be [0,0,1])
        normal = vec3(0.0, 0.0, 1.0);
    }

    // Sample specular map
    vec3 specularColor = material.specular;
    if (uUseSpecularTexture) {
        specularColor = texture(uSpecularTexture, fragTexCoord).rgb * material.specular;
    }

    // Calculate view direction in tangent space
    vec3 viewDir = normalize(tangentViewPos - tangentFragPos);

    // Start with ambient lighting
    vec3 result = uAmbientColor * baseColor;

    // Calculate contribution from each light in tangent space
    for (int i = 0; i < uNumLights && i < MAX_LIGHTS; i++) {
        Light light = uLights[i];

        if (light.type == DIRECTIONAL_LIGHT) {
            // Transform light direction to tangent space
            vec3 lightDirTS = normalize(-light.direction); // This needs proper transformation
            result += calculateDirectionalLightNM(light, normal, viewDir, baseColor, specularColor, material.shininess);
        } else if (light.type == POINT_LIGHT) {
            // Use tangent space light position
            result += calculatePointLightNM(light, normal, viewDir, baseColor, specularColor, material.shininess, tangentFragPos);
        }
        // Note: Spot lights would need similar tangent space handling
    }

    // Add emissive contribution
    if (uUseEmissiveTexture) {
        vec3 emissive = texture(uEmissiveTexture, fragTexCoord).rgb * material.emissive * material.emissiveStrength;
        result += emissive;
    } else if (material.emissiveStrength > 0.0) {
        result += material.emissive * material.emissiveStrength;
    }

    // Apply gamma correction
    result = pow(result, vec3(1.0 / 2.2));

    // Ensure we don't exceed maximum brightness
    result = min(result, vec3(1.0));

    // Output final color with material opacity
    FragColor = vec4(result, material.opacity);
}