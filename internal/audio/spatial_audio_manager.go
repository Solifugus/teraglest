package audio

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// SpatialAudioManager handles 3D positional audio for the game world
type SpatialAudioManager struct {
	backend  AudioBackend
	settings *AudioSettings

	// Listener (camera/player) properties
	listenerPosition    Vector3
	listenerOrientation ListenerOrientation
	listenerVelocity    Vector3

	// Active 3D sounds
	spatialSounds map[string]*SpatialSoundInstance
	ambientSounds map[string]*AmbientSoundInstance

	// Audio zones and environments
	audioZones     map[string]*AudioZone
	currentZone    *AudioZone
	environmentFX  *EnvironmentEffects

	// Distance models and attenuation
	globalMaxDistance   float32
	globalRolloffFactor float32
	dopplerFactor       float32

	// Performance optimization
	maxSpatialSounds    int
	soundUpdateBudget   int // Max sounds to update per frame
	lastUpdateIndex     int

	// Ambient system
	ambientLayers    map[string]*AmbientLayer
	weatherIntensity float32
	timeOfDay        float32 // 0.0 = midnight, 0.5 = noon

	mutex sync.RWMutex
}

// ListenerOrientation defines the listener's orientation in 3D space
type ListenerOrientation struct {
	Forward Vector3 // Direction the listener is facing
	Up      Vector3 // Up direction for the listener
}

// SpatialSoundInstance represents a 3D positioned sound
type SpatialSoundInstance struct {
	ID       string
	Sound    *Sound
	Position Vector3
	Velocity Vector3

	// 3D Audio properties
	MinDistance      float32
	MaxDistance      float32
	AttenuationModel string
	Volume           float32
	Pitch            float32

	// State
	IsActive    bool
	IsLooping   bool
	StartTime   time.Time
	LastUpdate  time.Time

	// Calculated properties (updated per frame)
	DistanceToListener float32
	Direction          Vector3 // Direction from listener to sound
	EffectiveVolume    float32
	EffectivePitch     float32

	// Backend data
	BackendInstance interface{}
}

// AmbientSoundInstance represents ambient environmental sounds
type AmbientSoundInstance struct {
	ID          string
	Sound       *Sound
	Position    *Vector3 // nil for global ambient sounds
	Radius      float32  // Area of effect
	Volume      float32
	IsActive    bool
	LayerName   string
}

// AudioZone defines an area with specific audio properties
type AudioZone struct {
	Name        string
	Bounds      BoundingBox
	Reverb      ReverbSettings
	Occlusion   OcclusionSettings
	AmbientSets []string // Ambient sound sets for this zone

	// Environmental properties
	Material    string  // affects reverb and occlusion
	Humidity    float32 // affects sound propagation
	Temperature float32 // affects sound speed
	WindFactor  float32 // affects ambient sounds
}

// BoundingBox defines a 3D rectangular area
type BoundingBox struct {
	Min Vector3
	Max Vector3
}

// ReverbSettings control reverb effects in audio zones
type ReverbSettings struct {
	Enabled     bool
	RoomSize    float32 // 0.0 - 1.0
	Damping     float32 // 0.0 - 1.0
	WetLevel    float32 // 0.0 - 1.0
	DryLevel    float32 // 0.0 - 1.0
	Delay       float32 // milliseconds
}

// OcclusionSettings control how objects block sound
type OcclusionSettings struct {
	Enabled         bool
	LowPassFilter   float32 // Frequency cutoff for occluded sounds
	VolumeReduction float32 // Volume reduction for occluded sounds
}

// EnvironmentEffects manages environmental audio effects
type EnvironmentEffects struct {
	ReverbEnabled    bool
	OcclusionEnabled bool
	DopplerEnabled   bool
	AirAbsorption    bool
}

// AmbientLayer represents a layer of ambient sounds
type AmbientLayer struct {
	Name      string
	Sounds    []*AmbientSoundInstance
	Volume    float32
	IsActive  bool
	Priority  int
}

