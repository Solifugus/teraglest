package audio

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// MusicManager handles background music playback and transitions
type MusicManager struct {
	backend  AudioBackend
	settings *AudioSettings
	library  *SoundLibrary

	// Current music state
	currentMusic    *Music
	nextMusic       *Music
	isPlaying       bool
	isPaused        bool
	isTransitioning bool

	// Volume and fading
	currentVolume  float32
	targetVolume   float32
	fadeSpeed      float32
	isFading       bool

	// Music categories and playlists
	musicTracks     map[string]*Music           // All loaded music tracks
	playlists       map[string][]*Music         // Organized playlists
	categoryTracks  map[string][]*Music         // Tracks by category
	currentPlaylist string

	// Playback settings
	shuffleMode      bool
	loopMode         MusicLoopMode
	crossfadeEnabled bool
	crossfadeDuration time.Duration

	// Adaptive music system
	currentMood     MusicMood
	combatIntensity float32 // 0.0 = peaceful, 1.0 = intense combat
	stealthMode     bool
	explorationMode bool

	// Transition system
	transitionQueue    []*MusicTransition
	transitionCooldown time.Duration
	lastTransitionTime time.Time

	// Performance tracking
	playbackPosition time.Duration
	totalPlayTime    time.Duration

	mutex sync.RWMutex
}

// MusicLoopMode defines how music should loop
type MusicLoopMode int

const (
	LoopNone MusicLoopMode = iota // Play once
	LoopTrack                     // Loop current track
	LoopPlaylist                  // Loop entire playlist
	LoopShuffle                   // Shuffle and loop playlist
)

// MusicMood represents the current game mood for adaptive music
type MusicMood int

const (
	MoodPeaceful MusicMood = iota
	MoodTense
	MoodCombat
	MoodVictory
	MoodDefeat
	MoodExploration
	MoodStealth
	MoodBuilding
)

// MusicTransition defines a music transition
type MusicTransition struct {
	FromMusic    *Music
	ToMusic      *Music
	Duration     time.Duration
	FadeType     FadeType
	Crossfade    bool
	TriggerTime  time.Duration // When in the source track to trigger
}

// FadeType defines how music fades during transitions
type FadeType int

const (
	FadeLinear FadeType = iota
	FadeExponential
	FadeSmoothStep
	FadeCrossfade
)

// NewMusicManager creates a new music manager
func NewMusicManager(backend AudioBackend, settings *AudioSettings) (*MusicManager, error) {
	mm := &MusicManager{
		backend:          backend,
		settings:         settings,
		library:          NewSoundLibrary(settings.AudioCacheSize),
		musicTracks:      make(map[string]*Music),
		playlists:        make(map[string][]*Music),
		categoryTracks:   make(map[string][]*Music),
		currentVolume:    settings.GetEffectiveVolume("music"),
		targetVolume:     settings.GetEffectiveVolume("music"),
		loopMode:         LoopTrack,
		crossfadeEnabled: settings.IsEnabled("music_transition"),
		crossfadeDuration: 3 * time.Second,
		currentMood:      MoodPeaceful,
		transitionCooldown: 5 * time.Second,
	}

	// Initialize default playlists
	mm.initializeDefaultPlaylists()

	return mm, nil
}

// initializeDefaultPlaylists sets up default music organization
func (mm *MusicManager) initializeDefaultPlaylists() {
	// Create default playlists
	mm.playlists["main_menu"] = []*Music{}
	mm.playlists["peaceful"] = []*Music{}
	mm.playlists["combat"] = []*Music{}
	mm.playlists["victory"] = []*Music{}
	mm.playlists["defeat"] = []*Music{}
	mm.playlists["exploration"] = []*Music{}
	mm.playlists["building"] = []*Music{}

	// Initialize category tracks
	mm.categoryTracks["peace"] = []*Music{}
	mm.categoryTracks["combat"] = []*Music{}
	mm.categoryTracks["victory"] = []*Music{}
	mm.categoryTracks["defeat"] = []*Music{}
	mm.categoryTracks["ambient"] = []*Music{}
}

