package engine

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Tileset represents a complete terrain tileset with all visual and gameplay data
type Tileset struct {
	Name          string              `json:"name"`
	Surfaces      []SurfaceType       `json:"surfaces"`       // Terrain surface definitions
	Objects       []TerrainObject     `json:"objects"`        // Terrain object definitions
	AmbientSounds *AmbientSounds      `json:"ambient_sounds"` // Environmental audio
	Parameters    TerrainParameters   `json:"parameters"`     // Lighting, water, fog settings
	BasePath      string              `json:"base_path"`      // Base directory path for assets
}

// SurfaceType defines a terrain surface with multiple texture variations
type SurfaceType struct {
	Index         int               `json:"index"`
	Textures      []SurfaceTexture  `json:"textures"`  // Multiple textures with probabilities
	TotalProbability float32        `json:"total_probability"` // Sum of all texture probabilities
}

// SurfaceTexture represents a single texture variation for a surface
type SurfaceTexture struct {
	Path        string  `json:"path"`        // Relative path to texture file
	Probability float32 `json:"probability"` // Probability of this texture being used (0.0-1.0)
}

// TerrainObject represents a placeable terrain object (trees, rocks, etc.)
type TerrainObject struct {
	Index    int      `json:"index"`    // Object type index
	Models   []string `json:"models"`   // G3D model paths
	Walkable bool     `json:"walkable"` // Whether units can walk through this object
}

// AmbientSounds contains environmental audio configuration
type AmbientSounds struct {
	DaySound       AudioConfig `json:"day_sound"`
	NightSound     AudioConfig `json:"night_sound"`
	RainSound      AudioConfig `json:"rain_sound"`
	SnowSound      AudioConfig `json:"snow_sound"`
	DayStartSound  AudioConfig `json:"day_start_sound"`
	NightStartSound AudioConfig `json:"night_start_sound"`
}

// AudioConfig represents audio file configuration
type AudioConfig struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
	Volume  float32 `json:"volume"` // 0.0 to 1.0
}

// TerrainParameters contains visual and environmental parameters
type TerrainParameters struct {
	Water     WaterConfig   `json:"water"`
	Fog       FogConfig     `json:"fog"`
	Lighting  LightConfig   `json:"lighting"`
	Weather   WeatherConfig `json:"weather"`
}

// WaterConfig contains water rendering and animation settings
type WaterConfig struct {
	Effects       bool     `json:"effects"`        // Enable water effects
	Textures      []string `json:"textures"`       // Water animation frame textures
	FrameCount    int      `json:"frame_count"`    // Number of animation frames
	Speed         float32  `json:"speed"`          // Animation speed
}

// FogConfig contains fog rendering settings
type FogConfig struct {
	Enabled   bool    `json:"enabled"`
	Mode      string  `json:"mode"`      // Fog mode (linear, exponential)
	Density   float32 `json:"density"`   // Fog density (0.0-1.0)
	Red       float32 `json:"red"`       // Fog color red component (0.0-1.0)
	Green     float32 `json:"green"`     // Fog color green component (0.0-1.0)
	Blue      float32 `json:"blue"`      // Fog color blue component (0.0-1.0)
}

// LightConfig contains lighting settings
type LightConfig struct {
	// Sun light settings
	SunRed     float32 `json:"sun_red"`
	SunGreen   float32 `json:"sun_green"`
	SunBlue    float32 `json:"sun_blue"`

	// Moon light settings
	MoonRed    float32 `json:"moon_red"`
	MoonGreen  float32 `json:"moon_green"`
	MoonBlue   float32 `json:"moon_blue"`

	// Day/Night cycle
	DayTime    float32 `json:"day_time"`    // Duration of day phase
	NightTime  float32 `json:"night_time"`  // Duration of night phase
}

// WeatherConfig contains weather probability settings
type WeatherConfig struct {
	SunProbability   float32 `json:"sun_probability"`   // Probability of sunny weather (0.0-1.0)
	RainProbability  float32 `json:"rain_probability"`  // Probability of rain (0.0-1.0)
	SnowProbability  float32 `json:"snow_probability"`  // Probability of snow (0.0-1.0)
}

