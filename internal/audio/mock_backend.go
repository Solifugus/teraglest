package audio

import (
	"fmt"
	"sync"
	"time"
)

// MockAudioBackend provides a mock implementation of AudioBackend for testing and development
// This allows the audio system to be tested without requiring actual audio libraries
type MockAudioBackend struct {
	// State tracking
	initialized     bool
	masterVolume    float32
	musicVolume     float32

	// Mock playback state
	activeSounds    map[string]*MockSoundPlayback
	currentMusic    *MockMusicPlayback

	// Listener state
	listenerPos     Vector3
	listenerForward Vector3
	listenerUp      Vector3

	// Performance tracking
	totalSoundsPlayed int
	totalMusicPlayed  int

	// Mock audio device settings
	sampleRate      int
	bufferSize      int
	latency         time.Duration

	mutex sync.RWMutex
}

// MockSoundPlayback represents a playing sound in the mock backend
type MockSoundPlayback struct {
	SoundID     string
	StartTime   time.Time
	Duration    time.Duration
	Volume      float32
	Pitch       float32
	IsLooping   bool
	Is3D        bool
	Position    Vector3
	IsActive    bool
}

// MockMusicPlayback represents playing music in the mock backend
type MockMusicPlayback struct {
	MusicID     string
	StartTime   time.Time
	Duration    time.Duration
	Volume      float32
	IsLooping   bool
	IsPaused    bool
	IsActive    bool
}

// NewMockAudioBackend creates a new mock audio backend
func NewMockAudioBackend() *MockAudioBackend {
	return &MockAudioBackend{
		masterVolume:  1.0,
		musicVolume:   1.0,
		activeSounds:  make(map[string]*MockSoundPlayback),
		sampleRate:    44100,
		bufferSize:    1024,
		latency:       23 * time.Millisecond, // Typical low-latency value
	}
}

// Initialize initializes the mock audio backend
func (m *MockAudioBackend) Initialize() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.initialized {
		return fmt.Errorf("audio backend already initialized")
	}

	// Simulate initialization time
	time.Sleep(10 * time.Millisecond)

	m.initialized = true
	fmt.Printf("[MockAudio] Backend initialized - Sample Rate: %d, Buffer: %d, Latency: %v\n",
		m.sampleRate, m.bufferSize, m.latency)

	return nil
}

// Shutdown shuts down the mock audio backend
func (m *MockAudioBackend) Shutdown() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	// Stop all active sounds
	for id := range m.activeSounds {
		delete(m.activeSounds, id)
	}

	// Stop music
	m.currentMusic = nil

	m.initialized = false
	fmt.Printf("[MockAudio] Backend shutdown complete\n")

	return nil
}

// PlaySound plays a sound effect
func (m *MockAudioBackend) PlaySound(sound *Sound) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	if sound == nil {
		return fmt.Errorf("sound is nil")
	}

	// Create mock playback instance
	soundID := fmt.Sprintf("sound_%d", time.Now().UnixNano())
	playback := &MockSoundPlayback{
		SoundID:   soundID,
		StartTime: time.Now(),
		Duration:  sound.Duration,
		Volume:    sound.DefaultVolume,
		Pitch:     sound.DefaultPitch,
		IsLooping: sound.CanLoop,
		Is3D:      sound.Is3D,
		IsActive:  true,
	}

	m.activeSounds[soundID] = playback
	m.totalSoundsPlayed++

	fmt.Printf("[MockAudio] Playing sound: %s (ID: %s, Volume: %.2f)\n",
		sound.Name, soundID, playback.Volume)

	// Simulate non-looping sounds finishing
	if !sound.CanLoop {
		go func() {
			time.Sleep(sound.Duration)
			m.mutex.Lock()
			delete(m.activeSounds, soundID)
			m.mutex.Unlock()
		}()
	}

	return nil
}

// StopSound stops a playing sound
func (m *MockAudioBackend) StopSound(soundID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	playback, exists := m.activeSounds[soundID]
	if !exists {
		return fmt.Errorf("sound not found: %s", soundID)
	}

	playback.IsActive = false
	delete(m.activeSounds, soundID)

	fmt.Printf("[MockAudio] Stopped sound: %s\n", soundID)
	return nil
}

// PauseSound pauses a playing sound
func (m *MockAudioBackend) PauseSound(soundID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	playback, exists := m.activeSounds[soundID]
	if !exists {
		return fmt.Errorf("sound not found: %s", soundID)
	}

	playback.IsActive = false
	fmt.Printf("[MockAudio] Paused sound: %s\n", soundID)
	return nil
}

// ResumeSound resumes a paused sound
func (m *MockAudioBackend) ResumeSound(soundID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	playback, exists := m.activeSounds[soundID]
	if !exists {
		return fmt.Errorf("sound not found: %s", soundID)
	}

	playback.IsActive = true
	fmt.Printf("[MockAudio] Resumed sound: %s\n", soundID)
	return nil
}