// Update processes music system updates
func (mm *MusicManager) Update() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.settings.IsEnabled("music") {
		if mm.isPlaying {
			mm.stopCurrentMusic()
		}
		return
	}

	// Update current music playback position
	if mm.isPlaying && mm.currentMusic != nil {
		// This would be updated by the audio backend in a real implementation
		mm.playbackPosition += time.Duration(16) * time.Millisecond // ~60 FPS
		mm.totalPlayTime += time.Duration(16) * time.Millisecond
	}

	// Process volume fading
	mm.processFading()

	// Check for music transitions
	mm.processTransitions()

	// Handle music loop/repeat logic
	mm.processLooping()

	// Update adaptive music based on game state
	mm.updateAdaptiveMusic()
}

// processFading handles volume fade effects
func (mm *MusicManager) processFading() {
	if !mm.isFading || mm.currentMusic == nil {
		return
	}

	deltaTime := float32(16) / 1000.0 // 16ms in seconds
	oldVolume := mm.currentVolume

	// Update volume based on fade speed
	mm.currentVolume += mm.fadeSpeed * deltaTime

	// Check if fade is complete
	if (mm.fadeSpeed > 0 && mm.currentVolume >= mm.targetVolume) ||
		(mm.fadeSpeed < 0 && mm.currentVolume <= mm.targetVolume) {
		mm.currentVolume = mm.targetVolume
		mm.isFading = false

		// Handle fade completion
		if mm.currentVolume <= 0.0 {
			mm.stopCurrentMusic()
		}
	}

	// Clamp volume
	if mm.currentVolume < 0.0 {
		mm.currentVolume = 0.0
	}
	if mm.currentVolume > 1.0 {
		mm.currentVolume = 1.0
	}

	// Update backend volume if changed
	if oldVolume != mm.currentVolume {
		mm.backend.SetMusicVolume(mm.currentVolume * mm.settings.GetEffectiveVolume("music"))
	}
}

// processTransitions handles music transitions
func (mm *MusicManager) processTransitions() {
	if len(mm.transitionQueue) == 0 {
		return
	}

	// Check if enough time has passed since last transition
	if time.Since(mm.lastTransitionTime) < mm.transitionCooldown {
		return
	}

	// Process next transition
	transition := mm.transitionQueue[0]
	mm.transitionQueue = mm.transitionQueue[1:]

	mm.executeTransition(transition)
}

// executeTransition performs a music transition
func (mm *MusicManager) executeTransition(transition *MusicTransition) {
	if transition.Crossfade && mm.crossfadeEnabled {
		mm.executeCrossfadeTransition(transition)
	} else {
		mm.executeFadeTransition(transition)
	}

	mm.lastTransitionTime = time.Now()
}

// executeFadeTransition performs a fade-out/fade-in transition
func (mm *MusicManager) executeFadeTransition(transition *MusicTransition) {
	fadeDuration := float32(transition.Duration.Seconds())

	if mm.isPlaying {
		// Fade out current music
		mm.startFade(0.0, fadeDuration/2)

		// Schedule fade-in of new music
		go func() {
			time.Sleep(transition.Duration / 2)
			mm.playMusicInternal(transition.ToMusic)
			mm.startFade(mm.targetVolume, fadeDuration/2)
		}()
	} else {
		// Just start new music
		mm.playMusicInternal(transition.ToMusic)
		mm.startFade(mm.targetVolume, fadeDuration)
	}
}

// executeCrossfadeTransition performs a crossfade transition
func (mm *MusicManager) executeCrossfadeTransition(transition *MusicTransition) {
	// This would implement crossfading between two audio streams
	// For now, fallback to fade transition
	mm.executeFadeTransition(transition)
}

