#version 330 core

// Input vertex attributes
layout (location = 0) in vec3 aPosition;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;

// Transformation matrices
uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProjection;
uniform mat3 uNormalMatrix;

// Output to fragment shader
out vec3 fragPos;        // World space position
out vec3 fragNormal;     // World space normal
out vec2 fragTexCoord;   // Texture coordinates

void main() {
    // Transform vertex position to world space
    vec4 worldPos = uModel * vec4(aPosition, 1.0);
    fragPos = worldPos.xyz;

    // Transform normal to world space
    fragNormal = normalize(uNormalMatrix * aNormal);

    // Pass through texture coordinates
    fragTexCoord = aTexCoord;

    // Transform to clip space
    gl_Position = uProjection * uView * worldPos;
}