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

// Material uniforms
uniform sampler2D uDiffuseTexture;   // Diffuse texture
uniform bool uUseTexture;            // Whether to use texture

// Material properties
struct Material {
    vec3 diffuse;      // Diffuse color
    vec3 specular;     // Specular color
    float shininess;   // Specular shininess
    float opacity;     // Material opacity
};

uniform Material material;

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

// Function to calculate directional light contribution
vec3 calculateDirectionalLight(Light light, vec3 normal, vec3 viewDir, vec3 baseColor) {
    vec3 lightDir = normalize(-light.direction);

    // Diffuse lighting
    float diff = max(dot(normal, lightDir), 0.0);
    vec3 diffuse = diff * light.color * baseColor;

    // Specular lighting (Blinn-Phong)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), material.shininess);
    vec3 specular = spec * light.color * material.specular;

    return diffuse + specular;
}

// Function to calculate point light contribution
vec3 calculatePointLight(Light light, vec3 normal, vec3 viewDir, vec3 baseColor) {
    vec3 lightDir = normalize(light.position - fragPos);
    float distance = length(light.position - fragPos);

    // Attenuation
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));

    // Diffuse lighting
    float diff = max(dot(normal, lightDir), 0.0);
    vec3 diffuse = diff * light.color * baseColor;

    // Specular lighting (Blinn-Phong)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), material.shininess);
    vec3 specular = spec * light.color * material.specular;

    // Apply attenuation
    diffuse *= attenuation;
    specular *= attenuation;

    return diffuse + specular;
}

// Function to calculate spot light contribution
vec3 calculateSpotLight(Light light, vec3 normal, vec3 viewDir, vec3 baseColor) {
    vec3 lightDir = normalize(light.position - fragPos);
    float distance = length(light.position - fragPos);

    // Check if fragment is within the spotlight cone
    float theta = dot(lightDir, normalize(-light.direction));
    float epsilon = light.innerCone - light.outerCone;
    float intensity = clamp((theta - light.outerCone) / epsilon, 0.0, 1.0);

    if (intensity <= 0.0) {
        return vec3(0.0); // Outside the cone
    }

    // Attenuation
    float attenuation = 1.0 / (light.constant + light.linear * distance + light.quadratic * (distance * distance));

    // Diffuse lighting
    float diff = max(dot(normal, lightDir), 0.0);
    vec3 diffuse = diff * light.color * baseColor;

    // Specular lighting (Blinn-Phong)
    vec3 halfwayDir = normalize(lightDir + viewDir);
    float spec = pow(max(dot(normal, halfwayDir), 0.0), material.shininess);
    vec3 specular = spec * light.color * material.specular;

    // Apply attenuation and spot light intensity
    diffuse *= attenuation * intensity;
    specular *= attenuation * intensity;

    return diffuse + specular;
}

void main() {
    // Get base color from texture or material
    vec3 baseColor;
    if (uUseTexture) {
        vec4 textureColor = texture(uDiffuseTexture, fragTexCoord);
        baseColor = textureColor.rgb * material.diffuse;
        // Handle texture alpha for transparency
        if (textureColor.a < 0.1) {
            discard;
        }
    } else {
        baseColor = material.diffuse;
    }

    // Normalize vectors
    vec3 normal = normalize(fragNormal);
    vec3 viewDir = normalize(viewPos - fragPos);

    // Start with ambient lighting
    vec3 result = uAmbientColor * baseColor;

    // Calculate contribution from each light
    for (int i = 0; i < uNumLights && i < MAX_LIGHTS; i++) {
        Light light = uLights[i];

        if (light.type == DIRECTIONAL_LIGHT) {
            result += calculateDirectionalLight(light, normal, viewDir, baseColor);
        } else if (light.type == POINT_LIGHT) {
            result += calculatePointLight(light, normal, viewDir, baseColor);
        } else if (light.type == SPOT_LIGHT) {
            result += calculateSpotLight(light, normal, viewDir, baseColor);
        }
    }

    // Apply gamma correction
    result = pow(result, vec3(1.0 / 2.2));

    // Ensure we don't exceed maximum brightness
    result = min(result, vec3(1.0));

    // Output final color with material opacity
    FragColor = vec4(result, material.opacity);
}