// processLooping handles music looping logic
func (mm *MusicManager) processLooping() {
	if !mm.isPlaying || mm.currentMusic == nil {
		return
	}

	// Check if current track has finished
	trackFinished := mm.playbackPosition >= mm.currentMusic.Duration

	if trackFinished {
		switch mm.loopMode {
		case LoopNone:
			mm.stopCurrentMusic()

		case LoopTrack:
			mm.restartCurrentMusic()

		case LoopPlaylist, LoopShuffle:
			mm.playNextInPlaylist()
		}
	}
}

// updateAdaptiveMusic adjusts music based on game state
func (mm *MusicManager) updateAdaptiveMusic() {
	// Check if mood has changed and requires different music
	desiredCategory := mm.getMusicCategoryForMood(mm.currentMood)

	if mm.currentMusic == nil || mm.currentMusic.Category != desiredCategory {
		// Need to transition to appropriate music
		if tracks, exists := mm.categoryTracks[desiredCategory]; exists && len(tracks) > 0 {
			newTrack := mm.selectTrackForMood(tracks, mm.currentMood)
			if newTrack != nil && newTrack != mm.currentMusic {
				mm.transitionToMusic(newTrack, 2*time.Second)
			}
		}
	}

	// Adjust volume based on combat intensity
	if mm.currentMood == MoodCombat {
		intensityVolume := 0.7 + (mm.combatIntensity * 0.3) // Range from 0.7 to 1.0
		targetVolume := mm.settings.GetEffectiveVolume("music") * intensityVolume

		if math.Abs(float64(targetVolume-mm.currentVolume)) > 0.05 {
			mm.startFade(targetVolume, 1.0) // 1 second fade
		}
	}
}

// getMusicCategoryForMood returns the appropriate music category for a mood
func (mm *MusicManager) getMusicCategoryForMood(mood MusicMood) string {
	switch mood {
	case MoodPeaceful, MoodExploration, MoodBuilding:
		return "peace"
	case MoodTense, MoodStealth:
		return "ambient"
	case MoodCombat:
		return "combat"
	case MoodVictory:
		return "victory"
	case MoodDefeat:
		return "defeat"
	default:
		return "peace"
	}
}

// selectTrackForMood selects an appropriate track for the current mood
func (mm *MusicManager) selectTrackForMood(tracks []*Music, mood MusicMood) *Music {
	if len(tracks) == 0 {
		return nil
	}

	// Filter tracks by mood compatibility
	var suitableTracks []*Music
	for _, track := range tracks {
		if mm.isTrackSuitableForMood(track, mood) {
			suitableTracks = append(suitableTracks, track)
		}
	}

	if len(suitableTracks) == 0 {
		suitableTracks = tracks // Fallback to any track
	}

	// Select track (random, or based on recent history)
	if mm.shuffleMode {
		return suitableTracks[time.Now().UnixNano()%int64(len(suitableTracks))]
	}

	// Return first suitable track for now
	return suitableTracks[0]
}

// isTrackSuitableForMood determines if a track is suitable for the current mood
func (mm *MusicManager) isTrackSuitableForMood(track *Music, mood MusicMood) bool {
	// This would analyze track properties like tempo, key, mood metadata
	// For now, use simple category matching
	return track.Category == mm.getMusicCategoryForMood(mood)
}

// PlayMusic plays a specific music track
func (mm *MusicManager) PlayMusic(musicID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	music, err := mm.library.GetMusic(musicID)
	if err != nil {
		return fmt.Errorf("music not found: %s", musicID)
	}

	return mm.playMusicInternal(music)
}

// playMusicInternal plays music without locking (internal use)
func (mm *MusicManager) playMusicInternal(music *Music) error {
	if !mm.settings.IsEnabled("music") {
		return nil
	}

	// Stop current music if playing
	if mm.isPlaying {
		mm.stopCurrentMusic()
	}

	// Start new music
	err := mm.backend.PlayMusic(music)
	if err != nil {
		return fmt.Errorf("failed to play music: %w", err)
	}

	// Update state
	mm.currentMusic = music
	mm.isPlaying = true
	mm.isPaused = false
	mm.playbackPosition = 0

	// Set initial volume
	mm.currentVolume = mm.targetVolume
	mm.backend.SetMusicVolume(mm.currentVolume * mm.settings.GetEffectiveVolume("music"))

	return nil
}