// TilesetXML represents the XML structure for parsing tileset files
type TilesetXML struct {
	XMLName       xml.Name           `xml:"tileset"`
	Surfaces      SurfacesXML        `xml:"surfaces"`
	Objects       ObjectsXML         `xml:"objects"`
	AmbientSounds AmbientSoundsXML   `xml:"ambient-sounds"`
	Parameters    ParametersXML      `xml:"parameters"`
}

// SurfacesXML contains surface definitions
type SurfacesXML struct {
	Surface []SurfaceXML `xml:"surface"`
}

// SurfaceXML represents a single surface definition
type SurfaceXML struct {
	Textures []TextureXML `xml:"texture"`
}

// TextureXML represents a texture variation
type TextureXML struct {
	Path string `xml:"path,attr"`
	Prob string `xml:"prob,attr"`
}

// ObjectsXML contains terrain object definitions
type ObjectsXML struct {
	Object []ObjectXML `xml:"object"`
}

// ObjectXML represents a single terrain object
type ObjectXML struct {
	Walkable string     `xml:"walkable,attr"`
	Models   []ModelXML `xml:"model"`
}

// ModelXML represents a G3D model reference
type ModelXML struct {
	Path string `xml:"path,attr"`
}

// AmbientSoundsXML contains ambient sound configuration
type AmbientSoundsXML struct {
	DaySound       SoundXML `xml:"day-sound"`
	NightSound     SoundXML `xml:"night-sound"`
	RainSound      SoundXML `xml:"rain-sound"`
	SnowSound      SoundXML `xml:"snow-sound"`
	DayStartSound  SoundXML `xml:"day-start-sound"`
	NightStartSound SoundXML `xml:"night-start-sound"`
}

// SoundXML represents sound configuration
type SoundXML struct {
	Enabled string `xml:"enabled,attr"`
	Path    string `xml:"path,attr"`
	Volume  string `xml:"volume,attr"`
}

// ParametersXML contains environmental parameters
type ParametersXML struct {
	Water    WaterXML    `xml:"water"`
	Fog      FogXML      `xml:"fog"`
	Sun      LightXML    `xml:"sun"`
	Moon     LightXML    `xml:"moon"`
	Weather  WeatherXML  `xml:"weather"`
	DayTime  TimeXML     `xml:"day-time"`
	NightTime TimeXML    `xml:"night-time"`
}

// WaterXML contains water configuration
type WaterXML struct {
	Effects  string       `xml:"effects,attr"`
	Textures []TextureRefXML `xml:"texture"`
}

// TextureRefXML represents a texture reference
type TextureRefXML struct {
	Path string `xml:"path,attr"`
}

// FogXML contains fog configuration
type FogXML struct {
	Enabled string `xml:"enabled,attr"`
	Mode    string `xml:"mode,attr"`
	Density string `xml:"density,attr"`
	Red     string `xml:"red,attr"`
	Green   string `xml:"green,attr"`
	Blue    string `xml:"blue,attr"`
}

// LightXML contains lighting configuration
type LightXML struct {
	Red   string `xml:"red,attr"`
	Green string `xml:"green,attr"`
	Blue  string `xml:"blue,attr"`
}

// WeatherXML contains weather configuration
type WeatherXML struct {
	Sun  string `xml:"sun,attr"`
	Rain string `xml:"rain,attr"`
	Snow string `xml:"snow,attr"`
}

// TimeXML contains time configuration
type TimeXML struct {
	Value string `xml:"value,attr"`
}

// TilesetLoader handles loading and parsing of tileset XML files
type TilesetLoader struct {
	basePath string
}

// NewTilesetLoader creates a new tileset loader
func NewTilesetLoader(basePath string) *TilesetLoader {
	return &TilesetLoader{
		basePath: basePath,
	}
}

