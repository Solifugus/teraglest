package engine

import (
	"math"
	"sort"
	"sync"
	"time"
)

// FormationType defines different military formation patterns
type FormationType int

const (
	FormationLine     FormationType = iota // Units in a horizontal line
	FormationColumn                        // Units in a vertical column
	FormationWedge                         // V-shaped formation for attacks
	FormationBox                           // Square/rectangular formation
	FormationCircle                        // Circular defensive formation
	FormationScatter                       // Spread out formation
	FormationCustom                        // User-defined formation
)

// String returns the string representation of FormationType
func (ft FormationType) String() string {
	switch ft {
	case FormationLine:
		return "Line"
	case FormationColumn:
		return "Column"
	case FormationWedge:
		return "Wedge"
	case FormationBox:
		return "Box"
	case FormationCircle:
		return "Circle"
	case FormationScatter:
		return "Scatter"
	case FormationCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// FormationParameters defines spacing and behavior parameters for formations
type FormationParameters struct {
	UnitSpacing     float32 // Distance between units
	Spacing         float32 // Formation density
	MaxSpeed        float32 // Maximum movement speed while in formation
	BreakDistance   float32 // Distance at which formation breaks
	ReformDistance  float32 // Distance at which formation reforms
	TurnRadius      float32 // How tightly formation can turn
	CohesionFactor  float32 // How strongly units try to maintain formation
}

// DefaultFormationParameters returns standard formation parameters
func DefaultFormationParameters() FormationParameters {
	return FormationParameters{
		UnitSpacing:    2.0,
		Spacing:        1.5,
		MaxSpeed:       8.0,
		BreakDistance:  15.0,
		ReformDistance: 5.0,
		TurnRadius:     3.0,
		CohesionFactor: 0.8,
	}
}

// FormationPosition defines where a unit should be in a formation
type FormationPosition struct {
	RelativePos Vector3 // Position relative to formation center
	Rotation    float32 // Desired rotation for this position
	Priority    int     // Formation priority (0 = leader)
}

// UnitGroup represents a group of units that can move in formation
type UnitGroup struct {
	ID          int                    // Unique group ID
	PlayerID    int                    // Owner player ID
	Units       map[int]*GameUnit      // Units in this group (by unit ID)
	Formation   FormationType          // Current formation type
	Parameters  FormationParameters    // Formation behavior parameters
	Leader      *GameUnit              // Formation leader (usually first selected)
	CenterPos   Vector3                // Current center position of formation
	TargetPos   Vector3                // Target center position
	Direction   Vector3                // Formation facing direction

	// Formation state
	IsMoving    bool                   // Whether group is currently moving
	IsFormed    bool                   // Whether units are in formation positions
	LastUpdate  time.Time              // Last formation update time

	// Formation positions
	Positions   map[int]FormationPosition // Assigned positions for each unit

	mutex       sync.RWMutex           // Thread safety
}

// NewUnitGroup creates a new unit group with the specified units
func NewUnitGroup(id int, playerID int, units []*GameUnit, formationType FormationType) *UnitGroup {
	group := &UnitGroup{
		ID:         id,
		PlayerID:   playerID,
		Units:      make(map[int]*GameUnit),
		Formation:  formationType,
		Parameters: DefaultFormationParameters(),
		Positions:  make(map[int]FormationPosition),
		IsFormed:   false,
		LastUpdate: time.Now(),
	}

	// Add units to group
	for _, unit := range units {
		group.Units[unit.ID] = unit
	}

	// Set leader (first unit or prioritize command units)
	if len(units) > 0 {
		group.Leader = units[0]
		// TODO: Could prioritize commander units or heroes
	}

	// Calculate initial center position
	group.updateCenterPosition()

	// Generate formation positions
	group.generateFormationPositions()

	return group
}

// AddUnit adds a unit to the group
func (g *UnitGroup) AddUnit(unit *GameUnit) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.Units[unit.ID] = unit

	// If no leader set, make this the leader
	if g.Leader == nil {
		g.Leader = unit
	}

	// Regenerate formation positions
	g.generateFormationPositions()
}

// RemoveUnit removes a unit from the group
func (g *UnitGroup) RemoveUnit(unitID int) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	delete(g.Units, unitID)
	delete(g.Positions, unitID)

	// If leader was removed, assign new leader
	if g.Leader != nil && g.Leader.ID == unitID {
		g.assignNewLeader()
	}

	// Regenerate formation positions
	if len(g.Units) > 0 {
		g.generateFormationPositions()
	}
}

// GetUnitCount returns the number of units in the group
func (g *UnitGroup) GetUnitCount() int {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return len(g.Units)
}

// IsEmpty returns whether the group has no units
func (g *UnitGroup) IsEmpty() bool {
	return g.GetUnitCount() == 0
}

// SetFormation changes the formation type and regenerates positions
func (g *UnitGroup) SetFormation(formationType FormationType) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.Formation = formationType
	g.IsFormed = false
	g.generateFormationPositions()
}

