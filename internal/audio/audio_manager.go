package audio

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AudioManager is the central audio system manager for TeraGlest
type AudioManager struct {
	// Core components
	soundEffects *SoundEffectsManager
	music        *MusicManager
	spatialAudio *SpatialAudioManager
	settings     *AudioSettings

	// State management
	enabled      bool
	masterVolume float32
	mutex        sync.RWMutex

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Audio backend
	backend AudioBackend

	// Event callbacks
	eventCallbacks map[AudioEventType][]AudioEventCallback
}

// AudioEventType represents different audio events in the game
type AudioEventType int

const (
	// UI Events
	AudioEventUIClick AudioEventType = iota
	AudioEventUIHover
	AudioEventUIError
	AudioEventUISuccess

	// Combat Events
	AudioEventUnitAttack
	AudioEventUnitDeath
	AudioEventUnitMove
	AudioEventProjectileFire
	AudioEventExplosion

	// Building Events
	AudioEventBuildingConstruction
	AudioEventBuildingComplete
	AudioEventBuildingDestroy
	AudioEventProductionComplete

	// Resource Events
	AudioEventResourceGather
	AudioEventResourceDeposit

	// Environment Events
	AudioEventEnvironmentAmbient
	AudioEventWeatherChange

	// Music Events
	AudioEventMusicCombat
	AudioEventMusicPeace
	AudioEventMusicVictory
	AudioEventMusicDefeat
)

// AudioEventCallback defines callback function for audio events
type AudioEventCallback func(event AudioEvent)

// AudioEvent contains information about an audio event
type AudioEvent struct {
	Type      AudioEventType
	Position  *Vector3 // Optional 3D position
	Volume    float32
	Pitch     float32
	Loop      bool
	Metadata  map[string]interface{}
	Timestamp time.Time
}

// AudioBackend interface for different audio implementations
type AudioBackend interface {
	Initialize() error
	Shutdown() error

	// Sound playback
	PlaySound(sound *Sound) error
	StopSound(soundID string) error
	PauseSound(soundID string) error
	ResumeSound(soundID string) error

	// Music playback
	PlayMusic(music *Music) error
	StopMusic() error
	PauseMusicToggle() error
	SetMusicVolume(volume float32) error

	// 3D Audio
	SetListenerPosition(pos Vector3) error
	SetListenerOrientation(forward, up Vector3) error
	PlaySound3D(sound *Sound, position Vector3) error

	// General
	SetMasterVolume(volume float32) error
	IsInitialized() bool
}

// Vector3 represents a 3D position for spatial audio
type Vector3 struct {
	X, Y, Z float32
}

// NewAudioManager creates and initializes the audio system
func NewAudioManager(backend AudioBackend) (*AudioManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	am := &AudioManager{
		backend:        backend,
		enabled:        true,
		masterVolume:   1.0,
		ctx:            ctx,
		cancel:         cancel,
		eventCallbacks: make(map[AudioEventType][]AudioEventCallback),
	}

	// Initialize audio backend
	if err := backend.Initialize(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize audio backend: %w", err)
	}

	// Initialize sub-managers
	if err := am.initializeSubsystems(); err != nil {
		cancel()
		backend.Shutdown()
		return nil, fmt.Errorf("failed to initialize audio subsystems: %w", err)
	}

	// Start audio update loop
	go am.updateLoop()

	return am, nil
}

// initializeSubsystems creates and initializes all audio subsystems
func (am *AudioManager) initializeSubsystems() error {
	// Initialize settings
	settings, err := NewAudioSettings()
	if err != nil {
		return fmt.Errorf("failed to create audio settings: %w", err)
	}
	am.settings = settings

	// Initialize sound effects manager
	soundEffects, err := NewSoundEffectsManager(am.backend, settings)
	if err != nil {
		return fmt.Errorf("failed to create sound effects manager: %w", err)
	}
	am.soundEffects = soundEffects

	// Initialize music manager
	music, err := NewMusicManager(am.backend, settings)
	if err != nil {
		return fmt.Errorf("failed to create music manager: %w", err)
	}
	am.music = music

	// Initialize spatial audio manager
	spatialAudio, err := NewSpatialAudioManager(am.backend, settings)
	if err != nil {
		return fmt.Errorf("failed to create spatial audio manager: %w", err)
	}
	am.spatialAudio = spatialAudio

	return nil
}

