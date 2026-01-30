package audio

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Sound represents a sound effect that can be played
type Sound struct {
	// Identification
	ID       string
	Name     string
	FilePath string
	Category string

	// Audio properties
	Duration    time.Duration
	SampleRate  int
	Channels    int
	BitDepth    int
	FileSize    int64

	// Playback properties
	DefaultVolume float32
	DefaultPitch  float32
	CanLoop       bool
	Priority      int // Higher priority sounds can interrupt lower priority ones

	// 3D Audio properties
	Is3D             bool
	MinDistance      float32 // Distance where volume starts to decrease
	MaxDistance      float32 // Distance where sound becomes inaudible
	AttenuationModel string  // linear, inverse, exponential

	// Streaming properties
	IsStreamed    bool // True for large files that should be streamed
	PreloadBuffer []byte // Buffer for non-streamed sounds

	// State
	IsLoaded    bool
	LoadedAt    time.Duration
	PlayCount   int
	mutex       sync.RWMutex

	// Backend-specific data
	BackendData interface{}
}

// Music represents a music track
type Music struct {
	// Identification
	ID       string
	Name     string
	FilePath string
	Category string // combat, peace, victory, defeat, ambient

	// Audio properties
	Duration   time.Duration
	SampleRate int
	Channels   int
	BitDepth   int
	FileSize   int64

	// Playback properties
	DefaultVolume    float32
	CanLoop          bool
	IntroLength      time.Duration // Length of intro before looping portion
	LoopStartTime    time.Duration // When the loop portion starts
	FadeInDuration   time.Duration
	FadeOutDuration  time.Duration

	// Music-specific properties
	Tempo           int     // BPM for synchronization
	Key             string  // Musical key
	Mood            string  // peaceful, intense, mysterious, etc.
	TransitionPoints []time.Duration // Times where smooth transitions are possible

	// State
	IsLoaded     bool
	playing      bool
	CurrentTime  time.Duration
	Volume       float32
	IsFading     bool
	FadeTarget   float32
	FadeSpeed    float32
	mutex        sync.RWMutex

	// Backend-specific data
	BackendData interface{}
}

// SoundInstance represents a playing instance of a sound
type SoundInstance struct {
	ID         string
	Sound      *Sound
	StartTime  time.Time
	Volume     float32
	Pitch      float32
	IsLooping  bool
	IsPaused   bool
	Position   *Vector3 // 3D position if applicable
	Velocity   *Vector3 // For doppler effect

	// State tracking
	IsActive   bool
	IsFading   bool
	FadeTarget float32
	FadeSpeed  float32

	// Backend-specific instance data
	BackendInstance interface{}
}

// SoundPool manages a pool of similar sounds for efficient playback
type SoundPool struct {
	Name           string
	BaseSounds     []*Sound
	MaxInstances   int
	ActiveInstances []*SoundInstance
	StealingMode   string // oldest, quietest, lowest_priority

	mutex sync.RWMutex
}

// SoundLibrary manages the collection of loaded sounds
type SoundLibrary struct {
	sounds     map[string]*Sound
	music      map[string]*Music
	soundPools map[string]*SoundPool
	categories map[string][]*Sound

	loadedSounds  int
	totalMemory   int64
	maxMemory     int64

	mutex sync.RWMutex
}

// AudioFileFormat represents supported audio file formats
type AudioFileFormat string

const (
	FormatWAV  AudioFileFormat = ".wav"
	FormatOGG  AudioFileFormat = ".ogg"
	FormatMP3  AudioFileFormat = ".mp3"
	FormatFLAC AudioFileFormat = ".flac"
)