// StopMusic stops the current music
func (mm *MusicManager) StopMusic() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	return mm.stopCurrentMusic()
}

// stopCurrentMusic stops current music without locking (internal use)
func (mm *MusicManager) stopCurrentMusic() error {
	if !mm.isPlaying {
		return nil
	}

	err := mm.backend.StopMusic()
	if err != nil {
		return err
	}

	mm.currentMusic = nil
	mm.isPlaying = false
	mm.isPaused = false
	mm.playbackPosition = 0

	return nil
}

// PauseMusic pauses or resumes music playback
func (mm *MusicManager) PauseMusic() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isPlaying {
		return fmt.Errorf("no music currently playing")
	}

	err := mm.backend.PauseMusicToggle()
	if err != nil {
		return err
	}

	mm.isPaused = !mm.isPaused
	return nil
}

// SetVolume sets the music volume
func (mm *MusicManager) SetVolume(volume float32) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if volume < 0.0 {
		volume = 0.0
	}
	if volume > 1.0 {
		volume = 1.0
	}

	mm.targetVolume = volume

	if !mm.isFading {
		mm.currentVolume = volume
		if mm.isPlaying {
			return mm.backend.SetMusicVolume(volume * mm.settings.GetEffectiveVolume("music"))
		}
	}

	return nil
}

// GetVolume returns the current music volume
func (mm *MusicManager) GetVolume() float32 {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.currentVolume
}

// startFade begins a volume fade
func (mm *MusicManager) startFade(targetVolume, durationSeconds float32) {
	mm.isFading = true
	mm.targetVolume = targetVolume
	mm.fadeSpeed = (targetVolume - mm.currentVolume) / durationSeconds
}

// FadeOut fades out the current music over the specified duration
func (mm *MusicManager) FadeOut(duration time.Duration) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isPlaying {
		return fmt.Errorf("no music currently playing")
	}

	mm.startFade(0.0, float32(duration.Seconds()))
	return nil
}

// FadeIn fades in the current music over the specified duration
func (mm *MusicManager) FadeIn(duration time.Duration) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if !mm.isPlaying {
		return fmt.Errorf("no music currently playing")
	}

	mm.startFade(mm.targetVolume, float32(duration.Seconds()))
	return nil
}

// SetMood sets the current music mood for adaptive playback
func (mm *MusicManager) SetMood(mood MusicMood) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if mm.currentMood != mood {
		mm.currentMood = mood
		// Mood change will be processed in updateAdaptiveMusic
	}
}

// SetCombatIntensity sets the combat intensity level (0.0 - 1.0)
func (mm *MusicManager) SetCombatIntensity(intensity float32) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if intensity < 0.0 {
		intensity = 0.0
	}
	if intensity > 1.0 {
		intensity = 1.0
	}

	mm.combatIntensity = intensity
}

// transitionToMusic transitions to a new music track
func (mm *MusicManager) transitionToMusic(music *Music, duration time.Duration) {
	transition := &MusicTransition{
		FromMusic: mm.currentMusic,
		ToMusic:   music,
		Duration:  duration,
		FadeType:  FadeLinear,
		Crossfade: mm.crossfadeEnabled,
	}

	mm.transitionQueue = append(mm.transitionQueue, transition)
}

// HandleMusicEvent handles music-related audio events
func (mm *MusicManager) HandleMusicEvent(event AudioEvent) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	switch event.Type {
	case AudioEventMusicCombat:
		mm.SetMood(MoodCombat)
	case AudioEventMusicPeace:
		mm.SetMood(MoodPeaceful)
	case AudioEventMusicVictory:
		mm.SetMood(MoodVictory)
	case AudioEventMusicDefeat:
		mm.SetMood(MoodDefeat)
	}
}