// updateLoop runs the main audio update loop
func (am *AudioManager) updateLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	for {
		select {
		case <-am.ctx.Done():
			return
		case <-ticker.C:
			am.update()
		}
	}
}

// update processes audio system updates
func (am *AudioManager) update() {
	am.mutex.RLock()
	enabled := am.enabled
	am.mutex.RUnlock()

	if !enabled {
		return
	}

	// Update subsystems
	if am.soundEffects != nil {
		am.soundEffects.Update()
	}
	if am.music != nil {
		am.music.Update()
	}
	if am.spatialAudio != nil {
		am.spatialAudio.Update()
	}
}

// TriggerEvent triggers an audio event
func (am *AudioManager) TriggerEvent(eventType AudioEventType, event AudioEvent) {
	am.mutex.RLock()
	enabled := am.enabled
	callbacks := am.eventCallbacks[eventType]
	am.mutex.RUnlock()

	if !enabled {
		return
	}

	// Set timestamp
	event.Type = eventType
	event.Timestamp = time.Now()

	// Execute callbacks
	for _, callback := range callbacks {
		go callback(event) // Non-blocking callback execution
	}

	// Handle event based on type
	am.handleAudioEvent(event)
}

// handleAudioEvent processes specific audio events
func (am *AudioManager) handleAudioEvent(event AudioEvent) {
	switch event.Type {
	// UI Events
	case AudioEventUIClick, AudioEventUIHover, AudioEventUIError, AudioEventUISuccess:
		am.soundEffects.PlayUISound(event)

	// Combat Events
	case AudioEventUnitAttack, AudioEventUnitDeath, AudioEventProjectileFire, AudioEventExplosion:
		if event.Position != nil {
			am.spatialAudio.PlaySpatialSound(event, *event.Position)
		} else {
			am.soundEffects.PlayCombatSound(event)
		}

	// Building Events
	case AudioEventBuildingConstruction, AudioEventBuildingComplete, AudioEventBuildingDestroy:
		if event.Position != nil {
			am.spatialAudio.PlaySpatialSound(event, *event.Position)
		} else {
			am.soundEffects.PlayBuildingSound(event)
		}

	// Resource Events
	case AudioEventResourceGather, AudioEventResourceDeposit:
		if event.Position != nil {
			am.spatialAudio.PlaySpatialSound(event, *event.Position)
		} else {
			am.soundEffects.PlayResourceSound(event)
		}

	// Music Events
	case AudioEventMusicCombat, AudioEventMusicPeace, AudioEventMusicVictory, AudioEventMusicDefeat:
		am.music.HandleMusicEvent(event)

	// Environment Events
	case AudioEventEnvironmentAmbient:
		am.spatialAudio.PlayAmbientSound(event)
	}
}

// RegisterEventCallback registers a callback for audio events
func (am *AudioManager) RegisterEventCallback(eventType AudioEventType, callback AudioEventCallback) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.eventCallbacks[eventType] = append(am.eventCallbacks[eventType], callback)
}

// SetEnabled enables or disables the audio system
func (am *AudioManager) SetEnabled(enabled bool) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.enabled = enabled
}

// SetMasterVolume sets the master volume (0.0 - 1.0)
func (am *AudioManager) SetMasterVolume(volume float32) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	am.masterVolume = volume
	return am.backend.SetMasterVolume(volume)
}

// GetMasterVolume returns the current master volume
func (am *AudioManager) GetMasterVolume() float32 {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.masterVolume
}