// NewSound creates a new sound from file path
func NewSound(id, filePath, category string) (*Sound, error) {
	// Validate file format
	ext := strings.ToLower(filepath.Ext(filePath))
	if !isValidAudioFormat(ext) {
		return nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	sound := &Sound{
		ID:               id,
		Name:             filepath.Base(filePath),
		FilePath:         filePath,
		Category:         category,
		DefaultVolume:    1.0,
		DefaultPitch:     1.0,
		CanLoop:          true,
		Priority:         5,
		Is3D:             false,
		MinDistance:      1.0,
		MaxDistance:      50.0,
		AttenuationModel: "inverse",
		IsStreamed:       false,
		IsLoaded:         false,
		PlayCount:        0,
	}

	// Determine if sound should be streamed (large files)
	// This would be set based on file size analysis
	sound.IsStreamed = shouldStreamAudio(ext, 0) // File size would be checked here

	return sound, nil
}

// NewMusic creates a new music track from file path
func NewMusic(id, filePath, category string) (*Music, error) {
	// Validate file format
	ext := strings.ToLower(filepath.Ext(filePath))
	if !isValidAudioFormat(ext) {
		return nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	music := &Music{
		ID:               id,
		Name:             filepath.Base(filePath),
		FilePath:         filePath,
		Category:         category,
		DefaultVolume:    1.0,
		CanLoop:          true,
		FadeInDuration:   2 * time.Second,
		FadeOutDuration:  2 * time.Second,
		Tempo:            120,
		Key:              "C",
		Mood:             "neutral",
		IsLoaded:         false,
		playing:          false,
		Volume:           1.0,
		IsFading:         false,
	}

	return music, nil
}

// NewSoundLibrary creates a new sound library
func NewSoundLibrary(maxMemoryMB int) *SoundLibrary {
	return &SoundLibrary{
		sounds:     make(map[string]*Sound),
		music:      make(map[string]*Music),
		soundPools: make(map[string]*SoundPool),
		categories: make(map[string][]*Sound),
		maxMemory:  int64(maxMemoryMB) * 1024 * 1024,
	}
}

// LoadSound loads a sound into the library
func (sl *SoundLibrary) LoadSound(id, filePath, category string) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	// Check if already loaded
	if _, exists := sl.sounds[id]; exists {
		return fmt.Errorf("sound with ID %s already exists", id)
	}

	// Create sound
	sound, err := NewSound(id, filePath, category)
	if err != nil {
		return fmt.Errorf("failed to create sound: %w", err)
	}

	// Store in library
	sl.sounds[id] = sound

	// Add to category
	sl.categories[category] = append(sl.categories[category], sound)

	sl.loadedSounds++

	return nil
}

// LoadMusic loads a music track into the library
func (sl *SoundLibrary) LoadMusic(id, filePath, category string) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	// Check if already loaded
	if _, exists := sl.music[id]; exists {
		return fmt.Errorf("music with ID %s already exists", id)
	}

	// Create music
	music, err := NewMusic(id, filePath, category)
	if err != nil {
		return fmt.Errorf("failed to create music: %w", err)
	}

	// Store in library
	sl.music[id] = music

	return nil
}

// GetSound retrieves a sound by ID
func (sl *SoundLibrary) GetSound(id string) (*Sound, error) {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	sound, exists := sl.sounds[id]
	if !exists {
		return nil, fmt.Errorf("sound not found: %s", id)
	}

	return sound, nil
}

// GetMusic retrieves a music track by ID
func (sl *SoundLibrary) GetMusic(id string) (*Music, error) {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	music, exists := sl.music[id]
	if !exists {
		return nil, fmt.Errorf("music not found: %s", id)
	}

	return music, nil
}

// GetSoundsInCategory returns all sounds in a category
func (sl *SoundLibrary) GetSoundsInCategory(category string) []*Sound {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	sounds := sl.categories[category]
	if sounds == nil {
		return []*Sound{}
	}

	// Return copy to prevent external modification
	result := make([]*Sound, len(sounds))
	copy(result, sounds)
	return result
}

// CreateSoundPool creates a sound pool for efficient playback of similar sounds
func (sl *SoundLibrary) CreateSoundPool(name string, soundIDs []string, maxInstances int) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	// Check if pool already exists
	if _, exists := sl.soundPools[name]; exists {
		return fmt.Errorf("sound pool %s already exists", name)
	}

	// Collect sounds
	var sounds []*Sound
	for _, id := range soundIDs {
		sound, exists := sl.sounds[id]
		if !exists {
			return fmt.Errorf("sound not found for pool: %s", id)
		}
		sounds = append(sounds, sound)
	}

	// Create pool
	pool := &SoundPool{
		Name:            name,
		BaseSounds:      sounds,
		MaxInstances:    maxInstances,
		ActiveInstances: make([]*SoundInstance, 0, maxInstances),
		StealingMode:    "oldest",
	}

	sl.soundPools[name] = pool
	return nil
}

// GetSoundPool retrieves a sound pool by name
func (sl *SoundLibrary) GetSoundPool(name string) (*SoundPool, error) {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	pool, exists := sl.soundPools[name]
	if !exists {
		return nil, fmt.Errorf("sound pool not found: %s", name)
	}

	return pool, nil
}

// UnloadSound removes a sound from the library
func (sl *SoundLibrary) UnloadSound(id string) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	sound, exists := sl.sounds[id]
	if !exists {
		return fmt.Errorf("sound not found: %s", id)
	}

	// Remove from category
	category := sound.Category
	if categoryList, exists := sl.categories[category]; exists {
		for i, s := range categoryList {
			if s.ID == id {
				sl.categories[category] = append(categoryList[:i], categoryList[i+1:]...)
				break
			}
		}
	}

	// Remove from main map
	delete(sl.sounds, id)
	sl.loadedSounds--

	return nil
}