// NewSpatialAudioManager creates a new spatial audio manager
func NewSpatialAudioManager(backend AudioBackend, settings *AudioSettings) (*SpatialAudioManager, error) {
	sam := &SpatialAudioManager{
		backend:  backend,
		settings: settings,

		listenerPosition: Vector3{X: 0, Y: 0, Z: 0},
		listenerOrientation: ListenerOrientation{
			Forward: Vector3{X: 0, Y: 0, Z: -1},
			Up:      Vector3{X: 0, Y: 1, Z: 0},
		},

		spatialSounds:       make(map[string]*SpatialSoundInstance),
		ambientSounds:       make(map[string]*AmbientSoundInstance),
		audioZones:          make(map[string]*AudioZone),
		ambientLayers:       make(map[string]*AmbientLayer),

		globalMaxDistance:   settings.MaxAudioDistance,
		globalRolloffFactor: 1.0,
		dopplerFactor:       settings.DopplerFactor,

		maxSpatialSounds:  32,
		soundUpdateBudget: 8, // Update 8 sounds per frame

		environmentFX: &EnvironmentEffects{
			ReverbEnabled:    settings.IsEnabled("reverb"),
			OcclusionEnabled: true,
			DopplerEnabled:   true,
			AirAbsorption:    true,
		},
	}

	// Initialize default audio zones
	sam.initializeDefaultZones()

	// Initialize ambient layers
	sam.initializeAmbientLayers()

	return sam, nil
}

// initializeDefaultZones creates default audio zones
func (sam *SpatialAudioManager) initializeDefaultZones() {
	// Default outdoor zone
	outdoorZone := &AudioZone{
		Name: "outdoor",
		Bounds: BoundingBox{
			Min: Vector3{X: -1000, Y: -1000, Z: -1000},
			Max: Vector3{X: 1000, Y: 1000, Z: 1000},
		},
		Reverb: ReverbSettings{
			Enabled:  false,
			RoomSize: 0.1,
			Damping:  0.8,
			WetLevel: 0.1,
			DryLevel: 0.9,
		},
		Occlusion: OcclusionSettings{
			Enabled:         true,
			LowPassFilter:   1000.0,
			VolumeReduction: 0.5,
		},
		Material:    "grass",
		Humidity:    0.5,
		Temperature: 20.0,
		WindFactor:  1.0,
		AmbientSets: []string{"nature", "wind"},
	}

	sam.audioZones["outdoor"] = outdoorZone
	sam.currentZone = outdoorZone

	// Indoor/building zone template
	indoorZone := &AudioZone{
		Name: "indoor",
		Reverb: ReverbSettings{
			Enabled:  true,
			RoomSize: 0.5,
			Damping:  0.6,
			WetLevel: 0.3,
			DryLevel: 0.7,
		},
		Occlusion: OcclusionSettings{
			Enabled:         true,
			LowPassFilter:   800.0,
			VolumeReduction: 0.7,
		},
		Material:    "stone",
		Humidity:    0.3,
		Temperature: 18.0,
		WindFactor:  0.1,
		AmbientSets: []string{"indoor_ambient"},
	}

	sam.audioZones["indoor"] = indoorZone
}

// initializeAmbientLayers sets up ambient audio layers
func (sam *SpatialAudioManager) initializeAmbientLayers() {
	// Nature ambient layer
	natureLayer := &AmbientLayer{
		Name:     "nature",
		Sounds:   []*AmbientSoundInstance{},
		Volume:   1.0,
		IsActive: true,
		Priority: 1,
	}
	sam.ambientLayers["nature"] = natureLayer

	// Wind layer
	windLayer := &AmbientLayer{
		Name:     "wind",
		Sounds:   []*AmbientSoundInstance{},
		Volume:   0.7,
		IsActive: true,
		Priority: 2,
	}
	sam.ambientLayers["wind"] = windLayer

	// Combat ambient layer
	combatLayer := &AmbientLayer{
		Name:     "combat",
		Sounds:   []*AmbientSoundInstance{},
		Volume:   0.8,
		IsActive: false,
		Priority: 3,
	}
	sam.ambientLayers["combat"] = combatLayer
}

// Update processes spatial audio system updates
func (sam *SpatialAudioManager) Update() {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	if !sam.settings.IsEnabled("3d_audio") {
		return
	}

	// Update listener position and orientation in backend
	sam.backend.SetListenerPosition(sam.listenerPosition)
	sam.backend.SetListenerOrientation(sam.listenerOrientation.Forward, sam.listenerOrientation.Up)

	// Update spatial sounds (budget-limited)
	sam.updateSpatialSounds()

	// Update ambient sounds
	sam.updateAmbientSounds()

	// Update audio zones
	sam.updateAudioZones()

	// Update environmental effects
	sam.updateEnvironmentalEffects()
}

