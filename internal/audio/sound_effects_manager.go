package audio

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// SoundEffectsManager manages sound effect playback
type SoundEffectsManager struct {
	backend    AudioBackend
	settings   *AudioSettings
	library    *SoundLibrary

	// Active sound management
	activeSounds    map[string]*SoundInstance
	soundPools      map[string]*SoundPool
	maxActiveSounds int

	// Categories
	uiSounds        map[string]string    // UI event -> sound ID mapping
	combatSounds    map[string][]string  // Combat event -> list of sound IDs (for variation)
	buildingSounds  map[string][]string  // Building event -> list of sound IDs
	resourceSounds  map[string][]string  // Resource event -> list of sound IDs
	environmentSounds map[string][]string // Environment event -> list of sound IDs

	// Playback control
	globalCooldowns map[string]time.Time // Prevent rapid fire of same sound
	categoryVolumes map[string]float32   // Per-category volume multipliers

	// Performance tracking
	soundsPlayedThisFrame int
	maxSoundsPerFrame     int

	mutex sync.RWMutex
}

// NewSoundEffectsManager creates a new sound effects manager
func NewSoundEffectsManager(backend AudioBackend, settings *AudioSettings) (*SoundEffectsManager, error) {
	sem := &SoundEffectsManager{
		backend:           backend,
		settings:          settings,
		library:           NewSoundLibrary(settings.AudioCacheSize),
		activeSounds:      make(map[string]*SoundInstance),
		soundPools:        make(map[string]*SoundPool),
		maxActiveSounds:   settings.MaxSimultaneousSounds,
		uiSounds:          make(map[string]string),
		combatSounds:      make(map[string][]string),
		buildingSounds:    make(map[string][]string),
		resourceSounds:    make(map[string][]string),
		environmentSounds: make(map[string][]string),
		globalCooldowns:   make(map[string]time.Time),
		categoryVolumes:   make(map[string]float32),
		maxSoundsPerFrame: 8,
	}

	// Initialize default sound mappings
	sem.initializeDefaultSounds()

	// Initialize category volumes
	sem.updateCategoryVolumes()

	return sem, nil
}

// initializeDefaultSounds sets up default sound event mappings
func (sem *SoundEffectsManager) initializeDefaultSounds() {
	// UI Sounds
	sem.uiSounds["click"] = "ui_click"
	sem.uiSounds["hover"] = "ui_hover"
	sem.uiSounds["error"] = "ui_error"
	sem.uiSounds["success"] = "ui_success"
	sem.uiSounds["button_press"] = "ui_button_press"
	sem.uiSounds["menu_open"] = "ui_menu_open"
	sem.uiSounds["menu_close"] = "ui_menu_close"
	sem.uiSounds["selection"] = "ui_selection"

	// Combat Sounds (with variations)
	sem.combatSounds["sword_attack"] = []string{"sword_swing_1", "sword_swing_2", "sword_swing_3"}
	sem.combatSounds["bow_attack"] = []string{"bow_release_1", "bow_release_2"}
	sem.combatSounds["unit_death"] = []string{"death_1", "death_2", "death_3", "death_4"}
	sem.combatSounds["unit_hit"] = []string{"impact_1", "impact_2", "impact_3"}
	sem.combatSounds["armor_clank"] = []string{"armor_1", "armor_2"}
	sem.combatSounds["explosion"] = []string{"explosion_small", "explosion_medium", "explosion_large"}
	sem.combatSounds["projectile_hit"] = []string{"arrow_hit", "magic_hit", "stone_hit"}

	// Building Sounds
	sem.buildingSounds["construction"] = []string{"hammer_1", "hammer_2", "construction_1"}
	sem.buildingSounds["construction_complete"] = []string{"build_complete_1", "build_complete_2"}
	sem.buildingSounds["building_destroy"] = []string{"building_collapse_1", "building_collapse_2"}
	sem.buildingSounds["production_complete"] = []string{"production_done_1", "production_done_2"}

	// Resource Sounds
	sem.resourceSounds["wood_chop"] = []string{"axe_chop_1", "axe_chop_2", "tree_fall"}
	sem.resourceSounds["mining"] = []string{"pickaxe_1", "pickaxe_2", "mining_1"}
	sem.resourceSounds["farming"] = []string{"harvest_1", "harvest_2"}
	sem.resourceSounds["resource_deposit"] = []string{"coins_1", "deposit_1"}

	// Environment Sounds
	sem.environmentSounds["footsteps"] = []string{"step_dirt_1", "step_dirt_2", "step_stone_1", "step_stone_2"}
	sem.environmentSounds["ambient_birds"] = []string{"birds_1", "birds_2"}
	sem.environmentSounds["wind"] = []string{"wind_light", "wind_strong"}
}