// GetStats returns library statistics
func (sl *SoundLibrary) GetStats() SoundLibraryStats {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	return SoundLibraryStats{
		LoadedSounds:   sl.loadedSounds,
		LoadedMusic:    len(sl.music),
		SoundPools:     len(sl.soundPools),
		Categories:     len(sl.categories),
		TotalMemory:    sl.totalMemory,
		MaxMemory:      sl.maxMemory,
		MemoryUsage:    float32(sl.totalMemory) / float32(sl.maxMemory),
	}
}

// SoundLibraryStats provides statistics about the sound library
type SoundLibraryStats struct {
	LoadedSounds int
	LoadedMusic  int
	SoundPools   int
	Categories   int
	TotalMemory  int64
	MaxMemory    int64
	MemoryUsage  float32
}

// LoadFromDirectory loads all audio files from a directory
func (sl *SoundLibrary) LoadFromDirectory(directory, category string, recursive bool) error {
	// Implementation would scan directory and load audio files
	// This is a placeholder that would be implemented with actual file I/O
	return fmt.Errorf("directory loading not implemented yet")
}

// LoadFromManifest loads sounds and music from a manifest file
func (sl *SoundLibrary) LoadFromManifest(manifestPath string) error {
	// Implementation would parse a manifest file (JSON/XML) and load specified audio files
	// This is a placeholder that would be implemented with actual file parsing
	return fmt.Errorf("manifest loading not implemented yet")
}

// Utility functions

func isValidAudioFormat(ext string) bool {
	validFormats := []string{".wav", ".ogg", ".mp3", ".flac"}
	for _, format := range validFormats {
		if ext == format {
			return true
		}
	}
	return false
}

func shouldStreamAudio(format string, fileSize int64) bool {
	// Larger files should be streamed to save memory
	// Music files are typically streamed, sound effects are preloaded
	if format == ".mp3" && fileSize > 1024*1024 { // > 1MB
		return true
	}
	if format == ".ogg" && fileSize > 512*1024 { // > 512KB
		return true
	}
	return false
}

// Sound instance management methods

// Clone creates a new instance of this sound for playback
func (s *Sound) Clone() *SoundInstance {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	instance := &SoundInstance{
		ID:        fmt.Sprintf("%s_%d_%d", s.ID, s.PlayCount, time.Now().UnixNano()),
		Sound:     s,
		StartTime: time.Now(),
		Volume:    s.DefaultVolume,
		Pitch:     s.DefaultPitch,
		IsLooping: false,
		IsPaused:  false,
		IsActive:  false,
	}

	s.PlayCount++
	return instance
}

// GetDuration returns the duration of the sound (placeholder)
func (s *Sound) GetDuration() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.Duration
}

// SetBackendData stores backend-specific data
func (s *Sound) SetBackendData(data interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.BackendData = data
}

// GetBackendData retrieves backend-specific data
func (s *Sound) GetBackendData() interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.BackendData
}

// Music playback methods

// SetCurrentTime updates the current playback time
func (m *Music) SetCurrentTime(time time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.CurrentTime = time
}

// GetCurrentTime returns the current playback time
func (m *Music) GetCurrentTime() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.CurrentTime
}

// SetVolume sets the music volume
func (m *Music) SetVolume(volume float32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}
	m.Volume = volume
}

// GetVolume returns the current music volume
func (m *Music) GetVolume() float32 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Volume
}

// StartFade begins a volume fade
func (m *Music) StartFade(targetVolume, duration float32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.IsFading = true
	m.FadeTarget = targetVolume
	m.FadeSpeed = (targetVolume - m.Volume) / duration // Volume change per second
}

// UpdateFade updates the fade effect (called from update loop)
func (m *Music) UpdateFade(deltaTime float32) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.IsFading {
		return false
	}

	oldVolume := m.Volume
	m.Volume += m.FadeSpeed * deltaTime

	// Check if fade is complete
	if (m.FadeSpeed > 0 && m.Volume >= m.FadeTarget) ||
		(m.FadeSpeed < 0 && m.Volume <= m.FadeTarget) {
		m.Volume = m.FadeTarget
		m.IsFading = false
		return true // Fade complete
	}

	return oldVolume != m.Volume // Volume changed
}

// IsPlaying returns whether the music is currently playing
func (m *Music) IsPlaying() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.playing
}