// LoadTileset loads a tileset from the specified name
func (tl *TilesetLoader) LoadTileset(tilesetName string) (*Tileset, error) {
	// Construct path to tileset XML file
	xmlPath := filepath.Join(tl.basePath, "tilesets", tilesetName, tilesetName+".xml")

	// Check if file exists
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tileset file not found: %s", xmlPath)
	}

	// Read and parse XML file
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tileset file %s: %w", xmlPath, err)
	}

	// Parse XML
	var tilesetXML TilesetXML
	if err := xml.Unmarshal(xmlData, &tilesetXML); err != nil {
		return nil, fmt.Errorf("failed to parse tileset XML %s: %w", xmlPath, err)
	}

	// Convert XML structure to internal format
	tileset, err := tl.convertXMLToTileset(tilesetName, tilesetXML)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tileset data: %w", err)
	}

	// Set base path for asset resolution
	tileset.BasePath = filepath.Join(tl.basePath, "tilesets", tilesetName)

	return tileset, nil
}

// convertXMLToTileset converts parsed XML data to internal Tileset structure
func (tl *TilesetLoader) convertXMLToTileset(name string, xmlData TilesetXML) (*Tileset, error) {
	tileset := &Tileset{
		Name: name,
	}

	// Convert surfaces
	tileset.Surfaces = make([]SurfaceType, len(xmlData.Surfaces.Surface))
	for i, surfaceXML := range xmlData.Surfaces.Surface {
		surface := SurfaceType{
			Index:    i + 1, // Surface indices start at 1
			Textures: make([]SurfaceTexture, len(surfaceXML.Textures)),
		}

		var totalProb float32
		for j, texXML := range surfaceXML.Textures {
			prob, err := tl.parseFloat32(texXML.Prob, 1.0)
			if err != nil {
				return nil, fmt.Errorf("invalid texture probability: %s", texXML.Prob)
			}

			surface.Textures[j] = SurfaceTexture{
				Path:        texXML.Path,
				Probability: prob,
			}
			totalProb += prob
		}

		surface.TotalProbability = totalProb
		tileset.Surfaces[i] = surface
	}

	// Convert objects
	tileset.Objects = make([]TerrainObject, len(xmlData.Objects.Object))
	for i, objXML := range xmlData.Objects.Object {
		walkable, err := tl.parseBool(objXML.Walkable, false)
		if err != nil {
			return nil, fmt.Errorf("invalid walkable value: %s", objXML.Walkable)
		}

		object := TerrainObject{
			Index:    i + 1, // Object indices start at 1
			Walkable: walkable,
			Models:   make([]string, len(objXML.Models)),
		}

		for j, modelXML := range objXML.Models {
			object.Models[j] = modelXML.Path
		}

		tileset.Objects[i] = object
	}

	// Convert ambient sounds
	ambientSounds, err := tl.convertAmbientSounds(xmlData.AmbientSounds)
	if err != nil {
		return nil, fmt.Errorf("failed to convert ambient sounds: %w", err)
	}
	tileset.AmbientSounds = ambientSounds

	// Convert parameters
	parameters, err := tl.convertParameters(xmlData.Parameters)
	if err != nil {
		return nil, fmt.Errorf("failed to convert parameters: %w", err)
	}
	tileset.Parameters = *parameters

	return tileset, nil
}

// convertAmbientSounds converts ambient sounds XML to internal format
func (tl *TilesetLoader) convertAmbientSounds(xmlSounds AmbientSoundsXML) (*AmbientSounds, error) {
	sounds := &AmbientSounds{}

	var err error
	if sounds.DaySound, err = tl.convertAudioConfig(xmlSounds.DaySound); err != nil {
		return nil, fmt.Errorf("failed to convert day sound: %w", err)
	}
	if sounds.NightSound, err = tl.convertAudioConfig(xmlSounds.NightSound); err != nil {
		return nil, fmt.Errorf("failed to convert night sound: %w", err)
	}
	if sounds.RainSound, err = tl.convertAudioConfig(xmlSounds.RainSound); err != nil {
		return nil, fmt.Errorf("failed to convert rain sound: %w", err)
	}
	if sounds.SnowSound, err = tl.convertAudioConfig(xmlSounds.SnowSound); err != nil {
		return nil, fmt.Errorf("failed to convert snow sound: %w", err)
	}
	if sounds.DayStartSound, err = tl.convertAudioConfig(xmlSounds.DayStartSound); err != nil {
		return nil, fmt.Errorf("failed to convert day start sound: %w", err)
	}
	if sounds.NightStartSound, err = tl.convertAudioConfig(xmlSounds.NightStartSound); err != nil {
		return nil, fmt.Errorf("failed to convert night start sound: %w", err)
	}

	return sounds, nil
}

