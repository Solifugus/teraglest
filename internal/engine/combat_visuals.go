package engine

import (
	"math"
	"time"
)

// CombatVisualSystem handles visual feedback for combat actions
type CombatVisualSystem struct {
	world           *World
	activeEffects   []VisualEffect
	damageNumbers   []DamageNumber
	projectiles     []CombatProjectile
	explosions      []ExplosionEffect
	statusIndicators []StatusIndicator
}

// NewCombatVisualSystem creates a new combat visual system
func NewCombatVisualSystem(world *World) *CombatVisualSystem {
	return &CombatVisualSystem{
		world:           world,
		activeEffects:   make([]VisualEffect, 0),
		damageNumbers:   make([]DamageNumber, 0),
		projectiles:     make([]CombatProjectile, 0),
		explosions:      make([]ExplosionEffect, 0),
		statusIndicators: make([]StatusIndicator, 0),
	}
}

// VisualEffect represents a temporary visual effect
type VisualEffect struct {
	ID           int                    `json:"id"`
	Type         VisualEffectType       `json:"type"`
	Position     Vector3                `json:"position"`
	StartTime    time.Time              `json:"start_time"`
	Duration     time.Duration          `json:"duration"`
	Scale        float64                `json:"scale"`
	Color        Color                  `json:"color"`
	Rotation     float64                `json:"rotation"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// VisualEffectType represents different types of visual effects
type VisualEffectType int

const (
	EffectSparks      VisualEffectType = iota // Melee hit sparks
	EffectBlood                              // Blood splatter
	EffectMuzzleFlash                        // Ranged weapon flash
	EffectImpact                             // Projectile impact
	EffectExplosion                          // Area damage explosion
	EffectHealGlow                           // Healing visual
	EffectMagicCircle                        // Magic effect circle
	EffectSmoke                              // Smoke effect
	EffectDebris                             // Debris particles
)

// DamageNumber represents floating damage numbers
type DamageNumber struct {
	ID         int         `json:"id"`
	Text       string      `json:"text"`
	Position   Vector3     `json:"position"`
	StartTime  time.Time   `json:"start_time"`
	Duration   time.Duration `json:"duration"`
	Color      Color       `json:"color"`
	FontSize   float64     `json:"font_size"`
	Velocity   Vector3     `json:"velocity"`
	IsCritical bool        `json:"is_critical"`
	IsHeal     bool        `json:"is_heal"`
}

// CombatProjectile represents visual projectiles
type CombatProjectile struct {
	ID           int       `json:"id"`
	Type         string    `json:"type"`
	StartPos     Vector3   `json:"start_pos"`
	EndPos       Vector3   `json:"end_pos"`
	CurrentPos   Vector3   `json:"current_pos"`
	Speed        float64   `json:"speed"`
	StartTime    time.Time `json:"start_time"`
	FlightTime   time.Duration `json:"flight_time"`
	Scale        float64   `json:"scale"`
	Rotation     float64   `json:"rotation"`
	TrailLength  float64   `json:"trail_length"`
}

// ExplosionEffect represents explosion visual effects
type ExplosionEffect struct {
	ID          int       `json:"id"`
	Position    Vector3   `json:"position"`
	StartTime   time.Time `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	MaxRadius   float64   `json:"max_radius"`
	CurrentRadius float64 `json:"current_radius"`
	Intensity   float64   `json:"intensity"`
	Color       Color     `json:"color"`
	ShakeAmount float64   `json:"shake_amount"` // Camera shake intensity
}