// Update processes sound effects system updates
func (sem *SoundEffectsManager) Update() {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	// Reset per-frame counters
	sem.soundsPlayedThisFrame = 0

	// Update active sounds
	sem.updateActiveSounds()

	// Clean up finished sounds
	sem.cleanupFinishedSounds()

	// Update category volumes if settings changed
	sem.updateCategoryVolumes()
}

// updateActiveSounds updates all currently playing sounds
func (sem *SoundEffectsManager) updateActiveSounds() {
	for _, instance := range sem.activeSounds {
		// Update fade effects
		if instance.IsFading {
			oldVolume := instance.Volume
			if instance.FadeSpeed != 0 {
				instance.Volume += instance.FadeSpeed * (1.0 / 60.0) // Assume 60 FPS

				// Check if fade is complete
				if (instance.FadeSpeed > 0 && instance.Volume >= instance.FadeTarget) ||
					(instance.FadeSpeed < 0 && instance.Volume <= instance.FadeTarget) {
					instance.Volume = instance.FadeTarget
					instance.IsFading = false

					// If faded to zero, mark for cleanup
					if instance.Volume <= 0 {
						instance.IsActive = false
					}
				}

				// Update backend volume if changed
				if oldVolume != instance.Volume {
					sem.backend.SetMasterVolume(instance.Volume) // This would be per-sound in real implementation
				}
			}
		}

		// Check if sound should still be active
		if !instance.IsActive || (!instance.IsLooping && time.Since(instance.StartTime) > instance.Sound.GetDuration()) {
			// Mark for cleanup
			instance.IsActive = false
		}
	}
}

// cleanupFinishedSounds removes inactive sound instances
func (sem *SoundEffectsManager) cleanupFinishedSounds() {
	for id, instance := range sem.activeSounds {
		if !instance.IsActive {
			// Stop sound in backend
			sem.backend.StopSound(id)

			// Remove from active sounds
			delete(sem.activeSounds, id)
		}
	}
}

// updateCategoryVolumes updates volume settings from audio settings
func (sem *SoundEffectsManager) updateCategoryVolumes() {
	sem.categoryVolumes["ui"] = sem.settings.GetEffectiveVolume("ui")
	sem.categoryVolumes["combat"] = sem.settings.GetEffectiveVolume("combat")
	sem.categoryVolumes["building"] = sem.settings.GetEffectiveVolume("sound_effects")
	sem.categoryVolumes["resource"] = sem.settings.GetEffectiveVolume("sound_effects")
	sem.categoryVolumes["environment"] = sem.settings.GetEffectiveVolume("ambient")
}

// PlayUISound plays a UI sound effect
func (sem *SoundEffectsManager) PlayUISound(event AudioEvent) error {
	if !sem.settings.IsEnabled("ui_audio") {
		return nil
	}

	soundName, exists := sem.uiSounds["click"] // Default to click for now
	if event.Metadata != nil {
		if name, ok := event.Metadata["sound_name"].(string); ok {
			soundName = name
		}
	}

	if !exists {
		return fmt.Errorf("UI sound not mapped: %s", soundName)
	}

	return sem.playSound(soundName, "ui", event.Volume)
}

// PlayCombatSound plays a combat sound effect
func (sem *SoundEffectsManager) PlayCombatSound(event AudioEvent) error {
	if !sem.settings.IsEnabled("sound_effects") {
		return nil
	}

	// Determine sound based on event type
	var soundVariants []string
	var exists bool

	switch event.Type {
	case AudioEventUnitAttack:
		soundVariants, exists = sem.combatSounds["sword_attack"]
	case AudioEventUnitDeath:
		soundVariants, exists = sem.combatSounds["unit_death"]
	case AudioEventExplosion:
		soundVariants, exists = sem.combatSounds["explosion"]
	default:
		soundVariants, exists = sem.combatSounds["sword_attack"] // Default
	}

	if !exists || len(soundVariants) == 0 {
		return fmt.Errorf("combat sound not available for event type: %d", event.Type)
	}

	// Select random variant
	soundName := soundVariants[rand.Intn(len(soundVariants))]

	return sem.playSound(soundName, "combat", event.Volume)
}

