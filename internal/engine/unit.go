package engine

import (
	"sync"
	"time"

	"teraglest/internal/data"
)

// UnitState represents the current state of a unit
type UnitState int

const (
	UnitStateIdle     UnitState = iota // Unit is idle
	UnitStateMoving                    // Unit is moving to target
	UnitStateAttacking                 // Unit is attacking
	UnitStateGathering                 // Unit is gathering resources
	UnitStateBuilding                  // Unit is constructing building
	UnitStateDead                      // Unit is dead
)

// String returns the string representation of a UnitState
func (us UnitState) String() string {
	switch us {
	case UnitStateIdle:
		return "Idle"
	case UnitStateMoving:
		return "Moving"
	case UnitStateAttacking:
		return "Attacking"
	case UnitStateGathering:
		return "Gathering"
	case UnitStateBuilding:
		return "Building"
	case UnitStateDead:
		return "Dead"
	default:
		return "Unknown"
	}
}

// GameUnit represents an enhanced unit with advanced lifecycle management
type GameUnit struct {
	// Base properties
	ID           int                 `json:"id"`
	PlayerID     int                 `json:"player_id"`
	UnitType     string              `json:"unit_type"`
	Name         string              `json:"name"`

	// State management
	Position     Vector3             `json:"position"`      // World coordinates (continuous)
	GridPos      GridPosition        `json:"grid_pos"`      // Grid coordinates + sub-tile offset
	Rotation     float32             `json:"rotation"`
	Health       int                 `json:"health"`
	MaxHealth    int                 `json:"max_health"`
	Armor        int                 `json:"armor"`
	Energy       int                 `json:"energy"`
	MaxEnergy    int                 `json:"max_energy"`

	// Lifecycle
	State        UnitState           `json:"state"`
	CreationTime time.Time           `json:"creation_time"`
	LastUpdate   time.Time           `json:"last_update"`

	// Command system
	CommandQueue []UnitCommand       `json:"command_queue"`
	CurrentCommand *UnitCommand      `json:"current_command"`

	// Movement and pathfinding
	Speed        float32             `json:"speed"`
	Target       *Vector3            `json:"target"`
	Path         []Vector3           `json:"path"`
	PathIndex    int                 `json:"path_index"`

	// Combat
	AttackDamage int                 `json:"attack_damage"`
	AttackRange  float32             `json:"attack_range"`
	AttackSpeed  float32             `json:"attack_speed"`
	LastAttack   time.Time           `json:"last_attack"`
	AttackTarget *GameUnit           `json:"attack_target"`

	// Resource gathering
	CarriedResources map[string]int   `json:"carried_resources"`
	GatherRate      map[string]float32 `json:"gather_rate"`
	GatherTarget    *ResourceNode     `json:"gather_target"`

	// Building construction
	BuildTarget     *GameBuilding     `json:"build_target"`
	BuildProgress   float32           `json:"build_progress"`

	// Unit definition data
	UnitDef      *data.UnitDefinition `json:"-"`

	// Threading
	mutex        sync.RWMutex         `json:"-"`
}

// GameObject Interface Implementation

func (u *GameUnit) GetID() int {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.ID
}

func (u *GameUnit) GetPlayerID() int {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.PlayerID
}

func (u *GameUnit) GetPosition() Vector3 {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.Position
}

func (u *GameUnit) SetPosition(pos Vector3) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.Position = pos
	u.LastUpdate = time.Now()
}

// Grid-aware positioning methods

func (u *GameUnit) GetGridPosition() GridPosition {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.GridPos
}

func (u *GameUnit) SetGridPosition(gridPos GridPosition, tileSize float32) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.GridPos = gridPos
	u.Position = GridToWorld(gridPos, tileSize)
	u.LastUpdate = time.Now()
}

func (u *GameUnit) UpdatePositions(worldPos Vector3, tileSize float32) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.Position = worldPos
	u.GridPos = WorldToGrid(worldPos, tileSize)
	u.LastUpdate = time.Now()
}

func (u *GameUnit) SetGridTarget(targetGrid GridPosition, tileSize float32) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.GridPos = targetGrid
	u.Position = GridToWorld(targetGrid, tileSize)

	// Update movement target
	worldTarget := GridToWorld(targetGrid, tileSize)
	u.Target = &worldTarget
	u.State = UnitStateMoving
	u.LastUpdate = time.Now()
}

func (u *GameUnit) GetHealth() int {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.Health
}

func (u *GameUnit) SetHealth(health int) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	u.Health = health
	if u.Health <= 0 {
		u.Health = 0
		u.State = UnitStateDead
	}
	u.LastUpdate = time.Now()
}

func (u *GameUnit) GetMaxHealth() int {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.MaxHealth
}

func (u *GameUnit) IsAlive() bool {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.Health > 0 && u.State != UnitStateDead
}

func (u *GameUnit) GetType() string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.UnitType
}

// Update handles all unit behavior updates
func (u *GameUnit) Update(deltaTime time.Duration) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	if u.State == UnitStateDead {
		return
	}

	// Update last update time
	u.LastUpdate = time.Now()

	// Process current command
	u.processCurrentCommand(deltaTime)

	// Update behaviors based on state
	switch u.State {
	case UnitStateMoving:
		u.updateMovement(deltaTime)
	case UnitStateAttacking:
		u.updateCombat(deltaTime)
	case UnitStateGathering:
		u.updateResourceGathering(deltaTime)
	case UnitStateBuilding:
		u.updateConstruction(deltaTime)
	}

	// Handle passive systems
	u.regenerateHealth(deltaTime)
	u.processCommandQueue()
}

