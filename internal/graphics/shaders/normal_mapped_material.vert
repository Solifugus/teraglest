#version 330 core

// Input vertex attributes
layout (location = 0) in vec3 aPosition;
layout (location = 1) in vec3 aNormal;
layout (location = 2) in vec2 aTexCoord;
layout (location = 3) in vec3 aTangent;   // Tangent vector for normal mapping
layout (location = 4) in vec3 aBitangent; // Bitangent vector for normal mapping

// Transformation matrices
uniform mat4 uModel;
uniform mat4 uView;
uniform mat4 uProjection;
uniform mat3 uNormalMatrix;

// Camera position for lighting calculations
uniform vec3 uViewPosition;

// Material texture transform
uniform vec2 material_textureScale;
uniform vec2 material_textureOffset;

// Output to fragment shader
out vec3 fragPos;        // World space position
out vec3 fragNormal;     // World space normal
out vec2 fragTexCoord;   // Texture coordinates
out vec3 viewPos;        // Camera position in world space

// Tangent space vectors for normal mapping
out vec3 tangentLightPos[8];  // Light positions in tangent space (max 8 lights)
out vec3 tangentViewPos;      // Camera position in tangent space
out vec3 tangentFragPos;      // Fragment position in tangent space

// Light positions (set by renderer)
uniform vec3 uLightPositions[8];
uniform int uNumLights;

void main() {
    // Transform vertex position to world space
    vec4 worldPos = uModel * vec4(aPosition, 1.0);
    fragPos = worldPos.xyz;

    // Transform normal to world space using normal matrix
    fragNormal = normalize(uNormalMatrix * aNormal);

    // Apply texture transform to texture coordinates
    fragTexCoord = aTexCoord * material_textureScale + material_textureOffset;

    // Pass camera position
    viewPos = uViewPosition;

    // Calculate tangent space transformation for normal mapping
    vec3 T = normalize(uNormalMatrix * aTangent);
    vec3 B = normalize(uNormalMatrix * aBitangent);
    vec3 N = normalize(uNormalMatrix * aNormal);

    // Create TBN matrix for tangent space transformation
    mat3 TBN = transpose(mat3(T, B, N));

    // Transform light positions to tangent space
    for (int i = 0; i < uNumLights && i < 8; i++) {
        tangentLightPos[i] = TBN * uLightPositions[i];
    }

    // Transform camera and fragment positions to tangent space
    tangentViewPos = TBN * uViewPosition;
    tangentFragPos = TBN * fragPos;

    // Transform to clip space for final position
    gl_Position = uProjection * uView * worldPos;
}