// StatusIndicator represents visual indicators for status effects
type StatusIndicator struct {
	ID        int         `json:"id"`
	UnitID    int         `json:"unit_id"`
	EffectID  string      `json:"effect_id"`
	IconPath  string      `json:"icon_path"`
	Position  Vector3     `json:"position"`
	StartTime time.Time   `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Scale     float64     `json:"scale"`
	Opacity   float64     `json:"opacity"`
}

// Color represents RGBA color values
type Color struct {
	R, G, B, A float64 `json:"r,g,b,a"`
}

// Predefined colors for different effect types
var (
	ColorPhysicalDamage = Color{1.0, 0.8, 0.6, 1.0} // Orange-ish
	ColorMagicalDamage  = Color{0.6, 0.8, 1.0, 1.0} // Blue-ish
	ColorCriticalHit    = Color{1.0, 0.2, 0.2, 1.0} // Red
	ColorHealing        = Color{0.2, 1.0, 0.2, 1.0} // Green
	ColorPoisonDamage   = Color{0.6, 1.0, 0.2, 1.0} // Green-yellow
	ColorFireDamage     = Color{1.0, 0.4, 0.1, 1.0} // Orange-red
	ColorIceDamage      = Color{0.7, 0.9, 1.0, 1.0} // Light blue
)

// CreateMeleeHitEffect creates visual effects for melee combat
func (cvs *CombatVisualSystem) CreateMeleeHitEffect(attackPos, targetPos Vector3, damageType string, damage int, isCritical bool) {
	// Create sparks effect at impact point
	sparksEffect := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectSparks,
		Position:  targetPos,
		StartTime: time.Now(),
		Duration:  time.Millisecond * 300,
		Scale:     1.0,
		Color:     cvs.getColorForDamageType(damageType),
		Rotation:  cvs.calculateImpactAngle(attackPos, targetPos),
	}
	cvs.activeEffects = append(cvs.activeEffects, sparksEffect)

	// Create blood splatter (if appropriate)
	if cvs.shouldShowBlood(damageType) {
		bloodEffect := VisualEffect{
			ID:        cvs.generateEffectID(),
			Type:      EffectBlood,
			Position:  targetPos,
			StartTime: time.Now(),
			Duration:  time.Second * 2,
			Scale:     math.Min(1.5, float64(damage)/20.0), // Scale with damage
			Color:     Color{0.8, 0.1, 0.1, 0.8},
		}
		cvs.activeEffects = append(cvs.activeEffects, bloodEffect)
	}

	// Create floating damage number
	cvs.CreateDamageNumber(targetPos, damage, damageType, isCritical, false)
}

// CreateRangedAttackEffect creates visual effects for ranged attacks
func (cvs *CombatVisualSystem) CreateRangedAttackEffect(attackPos, targetPos Vector3, damageType string, projectileType string) {
	// Create muzzle flash at attacker position
	muzzleFlash := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectMuzzleFlash,
		Position:  attackPos,
		StartTime: time.Now(),
		Duration:  time.Millisecond * 150,
		Scale:     0.8,
		Color:     cvs.getColorForDamageType(damageType),
	}
	cvs.activeEffects = append(cvs.activeEffects, muzzleFlash)

	// Create projectile
	distance := cvs.world.CalculateDistance(attackPos, targetPos)
	speed := cvs.getProjectileSpeed(projectileType)
	flightTime := time.Duration(distance/speed) * time.Second

	projectile := CombatProjectile{
		ID:          cvs.generateEffectID(),
		Type:        projectileType,
		StartPos:    attackPos,
		EndPos:      targetPos,
		CurrentPos:  attackPos,
		Speed:       speed,
		StartTime:   time.Now(),
		FlightTime:  flightTime,
		Scale:       1.0,
		TrailLength: cvs.getTrailLength(projectileType),
	}
	cvs.projectiles = append(cvs.projectiles, projectile)
}

// CreateSplashDamageEffect creates visual effects for area of effect attacks
func (cvs *CombatVisualSystem) CreateSplashDamageEffect(center Vector3, radius float64, damageType string, victims []SplashVictim) {
	// Create main explosion effect
	explosion := ExplosionEffect{
		ID:          cvs.generateEffectID(),
		Position:    center,
		StartTime:   time.Now(),
		Duration:    time.Millisecond * 800,
		MaxRadius:   radius,
		CurrentRadius: 0.0,
		Intensity:   1.0,
		Color:       cvs.getColorForDamageType(damageType),
		ShakeAmount: math.Min(1.0, radius/5.0), // Camera shake based on explosion size
	}
	cvs.explosions = append(cvs.explosions, explosion)

	// Create shockwave effect
	shockwave := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectExplosion,
		Position:  center,
		StartTime: time.Now(),
		Duration:  time.Millisecond * 600,
		Scale:     radius,
		Color:     cvs.getColorForDamageType(damageType),
		Parameters: map[string]interface{}{
			"is_shockwave": true,
			"max_radius":   radius,
		},
	}
	cvs.activeEffects = append(cvs.activeEffects, shockwave)

	// Create damage numbers for each victim
	for _, victim := range victims {
		cvs.CreateDamageNumber(victim.Unit.Position, victim.Damage, damageType, false, false)
	}

	// Create debris effects around the explosion
	cvs.createDebrisEffect(center, radius)
}

// CreateStatusEffectVisual creates visual indicators for status effects
func (cvs *CombatVisualSystem) CreateStatusEffectVisual(unit *GameUnit, effectID string, duration time.Duration) {
	effect, exists := StatusEffects[effectID]
	if !exists {
		return
	}

	indicator := StatusIndicator{
		ID:        cvs.generateEffectID(),
		UnitID:    unit.ID,
		EffectID:  effectID,
		IconPath:  effect.IconPath,
		Position:  Vector3{X: unit.Position.X, Y: unit.Position.Y + 2.0, Z: unit.Position.Z}, // Above unit
		StartTime: time.Now(),
		Duration:  duration,
		Scale:     1.0,
		Opacity:   0.8,
	}
	cvs.statusIndicators = append(cvs.statusIndicators, indicator)

	// Create additional visual effects for certain status types
	switch effect.Type {
	case EffectPoison:
		cvs.createPoisonCloudEffect(unit.Position)
	case EffectBurn:
		cvs.createFireEffect(unit.Position)
	case EffectFreeze:
		cvs.createIceEffect(unit.Position)
	case EffectStun:
		cvs.createStunStarsEffect(unit.Position)
	}
}

// CreateDamageNumber creates floating damage/healing numbers
func (cvs *CombatVisualSystem) CreateDamageNumber(position Vector3, amount int, damageType string, isCritical, isHeal bool) {
	color := cvs.getColorForDamageType(damageType)
	if isHeal {
		color = ColorHealing
	} else if isCritical {
		color = ColorCriticalHit
	}

	fontSize := 1.0
	if isCritical {
		fontSize = 1.5
	}
	if isHeal {
		fontSize = 1.2
	}

	text := ""
	if isHeal {
		text = "+" + string(rune(amount))
	} else {
		text = string(rune(amount))
	}

	damageNumber := DamageNumber{
		ID:         cvs.generateEffectID(),
		Text:       text,
		Position:   Vector3{X: position.X, Y: position.Y + 1.0, Z: position.Z},
		StartTime:  time.Now(),
		Duration:   time.Millisecond * 1500,
		Color:      color,
		FontSize:   fontSize,
		Velocity:   Vector3{X: 0, Y: 2.0, Z: 0}, // Float upward
		IsCritical: isCritical,
		IsHeal:     isHeal,
	}

	cvs.damageNumbers = append(cvs.damageNumbers, damageNumber)
}

// Update updates all visual effects
func (cvs *CombatVisualSystem) Update(deltaTime time.Duration) {
	now := time.Now()

	// Update and clean up visual effects
	cvs.activeEffects = cvs.updateVisualEffects(cvs.activeEffects, now)
	cvs.damageNumbers = cvs.updateDamageNumbers(cvs.damageNumbers, now, deltaTime)
	cvs.projectiles = cvs.updateProjectiles(cvs.projectiles, now)
	cvs.explosions = cvs.updateExplosions(cvs.explosions, now, deltaTime)
	cvs.statusIndicators = cvs.updateStatusIndicators(cvs.statusIndicators, now)
}

// updateVisualEffects updates visual effects and removes expired ones
func (cvs *CombatVisualSystem) updateVisualEffects(effects []VisualEffect, now time.Time) []VisualEffect {
	active := make([]VisualEffect, 0, len(effects))

	for _, effect := range effects {
		if now.Sub(effect.StartTime) < effect.Duration {
			active = append(active, effect)
		}
	}

	return active
}

// updateDamageNumbers updates floating damage numbers
func (cvs *CombatVisualSystem) updateDamageNumbers(numbers []DamageNumber, now time.Time, deltaTime time.Duration) []DamageNumber {
	active := make([]DamageNumber, 0, len(numbers))

	for _, number := range numbers {
		if now.Sub(number.StartTime) < number.Duration {
			// Update position
			number.Position.X += number.Velocity.X * deltaTime.Seconds()
			number.Position.Y += number.Velocity.Y * deltaTime.Seconds()
			number.Position.Z += number.Velocity.Z * deltaTime.Seconds()

			// Apply gravity/deceleration
			number.Velocity.Y *= 0.98 // Slow down upward movement

			active = append(active, number)
		}
	}

	return active
}

// updateProjectiles updates projectile positions and handles impacts
func (cvs *CombatVisualSystem) updateProjectiles(projectiles []CombatProjectile, now time.Time) []CombatProjectile {
	active := make([]CombatProjectile, 0, len(projectiles))

	for _, proj := range projectiles {
		elapsed := now.Sub(proj.StartTime)
		if elapsed < proj.FlightTime {
			// Update projectile position
			progress := elapsed.Seconds() / proj.FlightTime.Seconds()
			proj.CurrentPos = cvs.interpolatePosition(proj.StartPos, proj.EndPos, progress)

			// Update rotation to face movement direction
			proj.Rotation = cvs.calculateMovementAngle(proj.StartPos, proj.EndPos)

			active = append(active, proj)
		} else {
			// Projectile reached target, create impact effect
			cvs.createProjectileImpactEffect(proj)
		}
	}

	return active
}

// updateExplosions updates explosion effects
func (cvs *CombatVisualSystem) updateExplosions(explosions []ExplosionEffect, now time.Time, deltaTime time.Duration) []ExplosionEffect {
	active := make([]ExplosionEffect, 0, len(explosions))

	for _, explosion := range explosions {
		elapsed := now.Sub(explosion.StartTime)
		if elapsed < explosion.Duration {
			// Update explosion radius
			progress := elapsed.Seconds() / explosion.Duration.Seconds()
			explosion.CurrentRadius = explosion.MaxRadius * progress

			// Fade intensity
			explosion.Intensity = 1.0 - progress

			active = append(active, explosion)
		}
	}

	return active
}

// updateStatusIndicators updates status effect indicators
func (cvs *CombatVisualSystem) updateStatusIndicators(indicators []StatusIndicator, now time.Time) []StatusIndicator {
	active := make([]StatusIndicator, 0, len(indicators))

	for _, indicator := range indicators {
		if now.Sub(indicator.StartTime) < indicator.Duration {
			// Update position to follow unit
			if unit := cvs.getUnitByID(indicator.UnitID); unit != nil {
				indicator.Position = Vector3{
					X: unit.Position.X,
					Y: unit.Position.Y + 2.0,
					Z: unit.Position.Z,
				}
			}

			// Fade out near end of duration
			timeLeft := indicator.Duration - now.Sub(indicator.StartTime)
			if timeLeft < time.Second {
				indicator.Opacity = timeLeft.Seconds()
			}

			active = append(active, indicator)
		}
	}

	return active
}

// Helper functions

func (cvs *CombatVisualSystem) generateEffectID() int {
	// Simple ID generation - could be improved with proper UUID
	return int(time.Now().UnixNano() % 1000000)
}

func (cvs *CombatVisualSystem) getColorForDamageType(damageType string) Color {
	switch damageType {
	case "sword", "axe", "hammer", "pierce":
		return ColorPhysicalDamage
	case "fireball", "burn":
		return ColorFireDamage
	case "lightning", "magic":
		return ColorMagicalDamage
	case "ice", "freeze":
		return ColorIceDamage
	case "poison":
		return ColorPoisonDamage
	default:
		return ColorPhysicalDamage
	}
}

func (cvs *CombatVisualSystem) shouldShowBlood(damageType string) bool {
	// Don't show blood for magic/elemental damage
	magicTypes := map[string]bool{
		"fireball": true, "lightning": true, "ice": true,
		"magic": true, "explosion": true,
	}
	return !magicTypes[damageType]
}

func (cvs *CombatVisualSystem) calculateImpactAngle(attackPos, targetPos Vector3) float64 {
	return math.Atan2(targetPos.Z-attackPos.Z, targetPos.X-attackPos.X)
}

func (cvs *CombatVisualSystem) getProjectileSpeed(projectileType string) float64 {
	speeds := map[string]float64{
		"arrow":     20.0,
		"bolt":      25.0,
		"fireball":  15.0,
		"lightning": 50.0,
		"stone":     12.0,
	}

	if speed, exists := speeds[projectileType]; exists {
		return speed
	}
	return 20.0 // Default speed
}

func (cvs *CombatVisualSystem) getTrailLength(projectileType string) float64 {
	trails := map[string]float64{
		"arrow":     0.5,
		"fireball":  1.5,
		"lightning": 2.0,
		"bolt":      0.3,
	}

	if trail, exists := trails[projectileType]; exists {
		return trail
	}
	return 0.5
}

func (cvs *CombatVisualSystem) interpolatePosition(start, end Vector3, progress float64) Vector3 {
	return Vector3{
		X: start.X + (end.X-start.X)*progress,
		Y: start.Y + (end.Y-start.Y)*progress,
		Z: start.Z + (end.Z-start.Z)*progress,
	}
}

func (cvs *CombatVisualSystem) calculateMovementAngle(start, end Vector3) float64 {
	return math.Atan2(end.Z-start.Z, end.X-start.X)
}

func (cvs *CombatVisualSystem) createProjectileImpactEffect(proj CombatProjectile) {
	impact := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectImpact,
		Position:  proj.EndPos,
		StartTime: time.Now(),
		Duration:  time.Millisecond * 200,
		Scale:     0.8,
		Color:     cvs.getColorForProjectileType(proj.Type),
	}
	cvs.activeEffects = append(cvs.activeEffects, impact)
}

func (cvs *CombatVisualSystem) createDebrisEffect(center Vector3, radius float64) {
	debris := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectDebris,
		Position:  center,
		StartTime: time.Now(),
		Duration:  time.Second * 3,
		Scale:     radius * 0.8,
		Color:     Color{0.6, 0.4, 0.2, 0.8}, // Brown debris
	}
	cvs.activeEffects = append(cvs.activeEffects, debris)
}

func (cvs *CombatVisualSystem) createPoisonCloudEffect(position Vector3) {
	cloud := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectSmoke,
		Position:  position,
		StartTime: time.Now(),
		Duration:  time.Second * 2,
		Scale:     1.5,
		Color:     ColorPoisonDamage,
	}
	cvs.activeEffects = append(cvs.activeEffects, cloud)
}

func (cvs *CombatVisualSystem) createFireEffect(position Vector3) {
	fire := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectMagicCircle,
		Position:  position,
		StartTime: time.Now(),
		Duration:  time.Second * 2,
		Scale:     1.0,
		Color:     ColorFireDamage,
	}
	cvs.activeEffects = append(cvs.activeEffects, fire)
}

func (cvs *CombatVisualSystem) createIceEffect(position Vector3) {
	ice := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectMagicCircle,
		Position:  position,
		StartTime: time.Now(),
		Duration:  time.Second * 3,
		Scale:     1.2,
		Color:     ColorIceDamage,
	}
	cvs.activeEffects = append(cvs.activeEffects, ice)
}

func (cvs *CombatVisualSystem) createStunStarsEffect(position Vector3) {
	stars := VisualEffect{
		ID:        cvs.generateEffectID(),
		Type:      EffectSparks,
		Position:  Vector3{X: position.X, Y: position.Y + 2.5, Z: position.Z},
		StartTime: time.Now(),
		Duration:  time.Second * 2,
		Scale:     0.8,
		Color:     Color{1.0, 1.0, 0.0, 1.0}, // Yellow stars
	}
	cvs.activeEffects = append(cvs.activeEffects, stars)
}

func (cvs *CombatVisualSystem) getColorForProjectileType(projectileType string) Color {
	switch projectileType {
	case "fireball":
		return ColorFireDamage
	case "lightning":
		return ColorMagicalDamage
	case "arrow", "bolt":
		return ColorPhysicalDamage
	default:
		return ColorPhysicalDamage
	}
}

func (cvs *CombatVisualSystem) getUnitByID(unitID int) *GameUnit {
	// Placeholder - would need access to object manager
	return nil
}

// GetActiveVisualEffects returns all currently active visual effects for rendering
func (cvs *CombatVisualSystem) GetActiveVisualEffects() []VisualEffect {
	return cvs.activeEffects
}

// GetActiveDamageNumbers returns all active damage numbers for UI rendering
func (cvs *CombatVisualSystem) GetActiveDamageNumbers() []DamageNumber {
	return cvs.damageNumbers
}

// GetActiveProjectiles returns all active projectiles for rendering
func (cvs *CombatVisualSystem) GetActiveProjectiles() []CombatProjectile {
	return cvs.projectiles
}

// GetActiveExplosions returns all active explosions for rendering
func (cvs *CombatVisualSystem) GetActiveExplosions() []ExplosionEffect {
	return cvs.explosions
}

// GetActiveStatusIndicators returns all active status indicators for UI rendering
func (cvs *CombatVisualSystem) GetActiveStatusIndicators() []StatusIndicator {
	return cvs.statusIndicators
}