// LoadMusicFromDirectory loads music from a directory
func (mm *MusicManager) LoadMusicFromDirectory(directory, category string) error {
	// This would scan directory and load music files
	return mm.library.LoadFromDirectory(directory, category, true)
}

// CreatePlaylist creates a new playlist
func (mm *MusicManager) CreatePlaylist(name string, musicIDs []string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	var tracks []*Music
	for _, id := range musicIDs {
		music, err := mm.library.GetMusic(id)
		if err != nil {
			return fmt.Errorf("music not found for playlist: %s", id)
		}
		tracks = append(tracks, music)
	}

	mm.playlists[name] = tracks
	return nil
}

// PlayPlaylist starts playing a playlist
func (mm *MusicManager) PlayPlaylist(playlistName string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	playlist, exists := mm.playlists[playlistName]
	if !exists {
		return fmt.Errorf("playlist not found: %s", playlistName)
	}

	if len(playlist) == 0 {
		return fmt.Errorf("playlist is empty: %s", playlistName)
	}

	mm.currentPlaylist = playlistName
	return mm.playMusicInternal(playlist[0])
}

// playNextInPlaylist plays the next track in the current playlist
func (mm *MusicManager) playNextInPlaylist() {
	if mm.currentPlaylist == "" {
		return
	}

	playlist, exists := mm.playlists[mm.currentPlaylist]
	if !exists || len(playlist) == 0 {
		return
	}

	// Find current track index
	currentIndex := -1
	for i, track := range playlist {
		if track == mm.currentMusic {
			currentIndex = i
			break
		}
	}

	var nextTrack *Music
	if mm.loopMode == LoopShuffle {
		// Random track
		nextTrack = playlist[time.Now().UnixNano()%int64(len(playlist))]
	} else {
		// Next track in order
		nextIndex := (currentIndex + 1) % len(playlist)
		nextTrack = playlist[nextIndex]
	}

	mm.playMusicInternal(nextTrack)
}

// restartCurrentMusic restarts the current music track
func (mm *MusicManager) restartCurrentMusic() {
	if mm.currentMusic != nil {
		mm.playMusicInternal(mm.currentMusic)
	}
}

// IsPlaying returns whether music is currently playing
func (mm *MusicManager) IsPlaying() bool {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.isPlaying
}

// GetCurrentTrack returns the currently playing track
func (mm *MusicManager) GetCurrentTrack() *Music {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	return mm.currentMusic
}

// GetStats returns music manager statistics
func (mm *MusicManager) GetStats() MusicStats {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	stats := MusicStats{
		IsPlaying:         mm.isPlaying,
		IsPaused:          mm.isPaused,
		CurrentVolume:     mm.currentVolume,
		PlaybackPosition:  mm.playbackPosition,
		TotalPlayTime:     mm.totalPlayTime,
		CurrentMood:       mm.currentMood,
		CombatIntensity:   mm.combatIntensity,
		LoadedTracks:      len(mm.musicTracks),
		Playlists:         len(mm.playlists),
	}

	if mm.currentMusic != nil {
		stats.CurrentTrack = mm.currentMusic.Name
		stats.TrackDuration = mm.currentMusic.Duration
	}

	return stats
}

// MusicStats provides statistics about the music system
type MusicStats struct {
	IsPlaying         bool
	IsPaused          bool
	CurrentTrack      string
	CurrentVolume     float32
	PlaybackPosition  time.Duration
	TrackDuration     time.Duration
	TotalPlayTime     time.Duration
	CurrentMood       MusicMood
	CombatIntensity   float32
	LoadedTracks      int
	Playlists         int
}

// Shutdown cleans up the music manager
func (mm *MusicManager) Shutdown() error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	return mm.stopCurrentMusic()
}