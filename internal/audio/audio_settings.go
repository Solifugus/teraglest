package audio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// AudioSettings manages audio configuration and preferences
type AudioSettings struct {
	// Volume controls (0.0 - 1.0)
	MasterVolume      float32 `json:"master_volume"`
	SoundEffectVolume float32 `json:"sound_effect_volume"`
	MusicVolume       float32 `json:"music_volume"`
	UIVolume          float32 `json:"ui_volume"`
	CombatVolume      float32 `json:"combat_volume"`
	AmbientVolume     float32 `json:"ambient_volume"`

	// Audio quality settings
	SampleRate        int    `json:"sample_rate"`
	BufferSize        int    `json:"buffer_size"`
	AudioDevice       string `json:"audio_device"`
	EnableSurround    bool   `json:"enable_surround"`

	// 3D Audio settings
	Enable3DAudio     bool    `json:"enable_3d_audio"`
	DistanceModel     string  `json:"distance_model"`
	DopplerFactor     float32 `json:"doppler_factor"`
	MaxAudioDistance  float32 `json:"max_audio_distance"`

	// Music settings
	EnableMusic       bool `json:"enable_music"`
	MusicTransition   bool `json:"music_transition"`
	CombatMusicEnabled bool `json:"combat_music_enabled"`

	// Sound settings
	EnableSoundEffects bool `json:"enable_sound_effects"`
	MaxSimultaneousSounds int `json:"max_simultaneous_sounds"`
	EnableEcho         bool `json:"enable_echo"`
	EnableReverb       bool `json:"enable_reverb"`

	// UI settings
	EnableUIAudio      bool `json:"enable_ui_audio"`
	UIClickVolume      float32 `json:"ui_click_volume"`
	UIHoverVolume      float32 `json:"ui_hover_volume"`

	// Performance settings
	EnableAudioStreaming bool `json:"enable_audio_streaming"`
	AudioCacheSize      int  `json:"audio_cache_size"`
	LowLatencyMode      bool `json:"low_latency_mode"`

	// File management
	configPath string
	mutex      sync.RWMutex
}

// AudioQualityPreset defines audio quality presets
type AudioQualityPreset int

const (
	AudioQualityLow AudioQualityPreset = iota
	AudioQualityMedium
	AudioQualityHigh
	AudioQualityUltra
)

// NewAudioSettings creates a new audio settings instance with defaults
func NewAudioSettings() (*AudioSettings, error) {
	settings := &AudioSettings{
		// Default volume settings
		MasterVolume:      1.0,
		SoundEffectVolume: 0.8,
		MusicVolume:       0.6,
		UIVolume:          0.7,
		CombatVolume:      0.9,
		AmbientVolume:     0.5,

		// Default quality settings
		SampleRate:        44100,
		BufferSize:        1024,
		AudioDevice:       "", // Use default device
		EnableSurround:    false,

		// Default 3D audio settings
		Enable3DAudio:     true,
		DistanceModel:     "inverse_distance",
		DopplerFactor:     1.0,
		MaxAudioDistance:  100.0,

		// Default music settings
		EnableMusic:       true,
		MusicTransition:   true,
		CombatMusicEnabled: true,

		// Default sound settings
		EnableSoundEffects: true,
		MaxSimultaneousSounds: 32,
		EnableEcho:         false,
		EnableReverb:       false,

		// Default UI settings
		EnableUIAudio:     true,
		UIClickVolume:     0.8,
		UIHoverVolume:     0.4,

		// Default performance settings
		EnableAudioStreaming: true,
		AudioCacheSize:      64, // MB
		LowLatencyMode:       false,
	}

	// Set config path
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to current directory
		settings.configPath = filepath.Join(".", "audio_settings.json")
	} else {
		appDir := filepath.Join(configDir, "teraglest")
		os.MkdirAll(appDir, 0755)
		settings.configPath = filepath.Join(appDir, "audio_settings.json")
	}

	// Try to load existing settings
	if err := settings.Load(); err != nil {
		// If loading fails, save defaults
		if saveErr := settings.Save(); saveErr != nil {
			return nil, fmt.Errorf("failed to save default settings: %w", saveErr)
		}
	}

	return settings, nil
}

