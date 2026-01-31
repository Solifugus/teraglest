package renderer

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Camera represents a 3D camera for rendering
type Camera struct {
	// Position and orientation
	Position mgl32.Vec3 // Camera position in world space
	Target   mgl32.Vec3 // Point camera is looking at
	Up       mgl32.Vec3 // Up vector

	// Projection parameters
	FOV        float32 // Field of view in radians
	AspectRatio float32 // Width/Height ratio
	NearPlane   float32 // Near clipping plane
	FarPlane    float32 // Far clipping plane

	// View frustum for culling
	ViewMatrix       mgl32.Mat4 // Cached view matrix
	ProjectionMatrix mgl32.Mat4 // Cached projection matrix
	isDirty          bool       // Whether matrices need recalculation
}

// NewCamera creates a new camera with default RTS parameters
func NewCamera(windowWidth, windowHeight int) *Camera {
	camera := &Camera{
		Position:    mgl32.Vec3{0, 15, 10},  // Elevated position for RTS view
		Target:      mgl32.Vec3{0, 0, 0},    // Looking at origin
		Up:          mgl32.Vec3{0, 1, 0},    // Y-up
		FOV:         mgl32.DegToRad(45.0),   // 45-degree field of view
		AspectRatio: float32(windowWidth) / float32(windowHeight),
		NearPlane:   0.1,                    // Close near plane for UI
		FarPlane:    1000.0,                 // Far plane for large maps
		isDirty:     true,
	}

	// Calculate initial matrices
	camera.updateMatrices()
	return camera
}

// NewRTSCamera creates a camera optimized for RTS gameplay
func NewRTSCamera(windowWidth, windowHeight int, mapSize float32) *Camera {
	// Position camera to see the entire map
	mapCenter := mapSize / 2.0
	cameraHeight := mapSize * 0.8 // Height proportional to map size
	cameraDistance := mapSize * 0.6

	camera := &Camera{
		Position:    mgl32.Vec3{mapCenter, cameraHeight, mapCenter + cameraDistance},
		Target:      mgl32.Vec3{mapCenter, 0, mapCenter}, // Look at map center
		Up:          mgl32.Vec3{0, 1, 0},
		FOV:         mgl32.DegToRad(60.0), // Wider FOV for RTS
		AspectRatio: float32(windowWidth) / float32(windowHeight),
		NearPlane:   1.0,
		FarPlane:    mapSize * 3.0, // Scale far plane with map size
		isDirty:     true,
	}

	camera.updateMatrices()
	return camera
}

// updateMatrices recalculates view and projection matrices if needed
func (c *Camera) updateMatrices() {
	if !c.isDirty {
		return
	}

	// Calculate view matrix (lookAt)
	c.ViewMatrix = mgl32.LookAtV(c.Position, c.Target, c.Up)

	// Calculate projection matrix (perspective)
	c.ProjectionMatrix = mgl32.Perspective(c.FOV, c.AspectRatio, c.NearPlane, c.FarPlane)

	c.isDirty = false
}

// GetViewMatrix returns the camera's view matrix
func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	c.updateMatrices()
	return c.ViewMatrix
}

// GetProjectionMatrix returns the camera's projection matrix
func (c *Camera) GetProjectionMatrix() mgl32.Mat4 {
	c.updateMatrices()
	return c.ProjectionMatrix
}

// SetPosition updates the camera position
func (c *Camera) SetPosition(x, y, z float32) {
	c.Position = mgl32.Vec3{x, y, z}
	c.isDirty = true
}

// SetTarget updates the camera target (look-at point)
func (c *Camera) SetTarget(x, y, z float32) {
	c.Target = mgl32.Vec3{x, y, z}
	c.isDirty = true
}

// LookAt sets both position and target
func (c *Camera) LookAt(eyeX, eyeY, eyeZ, targetX, targetY, targetZ float32) {
	c.Position = mgl32.Vec3{eyeX, eyeY, eyeZ}
	c.Target = mgl32.Vec3{targetX, targetY, targetZ}
	c.isDirty = true
}

// SetAspectRatio updates the aspect ratio (call when window is resized)
func (c *Camera) SetAspectRatio(windowWidth, windowHeight int) {
	c.AspectRatio = float32(windowWidth) / float32(windowHeight)
	c.isDirty = true
}

// SetFOV updates the field of view (in degrees)
func (c *Camera) SetFOV(fovDegrees float32) {
	c.FOV = mgl32.DegToRad(fovDegrees)
	c.isDirty = true
}

// SetClippingPlanes updates near and far clipping planes
func (c *Camera) SetClippingPlanes(nearPlane, farPlane float32) {
	c.NearPlane = nearPlane
	c.FarPlane = farPlane
	c.isDirty = true
}

// Move translates the camera by the given offset
func (c *Camera) Move(deltaX, deltaY, deltaZ float32) {
	offset := mgl32.Vec3{deltaX, deltaY, deltaZ}
	c.Position = c.Position.Add(offset)
	c.Target = c.Target.Add(offset)
	c.isDirty = true
}