// updateSpatialSounds updates 3D positioned sounds with performance budgeting
func (sam *SpatialAudioManager) updateSpatialSounds() {
	soundIDs := make([]string, 0, len(sam.spatialSounds))
	for id := range sam.spatialSounds {
		soundIDs = append(soundIDs, id)
	}

	if len(soundIDs) == 0 {
		return
	}

	// Budget processing across multiple frames
	soundsToUpdate := sam.soundUpdateBudget
	if soundsToUpdate > len(soundIDs) {
		soundsToUpdate = len(soundIDs)
	}

	for i := 0; i < soundsToUpdate; i++ {
		index := (sam.lastUpdateIndex + i) % len(soundIDs)
		soundID := soundIDs[index]
		sound := sam.spatialSounds[soundID]

		sam.updateSpatialSoundInstance(sound)
	}

	sam.lastUpdateIndex = (sam.lastUpdateIndex + soundsToUpdate) % len(soundIDs)

	// Remove inactive sounds
	for id, sound := range sam.spatialSounds {
		if !sound.IsActive {
			sam.backend.StopSound(id)
			delete(sam.spatialSounds, id)
		}
	}
}

// updateSpatialSoundInstance updates a single 3D sound instance
func (sam *SpatialAudioManager) updateSpatialSoundInstance(sound *SpatialSoundInstance) {
	// Calculate distance to listener
	dx := sound.Position.X - sam.listenerPosition.X
	dy := sound.Position.Y - sam.listenerPosition.Y
	dz := sound.Position.Z - sam.listenerPosition.Z
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

	sound.DistanceToListener = distance

	// Check if sound is within audible range
	if distance > sound.MaxDistance {
		sound.IsActive = false
		return
	}

	// Calculate direction vector
	if distance > 0 {
		sound.Direction = Vector3{
			X: dx / distance,
			Y: dy / distance,
			Z: dz / distance,
		}
	}

	// Calculate volume attenuation based on distance
	volumeAttenuation := sam.calculateDistanceAttenuation(distance, sound.MinDistance, sound.MaxDistance, sound.AttenuationModel)

	// Apply occlusion if enabled
	if sam.environmentFX.OcclusionEnabled {
		occlusionFactor := sam.calculateOcclusion(sam.listenerPosition, sound.Position)
		volumeAttenuation *= occlusionFactor
	}

	sound.EffectiveVolume = sound.Volume * volumeAttenuation

	// Calculate doppler effect if enabled
	if sam.environmentFX.DopplerEnabled {
		sound.EffectivePitch = sam.calculateDopplerPitch(sound)
	} else {
		sound.EffectivePitch = sound.Pitch
	}

	// Update backend with new audio parameters
	// This would call backend-specific update methods in a real implementation

	sound.LastUpdate = time.Now()
}

// calculateDistanceAttenuation calculates volume attenuation based on distance
func (sam *SpatialAudioManager) calculateDistanceAttenuation(distance, minDistance, maxDistance float32, model string) float32 {
	if distance <= minDistance {
		return 1.0
	}

	if distance >= maxDistance {
		return 0.0
	}

	normalizedDistance := (distance - minDistance) / (maxDistance - minDistance)

	switch model {
	case "linear":
		return 1.0 - normalizedDistance

	case "inverse":
		// Inverse distance falloff (more realistic)
		return minDistance / distance

	case "exponential":
		// Exponential falloff
		return float32(math.Pow(float64(1.0-normalizedDistance), 2.0))

	default:
		return 1.0 - normalizedDistance // Default to linear
	}
}

// calculateOcclusion determines if sound is occluded by objects
func (sam *SpatialAudioManager) calculateOcclusion(listenerPos, soundPos Vector3) float32 {
	// This would perform raycasting to check for occlusion
	// For now, return 1.0 (no occlusion)
	// Real implementation would check against game world geometry
	return 1.0
}

// calculateDopplerPitch calculates pitch shift due to doppler effect
func (sam *SpatialAudioManager) calculateDopplerPitch(sound *SpatialSoundInstance) float32 {
	// Calculate relative velocity
	relativeVelocity := Vector3{
		X: sound.Velocity.X - sam.listenerVelocity.X,
		Y: sound.Velocity.Y - sam.listenerVelocity.Y,
		Z: sound.Velocity.Z - sam.listenerVelocity.Z,
	}

	// Project velocity onto direction vector
	radialVelocity := dot(relativeVelocity, sound.Direction)

	// Calculate doppler shift
	speedOfSound := float32(343.0) // m/s
	dopplerRatio := (speedOfSound - radialVelocity) / speedOfSound

	// Apply doppler factor setting
	adjustedRatio := 1.0 + (dopplerRatio-1.0)*sam.dopplerFactor

	return sound.Pitch * adjustedRatio
}

