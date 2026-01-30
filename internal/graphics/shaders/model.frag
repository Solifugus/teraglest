#version 330 core

// Input from vertex shader
in vec3 fragPos;        // World space position
in vec3 fragNormal;     // World space normal
in vec2 fragTexCoord;   // Texture coordinates

// Uniforms
uniform sampler2D uDiffuseTexture;   // Diffuse texture
uniform bool uUseTexture;            // Whether to use texture
uniform vec3 uDiffuseColor;          // Diffuse material color
uniform vec3 uSpecularColor;         // Specular material color
uniform float uSpecularPower;        // Specular shininess
uniform float uOpacity;              // Material opacity

// Lighting uniforms
uniform vec3 uLightDirection;        // Directional light direction (normalized)
uniform vec3 uLightColor;            // Light color
uniform vec3 uAmbientColor;          // Ambient light color

// Output
out vec4 FragColor;

void main() {
    // Sample texture or use solid color
    vec3 baseColor;
    if (uUseTexture) {
        baseColor = texture(uDiffuseTexture, fragTexCoord).rgb * uDiffuseColor;
    } else {
        baseColor = uDiffuseColor;
    }

    // Normalize the normal vector
    vec3 normal = normalize(fragNormal);

    // Ambient lighting
    vec3 ambient = uAmbientColor * baseColor;

    // Diffuse lighting (Lambert)
    float diffuseStrength = max(dot(normal, -uLightDirection), 0.0);
    vec3 diffuse = diffuseStrength * uLightColor * baseColor;

    // Specular lighting (Blinn-Phong)
    vec3 viewDir = normalize(-fragPos); // Assume camera at origin for simplicity
    vec3 halfwayDir = normalize(-uLightDirection + viewDir);
    float specularStrength = pow(max(dot(normal, halfwayDir), 0.0), uSpecularPower);
    vec3 specular = specularStrength * uLightColor * uSpecularColor;

    // Combine lighting components
    vec3 result = ambient + diffuse + specular;

    // Apply gamma correction (simple)
    result = pow(result, vec3(1.0 / 2.2));

    // Output final color with alpha
    FragColor = vec4(result, uOpacity);
}