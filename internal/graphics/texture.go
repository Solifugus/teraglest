package graphics

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg" // JPEG support
	_ "image/png"  // PNG support
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// TextureManager handles loading and management of textures
type TextureManager struct {
	textureCache map[string]*Texture
}

// Texture represents an OpenGL texture
type Texture struct {
	ID     uint32 // OpenGL texture ID
	Width  int32  // Texture width
	Height int32  // Texture height
	Format uint32 // OpenGL format (GL_RGB, GL_RGBA)
	Path   string // Original file path
}

// NewTextureManager creates a new texture manager
func NewTextureManager() *TextureManager {
	return &TextureManager{
		textureCache: make(map[string]*Texture),
	}
}

// LoadTexture loads a texture from file path and uploads it to GPU
func (tm *TextureManager) LoadTexture(filePath string) (*Texture, error) {
	// Check cache first
	if texture, exists := tm.textureCache[filePath]; exists {
		return texture, nil
	}

	// Load image from file
	imageData, err := loadImageFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", filePath, err)
	}

	// Create OpenGL texture
	texture, err := createGLTexture(imageData, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenGL texture: %w", err)
	}

	// Cache the texture
	tm.textureCache[filePath] = texture

	return texture, nil
}

// LoadTextureFromImage loads a texture from an already loaded image
func (tm *TextureManager) LoadTextureFromImage(img image.Image, path string) (*Texture, error) {
	// Check cache first
	if texture, exists := tm.textureCache[path]; exists {
		return texture, nil
	}

	// Create OpenGL texture
	texture, err := createGLTexture(img, path)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenGL texture: %w", err)
	}

	// Cache the texture
	tm.textureCache[path] = texture

	return texture, nil
}

// GetTexture returns a cached texture or nil if not found
func (tm *TextureManager) GetTexture(filePath string) *Texture {
	return tm.textureCache[filePath]
}

// UnloadTexture removes a texture from cache and GPU
func (tm *TextureManager) UnloadTexture(filePath string) {
	if texture, exists := tm.textureCache[filePath]; exists {
		texture.Cleanup()
		delete(tm.textureCache, filePath)
	}
}

// GetCacheSize returns the number of cached textures
func (tm *TextureManager) GetCacheSize() int {
	return len(tm.textureCache)
}

// Cleanup releases all cached textures
func (tm *TextureManager) Cleanup() {
	for _, texture := range tm.textureCache {
		texture.Cleanup()
	}
	tm.textureCache = make(map[string]*Texture)
}

// loadImageFromFile loads an image from a file path
func loadImageFromFile(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Decode image (format is auto-detected based on file content)
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Log successful load for debugging
	fmt.Printf("Loaded %s image: %dx%d from %s\n", format, img.Bounds().Dx(), img.Bounds().Dy(), filePath)

	return img, nil
}

// createGLTexture creates an OpenGL texture from an image
func createGLTexture(img image.Image, path string) (*Texture, error) {
	// Convert image to RGBA format for OpenGL
	rgba := convertToRGBA(img)

	// Generate OpenGL texture
	var textureID uint32
	gl.GenTextures(1, &textureID)
	gl.BindTexture(gl.TEXTURE_2D, textureID)

	// Determine format
	var format uint32 = gl.RGBA
	if isOpaqueImage(img) {
		format = gl.RGB
	}

	// Upload texture data
	width := int32(rgba.Bounds().Dx())
	height := int32(rgba.Bounds().Dy())

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,                    // mipmap level
		int32(format),        // internal format
		width,                // width
		height,               // height
		0,                    // border
		format,               // format
		gl.UNSIGNED_BYTE,     // type
		gl.Ptr(rgba.Pix),     // data
	)

	// Generate mipmaps for better quality at distance
	gl.GenerateMipmap(gl.TEXTURE_2D)

	// Set texture parameters
	setTextureParameters()

	// Unbind texture
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// Check for OpenGL errors
	if err := gl.GetError(); err != 0 {
		gl.DeleteTextures(1, &textureID)
		return nil, fmt.Errorf("OpenGL error creating texture: %d", err)
	}

	texture := &Texture{
		ID:     textureID,
		Width:  width,
		Height: height,
		Format: format,
		Path:   path,
	}

	return texture, nil
}