// PlayMusic starts playing background music
func (m *MockAudioBackend) PlayMusic(music *Music) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	if music == nil {
		return fmt.Errorf("music is nil")
	}

	// Stop current music if playing
	if m.currentMusic != nil {
		fmt.Printf("[MockAudio] Stopping current music: %s\n", m.currentMusic.MusicID)
	}

	// Create new music playback
	m.currentMusic = &MockMusicPlayback{
		MusicID:   music.ID,
		StartTime: time.Now(),
		Duration:  music.Duration,
		Volume:    music.DefaultVolume,
		IsLooping: music.CanLoop,
		IsPaused:  false,
		IsActive:  true,
	}

	m.totalMusicPlayed++

	fmt.Printf("[MockAudio] Playing music: %s (Duration: %v, Looping: %t)\n",
		music.Name, music.Duration, music.CanLoop)

	return nil
}

// StopMusic stops the current background music
func (m *MockAudioBackend) StopMusic() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	if m.currentMusic == nil {
		return fmt.Errorf("no music currently playing")
	}

	fmt.Printf("[MockAudio] Stopping music: %s\n", m.currentMusic.MusicID)
	m.currentMusic = nil

	return nil
}

// PauseMusicToggle toggles music pause state
func (m *MockAudioBackend) PauseMusicToggle() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	if m.currentMusic == nil {
		return fmt.Errorf("no music currently playing")
	}

	m.currentMusic.IsPaused = !m.currentMusic.IsPaused

	if m.currentMusic.IsPaused {
		fmt.Printf("[MockAudio] Paused music: %s\n", m.currentMusic.MusicID)
	} else {
		fmt.Printf("[MockAudio] Resumed music: %s\n", m.currentMusic.MusicID)
	}

	return nil
}

// SetMusicVolume sets the music volume
func (m *MockAudioBackend) SetMusicVolume(volume float32) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	m.musicVolume = volume

	if m.currentMusic != nil {
		m.currentMusic.Volume = volume
	}

	fmt.Printf("[MockAudio] Set music volume: %.2f\n", volume)
	return nil
}

// SetListenerPosition sets the 3D audio listener position
func (m *MockAudioBackend) SetListenerPosition(pos Vector3) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	m.listenerPos = pos

	// Only print if position changed significantly (avoid spam)
	fmt.Printf("[MockAudio] Listener position: (%.1f, %.1f, %.1f)\n",
		pos.X, pos.Y, pos.Z)

	return nil
}

// SetListenerOrientation sets the 3D audio listener orientation
func (m *MockAudioBackend) SetListenerOrientation(forward, up Vector3) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	m.listenerForward = forward
	m.listenerUp = up

	fmt.Printf("[MockAudio] Listener orientation - Forward: (%.2f, %.2f, %.2f), Up: (%.2f, %.2f, %.2f)\n",
		forward.X, forward.Y, forward.Z, up.X, up.Y, up.Z)

	return nil
}

// PlaySound3D plays a 3D positioned sound
func (m *MockAudioBackend) PlaySound3D(sound *Sound, position Vector3) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	// Create mock 3D sound playback
	soundID := fmt.Sprintf("3d_sound_%d", time.Now().UnixNano())
	playback := &MockSoundPlayback{
		SoundID:   soundID,
		StartTime: time.Now(),
		Duration:  5 * time.Second, // Default duration for mock
		Volume:    1.0,
		Pitch:     1.0,
		IsLooping: false,
		Is3D:      true,
		Position:  position,
		IsActive:  true,
	}

	m.activeSounds[soundID] = playback
	m.totalSoundsPlayed++

	// Calculate mock distance attenuation
	dx := position.X - m.listenerPos.X
	dy := position.Y - m.listenerPos.Y
	dz := position.Z - m.listenerPos.Z
	distance := float32(dx*dx + dy*dy + dz*dz) // Squared distance for performance

	fmt.Printf("[MockAudio] Playing 3D sound at (%.1f, %.1f, %.1f), Distance²: %.1f\n",
		position.X, position.Y, position.Z, distance)

	return nil
}

// SetMasterVolume sets the global master volume
func (m *MockAudioBackend) SetMasterVolume(volume float32) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.initialized {
		return fmt.Errorf("audio backend not initialized")
	}

	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	m.masterVolume = volume

	fmt.Printf("[MockAudio] Set master volume: %.2f\n", volume)
	return nil
}

// IsInitialized returns whether the backend is initialized
func (m *MockAudioBackend) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.initialized
}