// convertAudioConfig converts audio configuration XML to internal format
func (tl *TilesetLoader) convertAudioConfig(xmlAudio SoundXML) (AudioConfig, error) {
	enabled, err := tl.parseBool(xmlAudio.Enabled, false)
	if err != nil {
		return AudioConfig{}, fmt.Errorf("invalid enabled value: %s", xmlAudio.Enabled)
	}

	volume, err := tl.parseFloat32(xmlAudio.Volume, 1.0)
	if err != nil {
		return AudioConfig{}, fmt.Errorf("invalid volume value: %s", xmlAudio.Volume)
	}

	return AudioConfig{
		Enabled: enabled,
		Path:    xmlAudio.Path,
		Volume:  volume,
	}, nil
}

// convertParameters converts environmental parameters XML to internal format
func (tl *TilesetLoader) convertParameters(xmlParams ParametersXML) (*TerrainParameters, error) {
	params := &TerrainParameters{}

	// Convert water configuration
	waterEffects, err := tl.parseBool(xmlParams.Water.Effects, false)
	if err != nil {
		return nil, fmt.Errorf("invalid water effects value: %s", xmlParams.Water.Effects)
	}

	params.Water = WaterConfig{
		Effects:    waterEffects,
		Textures:   make([]string, len(xmlParams.Water.Textures)),
		FrameCount: len(xmlParams.Water.Textures),
		Speed:      1.0, // Default animation speed
	}

	for i, texRef := range xmlParams.Water.Textures {
		params.Water.Textures[i] = texRef.Path
	}

	// Convert fog configuration
	fogEnabled, _ := tl.parseBool(xmlParams.Fog.Enabled, false)
	fogDensity, _ := tl.parseFloat32(xmlParams.Fog.Density, 0.0)
	fogRed, _ := tl.parseFloat32(xmlParams.Fog.Red, 1.0)
	fogGreen, _ := tl.parseFloat32(xmlParams.Fog.Green, 1.0)
	fogBlue, _ := tl.parseFloat32(xmlParams.Fog.Blue, 1.0)

	params.Fog = FogConfig{
		Enabled: fogEnabled,
		Mode:    strings.TrimSpace(xmlParams.Fog.Mode),
		Density: fogDensity,
		Red:     fogRed,
		Green:   fogGreen,
		Blue:    fogBlue,
	}

	// Convert lighting configuration
	sunRed, _ := tl.parseFloat32(xmlParams.Sun.Red, 1.0)
	sunGreen, _ := tl.parseFloat32(xmlParams.Sun.Green, 1.0)
	sunBlue, _ := tl.parseFloat32(xmlParams.Sun.Blue, 1.0)
	moonRed, _ := tl.parseFloat32(xmlParams.Moon.Red, 0.5)
	moonGreen, _ := tl.parseFloat32(xmlParams.Moon.Green, 0.5)
	moonBlue, _ := tl.parseFloat32(xmlParams.Moon.Blue, 0.7)
	dayTime, _ := tl.parseFloat32(xmlParams.DayTime.Value, 1000.0)
	nightTime, _ := tl.parseFloat32(xmlParams.NightTime.Value, 1000.0)

	params.Lighting = LightConfig{
		SunRed:    sunRed,
		SunGreen:  sunGreen,
		SunBlue:   sunBlue,
		MoonRed:   moonRed,
		MoonGreen: moonGreen,
		MoonBlue:  moonBlue,
		DayTime:   dayTime,
		NightTime: nightTime,
	}

	// Convert weather configuration
	sunProb, _ := tl.parseFloat32(xmlParams.Weather.Sun, 0.6)
	rainProb, _ := tl.parseFloat32(xmlParams.Weather.Rain, 0.2)
	snowProb, _ := tl.parseFloat32(xmlParams.Weather.Snow, 0.2)

	params.Weather = WeatherConfig{
		SunProbability:  sunProb,
		RainProbability: rainProb,
		SnowProbability: snowProb,
	}

	return params, nil
}