// Internal update methods

func (u *GameUnit) processCurrentCommand(deltaTime time.Duration) {
	// Command processing is now handled centrally by CommandProcessor.Update()
	// which is called from World.Update(). This method is kept for interface
	// compatibility but actual command logic is in CommandProcessor.ProcessCommand()
	//
	// The CommandProcessor.Update() method iterates through all units and calls
	// CommandProcessor.ProcessCommand() for units with CurrentCommand != nil
}

func (u *GameUnit) updateMovement(deltaTime time.Duration) {
	if u.Target == nil || u.State != UnitStateMoving {
		return
	}

	// Calculate movement towards target
	dx := u.Target.X - u.Position.X
	dz := u.Target.Z - u.Position.Z
	distance := (dx*dx + dz*dz)

	if distance < 0.1 { // Close enough to target
		u.Position = *u.Target
		u.Target = nil
		u.State = UnitStateIdle
		return
	}

	// Move towards target
	distance = distance * 0.5 // sqrt approximation
	moveDistance := u.Speed * float32(deltaTime.Seconds())

	if distance > 0 {
		u.Position.X += dx * float64(moveDistance) / distance
		u.Position.Z += dz * float64(moveDistance) / distance
	}
}

func (u *GameUnit) updateCombat(deltaTime time.Duration) {
	if u.AttackTarget == nil || !u.AttackTarget.IsAlive() {
		u.State = UnitStateIdle
		u.AttackTarget = nil
		return
	}

	// Check if target is in range
	targetPos := u.AttackTarget.GetPosition()
	dx := targetPos.X - u.Position.X
	dz := targetPos.Z - u.Position.Z
	distance := dx*dx + dz*dz

	if distance > float64(u.AttackRange*u.AttackRange) {
		// Move closer to target
		u.Target = &targetPos
		u.State = UnitStateMoving
		return
	}

	// Attack if enough time has passed
	timeSinceLastAttack := time.Since(u.LastAttack)
	attackCooldown := time.Duration(1.0/u.AttackSpeed) * time.Second

	if timeSinceLastAttack >= attackCooldown {
		// Perform attack
		damage := u.AttackDamage - u.AttackTarget.Armor
		if damage < 1 {
			damage = 1
		}

		newHealth := u.AttackTarget.GetHealth() - damage
		u.AttackTarget.SetHealth(newHealth)
		u.LastAttack = time.Now()

		if !u.AttackTarget.IsAlive() {
			u.State = UnitStateIdle
			u.AttackTarget = nil
		}
	}
}

func (u *GameUnit) updateResourceGathering(deltaTime time.Duration) {
	if u.GatherTarget == nil || u.GatherTarget.Amount <= 0 {
		u.State = UnitStateIdle
		u.GatherTarget = nil
		return
	}

	// Gather resources
	resourceType := u.GatherTarget.ResourceType
	gatherRate, exists := u.GatherRate[resourceType]
	if !exists {
		u.State = UnitStateIdle
		u.GatherTarget = nil
		return
	}

	gathered := int(gatherRate * float32(deltaTime.Seconds()))
	if gathered > u.GatherTarget.Amount {
		gathered = u.GatherTarget.Amount
	}

	// Update carried resources
	if u.CarriedResources == nil {
		u.CarriedResources = make(map[string]int)
	}
	u.CarriedResources[resourceType] += gathered
	u.GatherTarget.Amount -= gathered

	// Check capacity limits (simplified)
	totalCarried := 0
	for _, amount := range u.CarriedResources {
		totalCarried += amount
	}

	if totalCarried >= 100 || u.GatherTarget.Amount <= 0 {
		// Need to return to storage
		u.State = UnitStateIdle
		u.GatherTarget = nil
	}
}

func (u *GameUnit) updateConstruction(deltaTime time.Duration) {
	if u.BuildTarget == nil || u.BuildTarget.IsBuilt {
		u.State = UnitStateIdle
		u.BuildTarget = nil
		u.BuildProgress = 0.0
		return
	}

	// Progress construction (simplified)
	buildRate := 0.1 // 10% per second
	u.BuildProgress += float32(buildRate * deltaTime.Seconds())

	if u.BuildProgress >= 1.0 {
		u.BuildProgress = 1.0
		u.BuildTarget.IsBuilt = true
		u.BuildTarget.CompletionTime = time.Now()
		u.State = UnitStateIdle
		u.BuildTarget = nil
		u.BuildProgress = 0.0
	}
}

func (u *GameUnit) regenerateHealth(deltaTime time.Duration) {
	if u.Health < u.MaxHealth && u.State != UnitStateDead && u.UnitDef != nil {
		// Regenerate health based on unit definition
		regen := float32(u.UnitDef.Unit.Parameters.MaxHP.Regeneration) * float32(deltaTime.Seconds())
		u.Health += int(regen)
		if u.Health > u.MaxHealth {
			u.Health = u.MaxHealth
		}
	}
}

func (u *GameUnit) processCommandQueue() {
	if u.CurrentCommand == nil && len(u.CommandQueue) > 0 {
		// Start next command
		u.CurrentCommand = &u.CommandQueue[0]
		u.CommandQueue = u.CommandQueue[1:]
	}
}