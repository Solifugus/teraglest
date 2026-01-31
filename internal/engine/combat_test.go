package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

func TestCombatSystem_DamageCalculation(t *testing.T) {
	// Create test world with tech tree
	world := createTestCombatWorld(t)
	combat := NewCombatSystem(world)

	// Create test units
	attacker := createTestAttacker(1)
	target := createTestTarget(2)

	// Test damage calculation
	result := combat.CalculateDamage(attacker, target)

	if !result.CanAttack {
		t.Fatalf("Attack should be possible, got error: %s", result.ErrorMessage)
	}

	if result.Damage <= 0 {
		t.Errorf("Expected positive damage, got %d", result.Damage)
	}

	if result.BaseDamage != attacker.AttackDamage {
		t.Errorf("Expected base damage %d, got %d", attacker.AttackDamage, result.BaseDamage)
	}

	if result.Multiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0, got %.2f", result.Multiplier)
	}
}

func TestCombatSystem_ApplyDamage(t *testing.T) {
	world := createTestCombatWorld(t)
	combat := NewCombatSystem(world)

	target := createTestTarget(1)
	initialHealth := target.Health

	// Apply damage
	wasKilled := combat.ApplyDamage(target, 30)

	if wasKilled {
		t.Error("Target should not be killed by 30 damage")
	}

	expectedHealth := initialHealth - 30
	if target.Health != expectedHealth {
		t.Errorf("Expected health %d, got %d", expectedHealth, target.Health)
	}

	// Apply lethal damage
	wasKilled = combat.ApplyDamage(target, target.Health+10)

	if !wasKilled {
		t.Error("Target should be killed by lethal damage")
	}

	if target.Health != 0 {
		t.Errorf("Expected health 0 after death, got %d", target.Health)
	}

	if target.State != UnitStateDead {
		t.Errorf("Expected state Dead, got %v", target.State)
	}
}

func TestCombatSystem_RangeChecking(t *testing.T) {
	world := createTestCombatWorld(t)
	combat := NewCombatSystem(world)

	attacker := createTestAttacker(1)
	target := createTestTarget(2)

	// Test in-range attack
	attacker.Position = Vector3{X: 0, Y: 0, Z: 0}
	target.Position = Vector3{X: 3, Y: 0, Z: 0} // Within range
	attacker.AttackRange = 5.0

	canAttack, reason := combat.CanAttack(attacker, target)
	if !canAttack {
		t.Errorf("Attack should be possible, got: %s", reason)
	}

	// Test out-of-range attack
	target.Position = Vector3{X: 10, Y: 0, Z: 0} // Out of range

	canAttack, reason = combat.CanAttack(attacker, target)
	if canAttack {
		t.Error("Attack should not be possible due to range")
	}

	if reason == "" {
		t.Error("Expected error message for out-of-range attack")
	}
}

func TestCombatSystem_AttackCooldown(t *testing.T) {
	world := createTestCombatWorld(t)
	combat := NewCombatSystem(world)

	attacker := createTestAttacker(1)
	attacker.AttackSpeed = 2.0 // 2 attacks per second = 0.5s cooldown

	// First attack should be allowed
	canAttack := combat.canAttackNow(attacker)
	if !canAttack {
		t.Error("First attack should be allowed")
	}

	// Set last attack to very recent
	attacker.LastAttack = time.Now()

	// Immediate second attack should be blocked
	canAttack = combat.canAttackNow(attacker)
	if canAttack {
		t.Error("Immediate second attack should be blocked by cooldown")
	}
}

func TestCombatSystem_ExecuteAttack(t *testing.T) {
	world := createTestCombatWorld(t)
	combat := NewCombatSystem(world)

	attacker := createTestAttacker(1)
	target := createTestTarget(2)

	initialHealth := target.Health

	// Execute attack
	result := combat.ExecuteAttack(attacker, target)

	if !result.CanAttack {
		t.Fatalf("Attack execution should succeed, got error: %s", result.ErrorMessage)
	}

	// Verify damage was applied
	if target.Health >= initialHealth {
		t.Errorf("Target health should be reduced, was %d, now %d", initialHealth, target.Health)
	}

	// Verify attacker's last attack time was updated
	if attacker.LastAttack.IsZero() {
		t.Error("Attacker's LastAttack time should be updated")
	}
}

func TestCombatSystem_LineOfSight(t *testing.T) {
	world := createTestCombatWorld(t)
	combat := NewCombatSystem(world)

	// Test clear line of sight
	from := Vector2i{X: 0, Y: 0}
	to := Vector2i{X: 5, Y: 0}

	hasLOS := combat.checkLineOfSight(from, to)
	if !hasLOS {
		t.Error("Clear horizontal line should have line of sight")
	}

	// Test diagonal line of sight
	to = Vector2i{X: 3, Y: 3}
	hasLOS = combat.checkLineOfSight(from, to)
	if !hasLOS {
		t.Error("Clear diagonal line should have line of sight")
	}
}

// Test helper functions

func createTestCombatWorld(t *testing.T) *World {
	// Create basic world with minimal tech tree
	techTree := &data.TechTree{
		AttackTypes: []data.AttackType{
			{Name: "normal"},
			{Name: "slash"},
		},
		ArmorTypes: []data.ArmorType{
			{Name: "organic"},
			{Name: "metal"},
		},
		DamageMultipliers: []data.DamageMultiplier{
			{Attack: "slash", Armor: "organic", Value: 1.5},
			{Attack: "normal", Armor: "organic", Value: 1.0},
		},
	}

	world := &World{
		players:   make(map[int]*Player),
		techTree:  techTree,
		Width:     64,
		Height:    64,
		tileSize:  1.0,
	}

	// Initialize basic grid
	world.initializeGrid()

	// Initialize ObjectManager
	world.ObjectManager = NewObjectManager(world)

	// Create test players
	world.players[1] = &Player{
		ID:       1,
		Name:     "Player 1",
		IsActive: true,
		Resources: make(map[string]int),
		ResourcesGathered: make(map[string]int),
		ResourcesSpent: make(map[string]int),
	}

	world.players[2] = &Player{
		ID:       2,
		Name:     "Player 2",
		IsActive: true,
		Resources: make(map[string]int),
		ResourcesGathered: make(map[string]int),
		ResourcesSpent: make(map[string]int),
	}

	return world
}

func createTestAttacker(playerID int) *GameUnit {
	return &GameUnit{
		ID:           100,
		PlayerID:     playerID,
		UnitType:     "warrior",
		Name:         "Test Warrior",
		Position:     Vector3{X: 0, Y: 0, Z: 0},
		Health:       100,
		MaxHealth:    100,
		AttackDamage: 25,
		AttackRange:  5.0,
		AttackSpeed:  1.0,
		Armor:        5,
		State:        UnitStateIdle,
		CreationTime: time.Now(),
		LastUpdate:   time.Now(),
		CommandQueue: make([]UnitCommand, 0),
	}
}

func createTestTarget(playerID int) *GameUnit {
	return &GameUnit{
		ID:           200,
		PlayerID:     playerID,
		UnitType:     "archer",
		Name:         "Test Archer",
		Position:     Vector3{X: 3, Y: 0, Z: 0},
		Health:       80,
		MaxHealth:    80,
		AttackDamage: 20,
		AttackRange:  8.0,
		AttackSpeed:  1.5,
		Armor:        2,
		State:        UnitStateIdle,
		CreationTime: time.Now(),
		LastUpdate:   time.Now(),
		CommandQueue: make([]UnitCommand, 0),
	}
}