// Load loads audio settings from file
func (as *AudioSettings) Load() error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	data, err := os.ReadFile(as.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, use defaults
		}
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	if err := json.Unmarshal(data, as); err != nil {
		return fmt.Errorf("failed to parse settings file: %w", err)
	}

	// Validate loaded settings
	as.validateAndFix()

	return nil
}

// Save saves audio settings to file
func (as *AudioSettings) Save() error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	// Validate settings before saving
	as.validateAndFix()

	data, err := json.MarshalIndent(as, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(as.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// validateAndFix ensures settings are within valid ranges
func (as *AudioSettings) validateAndFix() {
	// Clamp volume values to valid range
	as.MasterVolume = clampFloat32(as.MasterVolume, 0.0, 1.0)
	as.SoundEffectVolume = clampFloat32(as.SoundEffectVolume, 0.0, 1.0)
	as.MusicVolume = clampFloat32(as.MusicVolume, 0.0, 1.0)
	as.UIVolume = clampFloat32(as.UIVolume, 0.0, 1.0)
	as.CombatVolume = clampFloat32(as.CombatVolume, 0.0, 1.0)
	as.AmbientVolume = clampFloat32(as.AmbientVolume, 0.0, 1.0)
	as.UIClickVolume = clampFloat32(as.UIClickVolume, 0.0, 1.0)
	as.UIHoverVolume = clampFloat32(as.UIHoverVolume, 0.0, 1.0)

	// Validate sample rate
	validSampleRates := []int{22050, 44100, 48000, 96000}
	if !contains(validSampleRates, as.SampleRate) {
		as.SampleRate = 44100
	}

	// Validate buffer size (power of 2)
	if as.BufferSize < 256 || as.BufferSize > 4096 || !isPowerOfTwo(as.BufferSize) {
		as.BufferSize = 1024
	}

	// Validate distance model
	validDistanceModels := []string{"linear_distance", "inverse_distance", "exponential_distance"}
	if !containsString(validDistanceModels, as.DistanceModel) {
		as.DistanceModel = "inverse_distance"
	}

	// Validate 3D audio settings
	as.DopplerFactor = clampFloat32(as.DopplerFactor, 0.0, 5.0)
	as.MaxAudioDistance = clampFloat32(as.MaxAudioDistance, 10.0, 1000.0)

	// Validate performance settings
	if as.MaxSimultaneousSounds < 8 || as.MaxSimultaneousSounds > 128 {
		as.MaxSimultaneousSounds = 32
	}

	if as.AudioCacheSize < 16 || as.AudioCacheSize > 512 {
		as.AudioCacheSize = 64
	}
}

// SetQualityPreset applies a predefined quality preset
func (as *AudioSettings) SetQualityPreset(preset AudioQualityPreset) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	switch preset {
	case AudioQualityLow:
		as.SampleRate = 22050
		as.BufferSize = 2048
		as.MaxSimultaneousSounds = 16
		as.EnableEcho = false
		as.EnableReverb = false
		as.Enable3DAudio = false
		as.EnableAudioStreaming = false
		as.AudioCacheSize = 32

	case AudioQualityMedium:
		as.SampleRate = 44100
		as.BufferSize = 1024
		as.MaxSimultaneousSounds = 24
		as.EnableEcho = false
		as.EnableReverb = false
		as.Enable3DAudio = true
		as.EnableAudioStreaming = true
		as.AudioCacheSize = 48

	case AudioQualityHigh:
		as.SampleRate = 44100
		as.BufferSize = 512
		as.MaxSimultaneousSounds = 32
		as.EnableEcho = true
		as.EnableReverb = true
		as.Enable3DAudio = true
		as.EnableAudioStreaming = true
		as.AudioCacheSize = 64

	case AudioQualityUltra:
		as.SampleRate = 48000
		as.BufferSize = 256
		as.MaxSimultaneousSounds = 64
		as.EnableEcho = true
		as.EnableReverb = true
		as.Enable3DAudio = true
		as.EnableSurround = true
		as.EnableAudioStreaming = true
		as.AudioCacheSize = 128
		as.LowLatencyMode = true
	}
}

// GetEffectiveVolume calculates the effective volume for a category
func (as *AudioSettings) GetEffectiveVolume(category string) float32 {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	var categoryVolume float32
	switch category {
	case "sound_effects":
		categoryVolume = as.SoundEffectVolume
	case "music":
		categoryVolume = as.MusicVolume
	case "ui":
		categoryVolume = as.UIVolume
	case "combat":
		categoryVolume = as.CombatVolume
	case "ambient":
		categoryVolume = as.AmbientVolume
	default:
		categoryVolume = 1.0
	}

	return as.MasterVolume * categoryVolume
}

// SetVolume sets volume for a specific category
func (as *AudioSettings) SetVolume(category string, volume float32) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	volume = clampFloat32(volume, 0.0, 1.0)

	switch category {
	case "master":
		as.MasterVolume = volume
	case "sound_effects":
		as.SoundEffectVolume = volume
	case "music":
		as.MusicVolume = volume
	case "ui":
		as.UIVolume = volume
	case "combat":
		as.CombatVolume = volume
	case "ambient":
		as.AmbientVolume = volume
	default:
		return fmt.Errorf("unknown volume category: %s", category)
	}

	return nil
}