// MoveToPosition commands the entire group to move to a target position
func (g *UnitGroup) MoveToPosition(target Vector3) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.TargetPos = target
	g.IsMoving = true
	g.IsFormed = false

	// Update formation direction based on movement
	if g.CenterPos.X != target.X || g.CenterPos.Z != target.Z {
		direction := Vector3{
			X: target.X - g.CenterPos.X,
			Y: 0,
			Z: target.Z - g.CenterPos.Z,
		}
		g.Direction = normalizeVector3(direction)
	}

	// Generate new formation positions at target
	g.generateFormationPositions()
}

// Update updates the formation state and unit positions
func (g *UnitGroup) Update(deltaTime time.Duration) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if len(g.Units) == 0 {
		return
	}

	g.LastUpdate = time.Now()

	// Update center position
	g.updateCenterPosition()

	// Check if formation should be maintained
	g.checkFormationCohesion()

	// Update formation if moving
	if g.IsMoving {
		g.updateFormationMovement()
	}
}

// updateCenterPosition calculates the current center of the group
func (g *UnitGroup) updateCenterPosition() {
	if len(g.Units) == 0 {
		return
	}

	var totalPos Vector3
	count := 0

	for _, unit := range g.Units {
		if unit.IsAlive() {
			totalPos.X += unit.Position.X
			totalPos.Y += unit.Position.Y
			totalPos.Z += unit.Position.Z
			count++
		}
	}

	if count > 0 {
		g.CenterPos = Vector3{
			X: totalPos.X / float64(count),
			Y: totalPos.Y / float64(count),
			Z: totalPos.Z / float64(count),
		}
	}
}

// generateFormationPositions creates formation positions for all units
func (g *UnitGroup) generateFormationPositions() {
	if len(g.Units) == 0 {
		return
	}

	// Clear existing positions
	g.Positions = make(map[int]FormationPosition)

	// Get sorted list of units for consistent positioning
	units := g.getSortedUnits()

	// Generate positions based on formation type
	switch g.Formation {
	case FormationLine:
		g.generateLineFormation(units)
	case FormationColumn:
		g.generateColumnFormation(units)
	case FormationWedge:
		g.generateWedgeFormation(units)
	case FormationBox:
		g.generateBoxFormation(units)
	case FormationCircle:
		g.generateCircleFormation(units)
	case FormationScatter:
		g.generateScatterFormation(units)
	default:
		g.generateLineFormation(units) // Default to line
	}
}

// getSortedUnits returns units sorted by ID for consistent formation positions
func (g *UnitGroup) getSortedUnits() []*GameUnit {
	units := make([]*GameUnit, 0, len(g.Units))
	for _, unit := range g.Units {
		if unit.IsAlive() {
			units = append(units, unit)
		}
	}

	// Sort by ID for consistency, leader first
	sort.Slice(units, func(i, j int) bool {
		if g.Leader != nil && units[i].ID == g.Leader.ID {
			return true
		}
		if g.Leader != nil && units[j].ID == g.Leader.ID {
			return false
		}
		return units[i].ID < units[j].ID
	})

	return units
}

// generateLineFormation creates a horizontal line formation
func (g *UnitGroup) generateLineFormation(units []*GameUnit) {
	spacing := g.Parameters.UnitSpacing
	unitCount := len(units)
	totalWidth := float32(unitCount-1) * spacing
	startOffset := -totalWidth / 2

	for i, unit := range units {
		position := FormationPosition{
			RelativePos: Vector3{
				X: float64(startOffset + float32(i)*spacing),
				Y: 0,
				Z: 0,
			},
			Rotation: 0,
			Priority: i,
		}
		g.Positions[unit.ID] = position
	}
}

// generateColumnFormation creates a vertical column formation
func (g *UnitGroup) generateColumnFormation(units []*GameUnit) {
	spacing := g.Parameters.UnitSpacing

	for i, unit := range units {
		position := FormationPosition{
			RelativePos: Vector3{
				X: 0,
				Y: 0,
				Z: float64(-float32(i) * spacing), // Negative Z to put leader in front
			},
			Rotation: 0,
			Priority: i,
		}
		g.Positions[unit.ID] = position
	}
}

// generateWedgeFormation creates a V-shaped wedge formation
func (g *UnitGroup) generateWedgeFormation(units []*GameUnit) {
	spacing := g.Parameters.UnitSpacing

	for i, unit := range units {
		var relPos Vector3
		if i == 0 {
			// Leader at the front center
			relPos = Vector3{X: 0, Y: 0, Z: 0}
		} else {
			// Alternate left and right sides
			side := 1
			if i%2 == 0 {
				side = -1
			}
			row := (i + 1) / 2

			relPos = Vector3{
				X: float64(float32(side) * float32(row) * spacing * 0.8),
				Y: 0,
				Z: float64(-float32(row) * spacing * 0.6),
			}
		}

		position := FormationPosition{
			RelativePos: relPos,
			Rotation:    0,
			Priority:    i,
		}
		g.Positions[unit.ID] = position
	}
}