// convertToRGBA converts any image to RGBA format
func convertToRGBA(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
	return rgba
}

// isOpaqueImage checks if an image has transparency
func isOpaqueImage(img image.Image) bool {
	// Check if the image model includes alpha
	switch img.ColorModel() {
	case color.RGBAModel, color.RGBA64Model, color.NRGBAModel, color.NRGBA64Model:
		return false // Has alpha channel
	default:
		return true  // No alpha channel
	}
}

// setTextureParameters sets standard texture parameters for 3D models
func setTextureParameters() {
	// Texture wrapping
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	// Texture filtering
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR) // Trilinear filtering
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)               // Linear magnification

	// Note: Anisotropic filtering extensions would go here if available
	// For now, we use standard filtering which works well
}

// CreateDefaultTexture creates a simple white texture for models without textures
func (tm *TextureManager) CreateDefaultTexture() (*Texture, error) {
	const defaultPath = "__default_white__"

	// Check if default texture already exists
	if texture := tm.GetTexture(defaultPath); texture != nil {
		return texture, nil
	}

	// Create a 2x2 white image
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	white := color.RGBA{255, 255, 255, 255}
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.Set(x, y, white)
		}
	}

	return tm.LoadTextureFromImage(img, defaultPath)
}

// CreateCheckerTexture creates a checkerboard texture for testing
func (tm *TextureManager) CreateCheckerTexture(size int) (*Texture, error) {
	checkerPath := fmt.Sprintf("__checker_%d__", size)

	// Check if checker texture already exists
	if texture := tm.GetTexture(checkerPath); texture != nil {
		return texture, nil
	}

	// Create checkerboard pattern
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	white := color.RGBA{255, 255, 255, 255}
	black := color.RGBA{0, 0, 0, 255}

	checkerSize := size / 8 // 8x8 checkerboard
	if checkerSize < 1 {
		checkerSize = 1
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			checkerX := x / checkerSize
			checkerY := y / checkerSize

			var color color.RGBA
			if (checkerX+checkerY)%2 == 0 {
				color = white
			} else {
				color = black
			}

			img.Set(x, y, color)
		}
	}

	return tm.LoadTextureFromImage(img, checkerPath)
}

// FindTextureForModel attempts to find an appropriate texture for a G3D model
func (tm *TextureManager) FindTextureForModel(modelPath string, textureNames []string) (*Texture, error) {
	// First, try loading textures specified in the G3D file
	for _, textureName := range textureNames {
		if textureName == "" {
			continue
		}

		// Try different common extensions
		extensions := []string{".png", ".jpg", ".jpeg", ".tga", ".bmp"}

		for _, ext := range extensions {
			// Remove existing extension from texture name
			baseName := strings.TrimSuffix(textureName, ".tga")
			baseName = strings.TrimSuffix(baseName, ".png")
			baseName = strings.TrimSuffix(baseName, ".jpg")
			baseName = strings.TrimSuffix(baseName, ".jpeg")
			baseName = strings.TrimSuffix(baseName, ".bmp")

			texturePath := baseName + ext

			// Try loading relative to model path
			modelDir := filepath.Dir(modelPath)
			fullPath := filepath.Join(modelDir, texturePath)

			if _, err := os.Stat(fullPath); err == nil {
				texture, err := tm.LoadTexture(fullPath)
				if err == nil {
					return texture, nil
				}
			}
		}
	}

	// If no texture found, create default white texture
	return tm.CreateDefaultTexture()
}

// Bind binds the texture to the specified texture unit
func (t *Texture) Bind(textureUnit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + textureUnit)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}

// Unbind unbinds any texture from the specified texture unit
func UnbindTexture(textureUnit uint32) {
	gl.ActiveTexture(gl.TEXTURE0 + textureUnit)
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Cleanup releases the OpenGL texture
func (t *Texture) Cleanup() {
	if t.ID != 0 {
		gl.DeleteTextures(1, &t.ID)
		t.ID = 0
	}
}