// updateAmbientSounds updates ambient sound layers
func (sam *SpatialAudioManager) updateAmbientSounds() {
	for _, layer := range sam.ambientLayers {
		if !layer.IsActive {
			continue
		}

		// Update ambient sounds in this layer
		for _, ambient := range layer.Sounds {
			if !ambient.IsActive {
				continue
			}

			// Calculate volume based on position and layer settings
			effectiveVolume := ambient.Volume * layer.Volume * sam.settings.GetEffectiveVolume("ambient")

			// For positioned ambient sounds, apply distance attenuation
			if ambient.Position != nil {
				distance := sam.calculateDistance(sam.listenerPosition, *ambient.Position)
				if distance <= ambient.Radius {
					// Fade in/out at edges of radius
					fadeZone := ambient.Radius * 0.2 // 20% fade zone
					if distance > ambient.Radius-fadeZone {
						fadeFactor := 1.0 - (distance-(ambient.Radius-fadeZone))/fadeZone
						effectiveVolume *= fadeFactor
					}
				} else {
					effectiveVolume = 0.0
				}
			}

			// Update backend volume
			// This would be implemented with backend-specific calls
		}
	}
}

// updateAudioZones checks if listener has moved to a different audio zone
func (sam *SpatialAudioManager) updateAudioZones() {
	for _, zone := range sam.audioZones {
		if sam.isPositionInBounds(sam.listenerPosition, zone.Bounds) {
			if sam.currentZone != zone {
				sam.transitionToZone(zone)
			}
			break
		}
	}
}

// updateEnvironmentalEffects updates environmental audio effects
func (sam *SpatialAudioManager) updateEnvironmentalEffects() {
	if sam.currentZone == nil {
		return
	}

	// Update reverb settings based on current zone
	if sam.environmentFX.ReverbEnabled && sam.currentZone.Reverb.Enabled {
		// Apply reverb settings to backend
		// This would be implemented with backend-specific reverb controls
	}

	// Update ambient layers based on zone
	for _, ambientSet := range sam.currentZone.AmbientSets {
		if layer, exists := sam.ambientLayers[ambientSet]; exists {
			layer.IsActive = true
		}
	}

	// Deactivate ambient layers not in current zone
	for name, layer := range sam.ambientLayers {
		inCurrentZone := false
		for _, setName := range sam.currentZone.AmbientSets {
			if name == setName {
				inCurrentZone = true
				break
			}
		}
		if !inCurrentZone {
			layer.IsActive = false
		}
	}
}

// PlaySpatialSound plays a 3D positioned sound
func (sam *SpatialAudioManager) PlaySpatialSound(event AudioEvent, position Vector3) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	if !sam.settings.IsEnabled("3d_audio") {
		return nil
	}

	// Check spatial sound limit
	if len(sam.spatialSounds) >= sam.maxSpatialSounds {
		// Remove oldest sound
		sam.removeOldestSpatialSound()
	}

	// Get sound name from metadata (for future use)
	_ = "default_sound"
	if event.Metadata != nil {
		if _, ok := event.Metadata["sound_name"].(string); ok {
			// Sound name retrieved but not used yet
		}
	}

	// Create spatial sound instance
	soundID := fmt.Sprintf("spatial_%d", time.Now().UnixNano())
	spatialSound := &SpatialSoundInstance{
		ID:               soundID,
		Position:         position,
		MinDistance:      1.0,
		MaxDistance:      sam.globalMaxDistance,
		AttenuationModel: "inverse",
		Volume:           event.Volume,
		Pitch:            1.0,
		IsActive:         true,
		IsLooping:        event.Loop,
		StartTime:        time.Now(),
	}

	// Play through backend
	err := sam.backend.PlaySound3D(nil, position) // Sound would be loaded first
	if err != nil {
		return fmt.Errorf("failed to play spatial sound: %w", err)
	}

	sam.spatialSounds[soundID] = spatialSound

	return nil
}

// PlayAmbientSound plays an ambient environmental sound
func (sam *SpatialAudioManager) PlayAmbientSound(event AudioEvent) error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	_ = "ambient_default"
	if event.Metadata != nil {
		if _, ok := event.Metadata["sound_name"].(string); ok {
			// Ambient sound name retrieved but not used yet
		}
	}

	// Create ambient sound instance
	soundID := fmt.Sprintf("ambient_%d", time.Now().UnixNano())
	ambientSound := &AmbientSoundInstance{
		ID:        soundID,
		Position:  event.Position,
		Volume:    event.Volume,
		IsActive:  true,
		LayerName: "nature", // Default layer
	}

	if event.Position != nil {
		ambientSound.Radius = 50.0 // Default ambient radius
	}

	sam.ambientSounds[soundID] = ambientSound

	return nil
}

// SetListenerPosition updates the listener's position
func (sam *SpatialAudioManager) SetListenerPosition(position Vector3) {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	sam.listenerPosition = position
}