// SetListenerPosition updates the 3D audio listener position
func (am *AudioManager) SetListenerPosition(position Vector3) error {
	return am.backend.SetListenerPosition(position)
}

// SetListenerOrientation updates the 3D audio listener orientation
func (am *AudioManager) SetListenerOrientation(forward, up Vector3) error {
	return am.backend.SetListenerOrientation(forward, up)
}

// GetSoundEffectsManager returns the sound effects manager
func (am *AudioManager) GetSoundEffectsManager() *SoundEffectsManager {
	return am.soundEffects
}

// GetMusicManager returns the music manager
func (am *AudioManager) GetMusicManager() *MusicManager {
	return am.music
}

// GetSpatialAudioManager returns the spatial audio manager
func (am *AudioManager) GetSpatialAudioManager() *SpatialAudioManager {
	return am.spatialAudio
}

// GetSettings returns the audio settings
func (am *AudioManager) GetSettings() *AudioSettings {
	return am.settings
}

// PlayUISound plays a UI sound effect
func (am *AudioManager) PlayUISound(soundName string, volume float32) error {
	if !am.enabled {
		return nil
	}
	return am.soundEffects.PlayUISound(AudioEvent{
		Type:   AudioEventUIClick,
		Volume: volume,
		Metadata: map[string]interface{}{
			"sound_name": soundName,
		},
	})
}

// PlayCombatSound plays a combat sound at a 3D position
func (am *AudioManager) PlayCombatSound(soundName string, position Vector3, volume float32) error {
	if !am.enabled {
		return nil
	}
	return am.spatialAudio.PlaySpatialSound(AudioEvent{
		Type:     AudioEventUnitAttack,
		Position: &position,
		Volume:   volume,
		Metadata: map[string]interface{}{
			"sound_name": soundName,
		},
	}, position)
}

// PlayMusic starts playing background music
func (am *AudioManager) PlayMusic(musicName string) error {
	if !am.enabled {
		return nil
	}
	return am.music.PlayMusic(musicName)
}

// StopMusic stops the current background music
func (am *AudioManager) StopMusic() error {
	return am.music.StopMusic()
}

// Shutdown gracefully shuts down the audio system
func (am *AudioManager) Shutdown() error {
	am.cancel()

	var errors []error

	// Shutdown subsystems
	if am.music != nil {
		if err := am.music.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("music manager shutdown: %w", err))
		}
	}

	if am.soundEffects != nil {
		if err := am.soundEffects.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("sound effects shutdown: %w", err))
		}
	}

	if am.spatialAudio != nil {
		if err := am.spatialAudio.Shutdown(); err != nil {
			errors = append(errors, fmt.Errorf("spatial audio shutdown: %w", err))
		}
	}

	// Shutdown backend
	if err := am.backend.Shutdown(); err != nil {
		errors = append(errors, fmt.Errorf("backend shutdown: %w", err))
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("audio system shutdown errors: %v", errors)
	}

	return nil
}

// IsEnabled returns whether the audio system is enabled
func (am *AudioManager) IsEnabled() bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	return am.enabled
}

// GetStats returns audio system statistics
func (am *AudioManager) GetStats() AudioStats {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	stats := AudioStats{
		Enabled:         am.enabled,
		MasterVolume:    am.masterVolume,
		BackendActive:   am.backend.IsInitialized(),
		RegisteredEvents: len(am.eventCallbacks),
	}

	if am.soundEffects != nil {
		stats.ActiveSounds = am.soundEffects.GetActiveCount()
	}

	if am.music != nil {
		stats.MusicPlaying = am.music.IsPlaying()
	}

	return stats
}

// AudioStats provides information about the audio system state
type AudioStats struct {
	Enabled          bool
	MasterVolume     float32
	BackendActive    bool
	ActiveSounds     int
	MusicPlaying     bool
	RegisteredEvents int
}