// GetStats returns mock backend statistics
func (m *MockAudioBackend) GetStats() MockBackendStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := MockBackendStats{
		Initialized:       m.initialized,
		ActiveSounds:      len(m.activeSounds),
		TotalSoundsPlayed: m.totalSoundsPlayed,
		TotalMusicPlayed:  m.totalMusicPlayed,
		MasterVolume:      m.masterVolume,
		MusicVolume:       m.musicVolume,
		ListenerPosition:  m.listenerPos,
		SampleRate:        m.sampleRate,
		BufferSize:        m.bufferSize,
		Latency:           m.latency,
	}

	if m.currentMusic != nil {
		stats.MusicPlaying = true
		stats.CurrentMusicID = m.currentMusic.MusicID
		stats.MusicPaused = m.currentMusic.IsPaused
	}

	return stats
}

// MockBackendStats provides statistics about the mock backend
type MockBackendStats struct {
	Initialized       bool
	ActiveSounds      int
	TotalSoundsPlayed int
	TotalMusicPlayed  int
	MasterVolume      float32
	MusicVolume       float32
	ListenerPosition  Vector3
	MusicPlaying      bool
	MusicPaused       bool
	CurrentMusicID    string
	SampleRate        int
	BufferSize        int
	Latency           time.Duration
}

// GetActiveSounds returns information about currently active sounds
func (m *MockAudioBackend) GetActiveSounds() map[string]*MockSoundPlayback {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return a copy to prevent external modification
	activeSounds := make(map[string]*MockSoundPlayback)
	for id, playback := range m.activeSounds {
		soundCopy := *playback
		activeSounds[id] = &soundCopy
	}

	return activeSounds
}

// SimulateLatency simulates audio latency for testing
func (m *MockAudioBackend) SimulateLatency(duration time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.latency = duration
	fmt.Printf("[MockAudio] Simulating latency: %v\n", duration)
}

// SetBufferSize sets the audio buffer size for testing different configurations
func (m *MockAudioBackend) SetBufferSize(size int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if size < 64 || size > 4096 {
		return fmt.Errorf("invalid buffer size: %d (must be 64-4096)", size)
	}

	m.bufferSize = size

	// Recalculate latency based on buffer size
	m.latency = time.Duration(float64(size)/float64(m.sampleRate)*1000) * time.Millisecond

	fmt.Printf("[MockAudio] Set buffer size: %d (Latency: %v)\n", size, m.latency)
	return nil
}

// SetSampleRate sets the sample rate for testing different configurations
func (m *MockAudioBackend) SetSampleRate(rate int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	validRates := []int{22050, 44100, 48000, 96000}
	valid := false
	for _, validRate := range validRates {
		if rate == validRate {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid sample rate: %d", rate)
	}

	m.sampleRate = rate

	// Recalculate latency based on new sample rate
	m.latency = time.Duration(float64(m.bufferSize)/float64(rate)*1000) * time.Millisecond

	fmt.Printf("[MockAudio] Set sample rate: %d Hz (Latency: %v)\n", rate, m.latency)
	return nil
}

// Clean up finished sounds (called periodically)
func (m *MockAudioBackend) CleanupFinishedSounds() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	cleaned := 0

	for id, playback := range m.activeSounds {
		if !playback.IsLooping && now.Sub(playback.StartTime) > playback.Duration {
			delete(m.activeSounds, id)
			cleaned++
		}
	}

	if cleaned > 0 {
		fmt.Printf("[MockAudio] Cleaned up %d finished sounds\n", cleaned)
	}
}

// PrintStatus prints the current backend status (useful for debugging)
func (m *MockAudioBackend) PrintStatus() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	fmt.Printf("\n[MockAudio] Backend Status:\n")
	fmt.Printf("  Initialized: %t\n", m.initialized)
	fmt.Printf("  Active Sounds: %d\n", len(m.activeSounds))
	fmt.Printf("  Music Playing: %t\n", m.currentMusic != nil)
	if m.currentMusic != nil {
		fmt.Printf("  Current Music: %s (Paused: %t)\n", m.currentMusic.MusicID, m.currentMusic.IsPaused)
	}
	fmt.Printf("  Master Volume: %.2f\n", m.masterVolume)
	fmt.Printf("  Music Volume: %.2f\n", m.musicVolume)
	fmt.Printf("  Listener: (%.1f, %.1f, %.1f)\n", m.listenerPos.X, m.listenerPos.Y, m.listenerPos.Z)
	fmt.Printf("  Total Sounds Played: %d\n", m.totalSoundsPlayed)
	fmt.Printf("  Total Music Played: %d\n", m.totalMusicPlayed)
	fmt.Printf("  Sample Rate: %d Hz\n", m.sampleRate)
	fmt.Printf("  Buffer Size: %d\n", m.bufferSize)
	fmt.Printf("  Latency: %v\n", m.latency)
	fmt.Printf("\n")
}