// SetListenerOrientation updates the listener's orientation
func (sam *SpatialAudioManager) SetListenerOrientation(forward, up Vector3) {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	sam.listenerOrientation.Forward = forward
	sam.listenerOrientation.Up = up
}

// SetListenerVelocity updates the listener's velocity for doppler effect
func (sam *SpatialAudioManager) SetListenerVelocity(velocity Vector3) {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	sam.listenerVelocity = velocity
}

// CreateAudioZone creates a new audio zone
func (sam *SpatialAudioManager) CreateAudioZone(zone *AudioZone) {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	sam.audioZones[zone.Name] = zone
}

// SetWeatherIntensity sets the weather intensity for ambient sounds
func (sam *SpatialAudioManager) SetWeatherIntensity(intensity float32) {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	sam.weatherIntensity = clampFloat32(intensity, 0.0, 1.0)

	// Adjust wind and weather ambient layers
	if windLayer, exists := sam.ambientLayers["wind"]; exists {
		windLayer.Volume = 0.3 + sam.weatherIntensity*0.7
	}
}

// SetTimeOfDay sets the time of day for dynamic ambient sounds
func (sam *SpatialAudioManager) SetTimeOfDay(timeNormalized float32) {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	sam.timeOfDay = clampFloat32(timeNormalized, 0.0, 1.0)

	// Adjust ambient layers based on time of day
	// Night sounds more active during night hours
	nightFactor := 1.0 - float32(math.Abs(float64(sam.timeOfDay-0.0))) // Peak at midnight

	if natureLayer, exists := sam.ambientLayers["nature"]; exists {
		// Nature sounds vary with time of day
		natureLayer.Volume = 0.5 + nightFactor*0.5
	}
}

// Utility functions

func (sam *SpatialAudioManager) calculateDistance(a, b Vector3) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func (sam *SpatialAudioManager) isPositionInBounds(position Vector3, bounds BoundingBox) bool {
	return position.X >= bounds.Min.X && position.X <= bounds.Max.X &&
		position.Y >= bounds.Min.Y && position.Y <= bounds.Max.Y &&
		position.Z >= bounds.Min.Z && position.Z <= bounds.Max.Z
}

func (sam *SpatialAudioManager) transitionToZone(zone *AudioZone) {
	sam.currentZone = zone
	// Additional zone transition effects would be implemented here
}

func (sam *SpatialAudioManager) removeOldestSpatialSound() {
	var oldestID string
	var oldestTime time.Time = time.Now()

	for id, sound := range sam.spatialSounds {
		if sound.StartTime.Before(oldestTime) {
			oldestTime = sound.StartTime
			oldestID = id
		}
	}

	if oldestID != "" {
		sam.backend.StopSound(oldestID)
		delete(sam.spatialSounds, oldestID)
	}
}

func dot(a, b Vector3) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// GetStats returns spatial audio manager statistics
func (sam *SpatialAudioManager) GetStats() SpatialAudioStats {
	sam.mutex.RLock()
	defer sam.mutex.RUnlock()

	stats := SpatialAudioStats{
		SpatialSounds:     len(sam.spatialSounds),
		AmbientSounds:     len(sam.ambientSounds),
		AudioZones:        len(sam.audioZones),
		AmbientLayers:     len(sam.ambientLayers),
		ListenerPosition:  sam.listenerPosition,
		CurrentZone:       "",
		WeatherIntensity:  sam.weatherIntensity,
		TimeOfDay:         sam.timeOfDay,
	}

	if sam.currentZone != nil {
		stats.CurrentZone = sam.currentZone.Name
	}

	// Count active ambient layers
	for _, layer := range sam.ambientLayers {
		if layer.IsActive {
			stats.ActiveAmbientLayers++
		}
	}

	return stats
}

// SpatialAudioStats provides statistics about the spatial audio system
type SpatialAudioStats struct {
	SpatialSounds        int
	AmbientSounds        int
	AudioZones           int
	AmbientLayers        int
	ActiveAmbientLayers  int
	ListenerPosition     Vector3
	CurrentZone          string
	WeatherIntensity     float32
	TimeOfDay            float32
}

// Shutdown cleans up the spatial audio manager
func (sam *SpatialAudioManager) Shutdown() error {
	sam.mutex.Lock()
	defer sam.mutex.Unlock()

	// Stop all spatial sounds
	for id := range sam.spatialSounds {
		sam.backend.StopSound(id)
	}

	// Clear all data
	sam.spatialSounds = make(map[string]*SpatialSoundInstance)
	sam.ambientSounds = make(map[string]*AmbientSoundInstance)

	return nil
}