// generateBoxFormation creates a rectangular box formation
func (g *UnitGroup) generateBoxFormation(units []*GameUnit) {
	spacing := g.Parameters.UnitSpacing
	unitCount := len(units)

	// Calculate box dimensions (roughly square)
	sideLength := int(math.Ceil(math.Sqrt(float64(unitCount))))

	for i, unit := range units {
		row := i / sideLength
		col := i % sideLength

		position := FormationPosition{
			RelativePos: Vector3{
				X: float64((float32(col) - float32(sideLength-1)/2) * spacing),
				Y: 0,
				Z: float64((float32(row) - float32(sideLength-1)/2) * spacing),
			},
			Rotation: 0,
			Priority: i,
		}
		g.Positions[unit.ID] = position
	}
}

// generateCircleFormation creates a circular defensive formation
func (g *UnitGroup) generateCircleFormation(units []*GameUnit) {
	unitCount := len(units)
	if unitCount == 1 {
		// Single unit at center
		g.Positions[units[0].ID] = FormationPosition{
			RelativePos: Vector3{X: 0, Y: 0, Z: 0},
			Rotation:    0,
			Priority:    0,
		}
		return
	}

	radius := g.Parameters.UnitSpacing * float32(unitCount) / (2 * math.Pi)
	if radius < g.Parameters.UnitSpacing {
		radius = g.Parameters.UnitSpacing
	}

	for i, unit := range units {
		angle := 2 * math.Pi * float64(i) / float64(unitCount)

		position := FormationPosition{
			RelativePos: Vector3{
				X: float64(radius * float32(math.Cos(angle))),
				Y: 0,
				Z: float64(radius * float32(math.Sin(angle))),
			},
			Rotation: float32(angle + math.Pi/2), // Face outward
			Priority: i,
		}
		g.Positions[unit.ID] = position
	}
}

// generateScatterFormation creates a loose scattered formation
func (g *UnitGroup) generateScatterFormation(units []*GameUnit) {
	spacing := g.Parameters.UnitSpacing * 1.5 // Wider spacing

	for i, unit := range units {
		// Use pseudo-random positioning based on unit ID for consistency
		pseudoRand := float32(unit.ID*31+unit.ID*17) / 1000.0

		position := FormationPosition{
			RelativePos: Vector3{
				X: float64((pseudoRand - 0.5) * spacing * 3),
				Y: 0,
				Z: float64((float32(int(pseudoRand*100)%100) - 50) * spacing * 0.02),
			},
			Rotation: 0,
			Priority: i,
		}
		g.Positions[unit.ID] = position
	}
}

// checkFormationCohesion determines if formation should be maintained or broken
func (g *UnitGroup) checkFormationCohesion() {
	if !g.IsFormed {
		return
	}

	maxDistance := float32(0)

	for _, unit := range g.Units {
		if unit.IsAlive() {
			distance := distanceVector3(unit.Position, g.CenterPos)
			if distance > maxDistance {
				maxDistance = distance
			}
		}
	}

	// Break formation if units are too spread out
	if maxDistance > g.Parameters.BreakDistance {
		g.IsFormed = false
	}
}

// updateFormationMovement handles formation movement logic
func (g *UnitGroup) updateFormationMovement() {
	// Check if we've reached the target
	distance := distanceVector3(g.CenterPos, g.TargetPos)
	if distance < 2.0 {
		g.IsMoving = false
		g.IsFormed = true
		return
	}

	// Update formation center toward target
	// This would be used by pathfinding system
}

// assignNewLeader selects a new leader when the current one is removed
func (g *UnitGroup) assignNewLeader() {
	g.Leader = nil

	for _, unit := range g.Units {
		if unit.IsAlive() {
			g.Leader = unit
			break
		}
	}

	// Regenerate positions with new leader
	if g.Leader != nil {
		g.generateFormationPositions()
	}
}

// GetFormationPosition returns the world position where a unit should be
func (g *UnitGroup) GetFormationPosition(unitID int) (Vector3, bool) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	formPos, exists := g.Positions[unitID]
	if !exists {
		return Vector3{}, false
	}

	// Transform relative position to world coordinates
	worldPos := g.transformToWorldPosition(formPos.RelativePos)
	return worldPos, true
}

// transformToWorldPosition converts formation-relative position to world coordinates
func (g *UnitGroup) transformToWorldPosition(relativePos Vector3) Vector3 {
	// For now, use target position as formation center
	center := g.TargetPos
	if !g.IsMoving {
		center = g.CenterPos
	}

	// Apply rotation based on formation direction
	// For simplicity, assume direction is along Z-axis for now
	// TODO: Implement full rotation matrix

	return Vector3{
		X: center.X + relativePos.X,
		Y: center.Y + relativePos.Y,
		Z: center.Z + relativePos.Z,
	}
}

// Helper functions

// normalizeVector3 normalizes a vector to unit length
func normalizeVector3(v Vector3) Vector3 {
	length := float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
	if length < 0.0001 {
		return Vector3{X: 0, Y: 0, Z: 1} // Default forward direction
	}
	return Vector3{X: v.X / float64(length), Y: v.Y / float64(length), Z: v.Z / float64(length)}
}

// distanceVector3 calculates distance between two vectors
func distanceVector3(a, b Vector3) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}