// Helper functions for parsing XML attributes with defaults

// parseBool parses a boolean string with a default value
func (tl *TilesetLoader) parseBool(s string, defaultVal bool) (bool, error) {
	if strings.TrimSpace(s) == "" {
		return defaultVal, nil
	}
	return strconv.ParseBool(strings.TrimSpace(s))
}

// parseFloat32 parses a float32 string with a default value
func (tl *TilesetLoader) parseFloat32(s string, defaultVal float32) (float32, error) {
	if strings.TrimSpace(s) == "" {
		return defaultVal, nil
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return float32(val), err
}

// Utility methods for the Tileset structure

// GetSurface returns the surface type at the specified index (1-based)
func (t *Tileset) GetSurface(index int) *SurfaceType {
	if index >= 1 && index <= len(t.Surfaces) {
		return &t.Surfaces[index-1]
	}
	return nil
}

// GetObject returns the terrain object at the specified index (1-based)
func (t *Tileset) GetObject(index int) *TerrainObject {
	if index >= 1 && index <= len(t.Objects) {
		return &t.Objects[index-1]
	}
	return nil
}

// IsObjectWalkable returns whether the object at the given index is walkable
func (t *Tileset) IsObjectWalkable(index int) bool {
	if obj := t.GetObject(index); obj != nil {
		return obj.Walkable
	}
	return true // No object = walkable
}

// GetRandomTexture returns a random texture path for the given surface based on probabilities
func (st *SurfaceType) GetRandomTexture(randomValue float32) string {
	if len(st.Textures) == 0 {
		return ""
	}

	// Normalize random value to total probability
	target := randomValue * st.TotalProbability
	accumulated := float32(0.0)

	for _, texture := range st.Textures {
		accumulated += texture.Probability
		if accumulated >= target {
			return texture.Path
		}
	}

	// Fallback to first texture
	return st.Textures[0].Path
}

// PrintSummary prints a summary of the tileset information
func (t *Tileset) PrintSummary() {
	fmt.Printf("Tileset: %s\n", t.Name)
	fmt.Printf("  Base Path: %s\n", t.BasePath)
	fmt.Printf("  Surfaces: %d types\n", len(t.Surfaces))
	fmt.Printf("  Objects: %d types\n", len(t.Objects))

	fmt.Printf("  Water: %t", t.Parameters.Water.Effects)
	if t.Parameters.Water.Effects {
		fmt.Printf(" (%d frames)\n", t.Parameters.Water.FrameCount)
	} else {
		fmt.Printf("\n")
	}

	fmt.Printf("  Fog: %t", t.Parameters.Fog.Enabled)
	if t.Parameters.Fog.Enabled {
		fmt.Printf(" (density: %.3f)\n", t.Parameters.Fog.Density)
	} else {
		fmt.Printf("\n")
	}

	fmt.Printf("  Weather: Sun %.1f%%, Rain %.1f%%, Snow %.1f%%\n",
		t.Parameters.Weather.SunProbability*100,
		t.Parameters.Weather.RainProbability*100,
		t.Parameters.Weather.SnowProbability*100)

	if t.AmbientSounds != nil {
		fmt.Printf("  Audio: Day=%t, Night=%t, Rain=%t, Snow=%t\n",
			t.AmbientSounds.DaySound.Enabled,
			t.AmbientSounds.NightSound.Enabled,
			t.AmbientSounds.RainSound.Enabled,
			t.AmbientSounds.SnowSound.Enabled)
	}
}