// PlayBuildingSound plays a building-related sound effect
func (sem *SoundEffectsManager) PlayBuildingSound(event AudioEvent) error {
	if !sem.settings.IsEnabled("sound_effects") {
		return nil
	}

	var soundVariants []string
	var exists bool

	switch event.Type {
	case AudioEventBuildingConstruction:
		soundVariants, exists = sem.buildingSounds["construction"]
	case AudioEventBuildingComplete:
		soundVariants, exists = sem.buildingSounds["construction_complete"]
	case AudioEventBuildingDestroy:
		soundVariants, exists = sem.buildingSounds["building_destroy"]
	case AudioEventProductionComplete:
		soundVariants, exists = sem.buildingSounds["production_complete"]
	default:
		soundVariants, exists = sem.buildingSounds["construction"]
	}

	if !exists || len(soundVariants) == 0 {
		return fmt.Errorf("building sound not available for event type: %d", event.Type)
	}

	// Select random variant
	soundName := soundVariants[rand.Intn(len(soundVariants))]

	return sem.playSound(soundName, "building", event.Volume)
}

// PlayResourceSound plays a resource-related sound effect
func (sem *SoundEffectsManager) PlayResourceSound(event AudioEvent) error {
	if !sem.settings.IsEnabled("sound_effects") {
		return nil
	}

	var soundVariants []string
	var exists bool

	switch event.Type {
	case AudioEventResourceGather:
		// Would determine resource type from metadata
		soundVariants, exists = sem.resourceSounds["wood_chop"] // Default
	case AudioEventResourceDeposit:
		soundVariants, exists = sem.resourceSounds["resource_deposit"]
	default:
		soundVariants, exists = sem.resourceSounds["wood_chop"]
	}

	if !exists || len(soundVariants) == 0 {
		return fmt.Errorf("resource sound not available for event type: %d", event.Type)
	}

	// Select random variant
	soundName := soundVariants[rand.Intn(len(soundVariants))]

	return sem.playSound(soundName, "resource", event.Volume)
}

// playSound plays a sound with the specified parameters
func (sem *SoundEffectsManager) playSound(soundID, category string, volume float32) error {
	// Check if we've exceeded per-frame limit
	if sem.soundsPlayedThisFrame >= sem.maxSoundsPerFrame {
		return nil // Silently ignore to prevent audio overload
	}

	// Check cooldown
	cooldownKey := fmt.Sprintf("%s_%s", category, soundID)
	if lastPlayed, exists := sem.globalCooldowns[cooldownKey]; exists {
		if time.Since(lastPlayed) < 50*time.Millisecond { // Minimum 50ms between same sounds
			return nil
		}
	}

	// Check if we have too many active sounds
	if len(sem.activeSounds) >= sem.maxActiveSounds {
		// Try to steal a sound (remove oldest or quietest)
		if !sem.stealSound() {
			return fmt.Errorf("too many active sounds and cannot steal")
		}
	}

	// Get sound from library
	sound, err := sem.library.GetSound(soundID)
	if err != nil {
		// Sound not loaded, would trigger async loading in real implementation
		return fmt.Errorf("sound not loaded: %s", soundID)
	}

	// Create sound instance
	instance := sound.Clone()

	// Apply volume settings
	categoryVolume := sem.categoryVolumes[category]
	instance.Volume = volume * categoryVolume

	// Start playback
	err = sem.backend.PlaySound(sound)
	if err != nil {
		return fmt.Errorf("failed to play sound: %w", err)
	}

	// Track active sound
	instance.IsActive = true
	sem.activeSounds[instance.ID] = instance

	// Update cooldown
	sem.globalCooldowns[cooldownKey] = time.Now()

	// Update counters
	sem.soundsPlayedThisFrame++

	return nil
}

// stealSound removes an active sound to make room for a new one
func (sem *SoundEffectsManager) stealSound() bool {
	if len(sem.activeSounds) == 0 {
		return false
	}

	// Find oldest sound
	var oldestID string
	var oldestTime time.Time = time.Now()

	for id, instance := range sem.activeSounds {
		if instance.StartTime.Before(oldestTime) {
			oldestTime = instance.StartTime
			oldestID = id
		}
	}

	if oldestID != "" {
		// Stop and remove oldest sound
		sem.backend.StopSound(oldestID)
		delete(sem.activeSounds, oldestID)
		return true
	}

	return false
}