// Orbit rotates the camera around the target point
func (c *Camera) Orbit(angleX, angleY float32) {
	// Calculate direction from target to camera
	direction := c.Position.Sub(c.Target)
	distance := direction.Len()

	// Convert to spherical coordinates
	theta := float32(math.Atan2(float64(direction.Z()), float64(direction.X())))
	phi := float32(math.Acos(float64(direction.Y()) / float64(distance)))

	// Apply rotation
	theta += angleY
	phi += angleX

	// Clamp phi to prevent flipping
	const epsilon = 0.01
	phi = mgl32.Clamp(phi, epsilon, math.Pi-epsilon)

	// Convert back to cartesian coordinates
	newDirection := mgl32.Vec3{
		float32(math.Sin(float64(phi)) * math.Cos(float64(theta))),
		float32(math.Cos(float64(phi))),
		float32(math.Sin(float64(phi)) * math.Sin(float64(theta))),
	}.Mul(distance)

	c.Position = c.Target.Add(newDirection)
	c.isDirty = true
}

// Zoom moves the camera closer to or further from the target
func (c *Camera) Zoom(delta float32) {
	direction := c.Position.Sub(c.Target)
	distance := direction.Len()

	// Calculate new distance
	newDistance := distance - delta
	minDistance := float32(1.0)
	maxDistance := c.FarPlane * 0.8

	newDistance = mgl32.Clamp(newDistance, minDistance, maxDistance)

	// Update position
	if newDistance != distance {
		direction = direction.Normalize().Mul(newDistance)
		c.Position = c.Target.Add(direction)
		c.isDirty = true
	}
}

// GetForwardVector returns the camera's forward direction
func (c *Camera) GetForwardVector() mgl32.Vec3 {
	return c.Target.Sub(c.Position).Normalize()
}

// GetRightVector returns the camera's right direction
func (c *Camera) GetRightVector() mgl32.Vec3 {
	forward := c.GetForwardVector()
	return forward.Cross(c.Up).Normalize()
}

// GetUpVector returns the camera's up direction
func (c *Camera) GetUpVector() mgl32.Vec3 {
	return c.Up
}

// ScreenToWorldRay converts screen coordinates to a world space ray
func (c *Camera) ScreenToWorldRay(screenX, screenY, screenWidth, screenHeight int) (origin, direction mgl32.Vec3) {
	c.updateMatrices()

	// Convert screen coordinates to normalized device coordinates
	x := (2.0*float32(screenX))/float32(screenWidth) - 1.0
	y := 1.0 - (2.0*float32(screenY))/float32(screenHeight)

	// Create points in normalized device coordinates
	nearPoint := mgl32.Vec4{x, y, -1.0, 1.0}  // Near plane
	farPoint := mgl32.Vec4{x, y, 1.0, 1.0}    // Far plane

	// Transform to world space
	inverseVP := c.ProjectionMatrix.Mul4(c.ViewMatrix).Inv()

	worldNear := inverseVP.Mul4x1(nearPoint)
	worldFar := inverseVP.Mul4x1(farPoint)

	// Perspective divide
	worldNear = worldNear.Mul(1.0 / worldNear.W())
	worldFar = worldFar.Mul(1.0 / worldFar.W())

	// Calculate ray
	origin = worldNear.Vec3()
	direction = worldFar.Vec3().Sub(worldNear.Vec3()).Normalize()

	return origin, direction
}

// FrustumCulling checks if a bounding box is visible in the camera's frustum
func (c *Camera) IsInFrustum(min, max mgl32.Vec3) bool {
	// Simple frustum culling using view-projection matrix
	c.updateMatrices()
	viewProj := c.ProjectionMatrix.Mul4(c.ViewMatrix)

	// Test all 8 corners of the bounding box
	corners := [8]mgl32.Vec3{
		{min.X(), min.Y(), min.Z()},
		{max.X(), min.Y(), min.Z()},
		{min.X(), max.Y(), min.Z()},
		{max.X(), max.Y(), min.Z()},
		{min.X(), min.Y(), max.Z()},
		{max.X(), min.Y(), max.Z()},
		{min.X(), max.Y(), max.Z()},
		{max.X(), max.Y(), max.Z()},
	}

	for _, corner := range corners {
		clipPos := viewProj.Mul4x1(mgl32.Vec4{corner.X(), corner.Y(), corner.Z(), 1.0})

		// Check if point is inside frustum
		w := clipPos.W()
		if clipPos.X() >= -w && clipPos.X() <= w &&
		   clipPos.Y() >= -w && clipPos.Y() <= w &&
		   clipPos.Z() >= -w && clipPos.Z() <= w {
			return true // At least one corner is visible
		}
	}

	return false // All corners are outside frustum
}

// GetViewFrustum returns the camera's view frustum for culling
type ViewFrustum struct {
	Near, Far             float32
	Left, Right           float32
	Top, Bottom           float32
	Position              mgl32.Vec3
	Forward, RightVec, Up mgl32.Vec3
}

func (c *Camera) GetViewFrustum() ViewFrustum {
	c.updateMatrices()

	// Calculate frustum dimensions at near plane
	halfHeight := c.NearPlane * float32(math.Tan(float64(c.FOV)/2.0))
	halfWidth := halfHeight * c.AspectRatio

	return ViewFrustum{
		Near:     c.NearPlane,
		Far:      c.FarPlane,
		Left:     -halfWidth,
		Right:    halfWidth,
		Top:      halfHeight,
		Bottom:   -halfHeight,
		Position: c.Position,
		Forward:  c.GetForwardVector(),
		RightVec: c.GetRightVector(),
		Up:       c.GetUpVector(),
	}
}