// GetVolume gets volume for a specific category
func (as *AudioSettings) GetVolume(category string) float32 {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	switch category {
	case "master":
		return as.MasterVolume
	case "sound_effects":
		return as.SoundEffectVolume
	case "music":
		return as.MusicVolume
	case "ui":
		return as.UIVolume
	case "combat":
		return as.CombatVolume
	case "ambient":
		return as.AmbientVolume
	default:
		return 1.0
	}
}

// IsEnabled checks if a feature is enabled
func (as *AudioSettings) IsEnabled(feature string) bool {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	switch feature {
	case "music":
		return as.EnableMusic
	case "sound_effects":
		return as.EnableSoundEffects
	case "3d_audio":
		return as.Enable3DAudio
	case "ui_audio":
		return as.EnableUIAudio
	case "echo":
		return as.EnableEcho
	case "reverb":
		return as.EnableReverb
	case "surround":
		return as.EnableSurround
	case "streaming":
		return as.EnableAudioStreaming
	case "combat_music":
		return as.CombatMusicEnabled
	case "music_transition":
		return as.MusicTransition
	case "low_latency":
		return as.LowLatencyMode
	default:
		return false
	}
}

// SetEnabled enables or disables a feature
func (as *AudioSettings) SetEnabled(feature string, enabled bool) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	switch feature {
	case "music":
		as.EnableMusic = enabled
	case "sound_effects":
		as.EnableSoundEffects = enabled
	case "3d_audio":
		as.Enable3DAudio = enabled
	case "ui_audio":
		as.EnableUIAudio = enabled
	case "echo":
		as.EnableEcho = enabled
	case "reverb":
		as.EnableReverb = enabled
	case "surround":
		as.EnableSurround = enabled
	case "streaming":
		as.EnableAudioStreaming = enabled
	case "combat_music":
		as.CombatMusicEnabled = enabled
	case "music_transition":
		as.MusicTransition = enabled
	case "low_latency":
		as.LowLatencyMode = enabled
	default:
		return fmt.Errorf("unknown feature: %s", feature)
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func (as *AudioSettings) GetConfigPath() string {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.configPath
}

// Reset resets all settings to defaults
func (as *AudioSettings) Reset() {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	// Reset to default values (same as constructor)
	as.MasterVolume = 1.0
	as.SoundEffectVolume = 0.8
	as.MusicVolume = 0.6
	as.UIVolume = 0.7
	as.CombatVolume = 0.9
	as.AmbientVolume = 0.5

	as.SampleRate = 44100
	as.BufferSize = 1024
	as.AudioDevice = ""
	as.EnableSurround = false

	as.Enable3DAudio = true
	as.DistanceModel = "inverse_distance"
	as.DopplerFactor = 1.0
	as.MaxAudioDistance = 100.0

	as.EnableMusic = true
	as.MusicTransition = true
	as.CombatMusicEnabled = true

	as.EnableSoundEffects = true
	as.MaxSimultaneousSounds = 32
	as.EnableEcho = false
	as.EnableReverb = false

	as.EnableUIAudio = true
	as.UIClickVolume = 0.8
	as.UIHoverVolume = 0.4

	as.EnableAudioStreaming = true
	as.AudioCacheSize = 64
	as.LowLatencyMode = false
}

// Utility functions
func clampFloat32(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}