// StopSound stops a specific sound
func (sem *SoundEffectsManager) StopSound(soundID string) error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	instance, exists := sem.activeSounds[soundID]
	if !exists {
		return fmt.Errorf("sound not active: %s", soundID)
	}

	// Stop in backend
	err := sem.backend.StopSound(soundID)
	if err != nil {
		return err
	}

	// Remove from active sounds
	delete(sem.activeSounds, soundID)
	instance.IsActive = false

	return nil
}

// StopAllSounds stops all currently playing sounds
func (sem *SoundEffectsManager) StopAllSounds() error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	var errors []error

	for id, instance := range sem.activeSounds {
		err := sem.backend.StopSound(id)
		if err != nil {
			errors = append(errors, err)
		}
		instance.IsActive = false
	}

	// Clear active sounds
	sem.activeSounds = make(map[string]*SoundInstance)

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping sounds: %v", errors)
	}

	return nil
}

// FadeOutSound gradually reduces a sound's volume to zero
func (sem *SoundEffectsManager) FadeOutSound(soundID string, duration float32) error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	instance, exists := sem.activeSounds[soundID]
	if !exists {
		return fmt.Errorf("sound not active: %s", soundID)
	}

	// Set up fade
	instance.IsFading = true
	instance.FadeTarget = 0.0
	instance.FadeSpeed = -instance.Volume / duration

	return nil
}

// SetVolume sets the volume for a specific sound category
func (sem *SoundEffectsManager) SetVolume(category string, volume float32) error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	sem.categoryVolumes[category] = volume

	// Update settings
	return sem.settings.SetVolume(category, volume)
}

// GetVolume gets the volume for a specific sound category
func (sem *SoundEffectsManager) GetVolume(category string) float32 {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	if volume, exists := sem.categoryVolumes[category]; exists {
		return volume
	}
	return 1.0
}

// LoadSoundsFromDirectory loads sound effects from a directory
func (sem *SoundEffectsManager) LoadSoundsFromDirectory(directory, category string) error {
	// This would scan the directory and load all audio files
	// Implementation placeholder
	return sem.library.LoadFromDirectory(directory, category, true)
}

// GetActiveCount returns the number of currently active sounds
func (sem *SoundEffectsManager) GetActiveCount() int {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()
	return len(sem.activeSounds)
}

// GetStats returns sound effects manager statistics
func (sem *SoundEffectsManager) GetStats() SoundEffectsStats {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	stats := SoundEffectsStats{
		ActiveSounds:          len(sem.activeSounds),
		MaxActiveSounds:       sem.maxActiveSounds,
		SoundsPlayedThisFrame: sem.soundsPlayedThisFrame,
		MaxSoundsPerFrame:     sem.maxSoundsPerFrame,
		LibraryStats:          sem.library.GetStats(),
	}

	// Count sounds by category
	stats.SoundsByCategory = make(map[string]int)
	for _, instance := range sem.activeSounds {
		stats.SoundsByCategory[instance.Sound.Category]++
	}

	return stats
}

// SoundEffectsStats provides statistics about the sound effects system
type SoundEffectsStats struct {
	ActiveSounds          int
	MaxActiveSounds       int
	SoundsPlayedThisFrame int
	MaxSoundsPerFrame     int
	SoundsByCategory      map[string]int
	LibraryStats          SoundLibraryStats
}

// Shutdown cleans up the sound effects manager
func (sem *SoundEffectsManager) Shutdown() error {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	// Stop all sounds
	err := sem.StopAllSounds()

	// Clear all data structures
	sem.activeSounds = make(map[string]*SoundInstance)
	sem.globalCooldowns = make(map[string]time.Time)

	return err
}

// RegisterCustomSound registers a custom sound mapping for events
func (sem *SoundEffectsManager) RegisterCustomSound(eventType, soundID string) {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	// This would register custom sound mappings
	// Implementation depends on how events are structured
	sem.uiSounds[eventType] = soundID
}

// SetSoundCooldown sets a cooldown period for a specific sound
func (sem *SoundEffectsManager) SetSoundCooldown(soundID string, cooldown time.Duration) {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	sem.globalCooldowns[soundID] = time.Now().Add(cooldown)
}