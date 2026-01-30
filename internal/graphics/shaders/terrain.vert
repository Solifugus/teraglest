#version 330 core

// Input vertex attributes
layout (location = 0) in vec3 aPosition;    // Terrain vertex position (with height)
layout (location = 1) in vec3 aNormal;      // Calculated terrain normal
layout (location = 2) in vec2 aTexCoord;    // Texture coordinates for tiling
layout (location = 3) in float aSurfaceType; // Surface type index for texture blending

// Transformation matrices
uniform mat4 uView;
uniform mat4 uProjection;

// Output to fragment shader
out vec3 fragPos;        // World space position
out vec3 fragNormal;     // World space normal
out vec2 fragTexCoord;   // Texture coordinates
out float fragSurfaceType; // Surface type for texture selection

void main() {
    // Terrain is in world space already (no model matrix needed)
    fragPos = aPosition;
    fragNormal = normalize(aNormal);
    fragTexCoord = aTexCoord;
    fragSurfaceType = aSurfaceType;

    // Transform to clip space (no model transform for terrain)
    gl_Position = uProjection * uView * vec4(aPosition, 1.0);
}