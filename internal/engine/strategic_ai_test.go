package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// TestStrategicAICreation tests basic strategic AI creation and initialization
func TestStrategicAICreation(t *testing.T) {
	// Create test world
	world := createTestWorldForAI()

	// Test AI creation
	ai := NewStrategicAI(1, world, BalancedPersonality, DifficultyNormal)
	if ai == nil {
		t.Fatal("Failed to create strategic AI")
	}

	if ai.playerID != 1 {
		t.Errorf("Expected player ID 1, got %d", ai.playerID)
	}

	if ai.personality.Name != "Balanced" {
		t.Errorf("Expected Balanced personality, got %s", ai.personality.Name)
	}

	if ai.difficulty != DifficultyNormal {
		t.Errorf("Expected Normal difficulty, got %v", ai.difficulty)
	}
}

// TestStrategicAIPersonalities tests different AI personalities
func TestStrategicAIPersonalities(t *testing.T) {
	world := createTestWorldForAI()

	testCases := []struct {
		personality AIPersonality
		expectedName string
		minAggression float64
		maxAggression float64
	}{
		{ConservativePersonality, "Conservative", 0.0, 0.3},
		{AggressivePersonality, "Aggressive", 0.8, 1.0},
		{BalancedPersonality, "Balanced", 0.4, 0.7},
		{TechnologicalPersonality, "Technological", 0.0, 0.5},
		{ExpansionistPersonality, "Expansionist", 0.3, 0.8},
	}

	for _, tc := range testCases {
		ai := NewStrategicAI(1, world, tc.personality, DifficultyNormal)
		if ai.personality.Name != tc.expectedName {
			t.Errorf("Expected personality %s, got %s", tc.expectedName, ai.personality.Name)
		}

		aggression := ai.personality.AggressionLevel
		if aggression < tc.minAggression || aggression > tc.maxAggression {
			t.Errorf("Aggression level %f out of range [%f, %f] for %s",
				aggression, tc.minAggression, tc.maxAggression, tc.expectedName)
		}
	}
}

// TestStrategicAIStateAssessment tests strategy state evaluation
func TestStrategicAIStateAssessment(t *testing.T) {
	world := createTestWorldForAI()
	ai := NewStrategicAI(1, world, BalancedPersonality, DifficultyNormal)

	// Add test player with some units and resources
	world.AddPlayer(1, "Test AI", "tech", true)

	// Test initial state assessment through update
	ai.Update(time.Millisecond * 100)

	state := ai.GetStrategyState()

	// Basic state validation
	if state.EconomicStrength < 0 || state.EconomicStrength > 1 {
		t.Errorf("Economic strength %f out of valid range [0,1]", state.EconomicStrength)
	}

	if state.MilitaryStrength < 0 || state.MilitaryStrength > 1 {
		t.Errorf("Military strength %f out of valid range [0,1]", state.MilitaryStrength)
	}

	if state.ThreatLevel < 0 || state.ThreatLevel > 1 {
		t.Errorf("Threat level %f out of valid range [0,1]", state.ThreatLevel)
	}
}

// TestStrategicAIDecisionMaking tests the decision generation system
func TestStrategicAIDecisionMaking(t *testing.T) {
	world := createTestWorldForAI()
	ai := NewStrategicAI(1, world, BalancedPersonality, DifficultyNormal)

	// Add test player
	world.AddPlayer(1, "Test AI", "tech", true)

	// Test decision making through update cycle
	ai.Update(time.Millisecond * 100)

	// Check that decisions were made
	decisions := ai.GetRecentDecisions()
	if len(decisions) == 0 {
		t.Error("Expected at least one decision after update")
		return
	}

	decision := decisions[0]

	// Validate decision
	if decision.Type < DecisionExpand || decision.Type > DecisionUpgrade {
		t.Errorf("Decision type %v is out of valid range", decision.Type)
	}

	if decision.Priority < 0 || decision.Priority > 1 {
		t.Errorf("Decision priority %f out of valid range [0,1]", decision.Priority)
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		t.Errorf("Decision confidence %f out of valid range [0,1]", decision.Confidence)
	}

	if decision.Rationale == "" {
		t.Error("Decision should have rationale")
	}
}

// TestStrategicAIUpdate tests the AI update cycle
func TestStrategicAIUpdate(t *testing.T) {
	world := createTestWorldForAI()
	ai := NewStrategicAI(1, world, BalancedPersonality, DifficultyNormal)

	// Add test player
	world.AddPlayer(1, "Test AI", "tech", true)

	// Test update
	ai.Update(time.Millisecond * 100)

	// Check that decisions were made (update was successful)

	// Check that decisions were recorded
	decisions := ai.GetRecentDecisions()
	if len(decisions) == 0 {
		t.Error("Expected at least one decision after update")
	}
}

// TestDifficultyModifiers tests AI difficulty behavior
func TestDifficultyModifiers(t *testing.T) {
	world := createTestWorldForAI()

	difficulties := []AIDifficulty{DifficultyEasy, DifficultyNormal, DifficultyHard, DifficultyExpert}

	for _, difficulty := range difficulties {
		ai := NewStrategicAI(1, world, BalancedPersonality, difficulty)

		basePriority := 0.8
		modifiedPriority := ai.applyDifficultyModifier(basePriority)

		switch difficulty {
		case DifficultyEasy:
			if modifiedPriority >= basePriority {
				t.Errorf("Easy difficulty should reduce decision quality, got %f >= %f", modifiedPriority, basePriority)
			}
		case DifficultyExpert:
			if modifiedPriority <= basePriority {
				t.Errorf("Expert difficulty should enhance decision quality, got %f <= %f", modifiedPriority, basePriority)
			}
		}
	}
}

// TestStrategicAIManager tests the AI manager coordination
func TestStrategicAIManager(t *testing.T) {
	world := createTestWorldForAI()
	mgr := NewStrategicAIManager(world)

	if mgr == nil {
		t.Fatal("Failed to create strategic AI manager")
	}

	// Test adding AI player
	world.AddPlayer(1, "AI Player 1", "tech", true)
	err := mgr.InitializeAIPlayer(1, ConservativePersonality, DifficultyNormal)
	if err != nil {
		t.Errorf("Failed to initialize AI player: %v", err)
	}

	// Test AI player retrieval
	ai := mgr.GetAIPlayer(1)
	if ai == nil {
		t.Error("Failed to retrieve AI player")
	}

	// Test AI player count
	count := mgr.GetAIPlayerCount()
	if count != 1 {
		t.Errorf("Expected 1 AI player, got %d", count)
	}

	// Test update
	mgr.Update(time.Millisecond * 600) // Trigger update (rate is 500ms)

	// Test removing AI player
	mgr.RemoveAIPlayer(1)
	count = mgr.GetAIPlayerCount()
	if count != 0 {
		t.Errorf("Expected 0 AI players after removal, got %d", count)
	}
}

// TestWorldAIIntegration tests AI integration with World
func TestWorldAIIntegration(t *testing.T) {
	world := createTestWorldForAI()

	// Test AI player initialization
	err := world.AddPlayer(1, "AI Player", "tech", true)
	if err != nil {
		t.Errorf("Failed to add AI player: %v", err)
	}

	err = world.InitializeAIPlayer(1, "balanced", "normal")
	if err != nil {
		t.Errorf("Failed to initialize AI player: %v", err)
	}

	// Test world update with AI
	world.Update(time.Millisecond * 16)

	// Verify AI exists and is active
	player := world.GetPlayer(1)
	if player == nil || !player.IsAI {
		t.Error("AI player should exist and be marked as AI")
	}
}

// TestAICommandExecution tests that AI decisions result in actual commands
func TestAICommandExecution(t *testing.T) {
	world := createTestWorldForAI()

	// Initialize world properly
	err := world.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize world: %v", err)
	}

	// Add AI player
	err = world.AddPlayer(1, "Test AI", "tech", true)
	if err != nil {
		t.Fatalf("Failed to add AI player: %v", err)
	}

	ai := NewStrategicAI(1, world, AggressivePersonality, DifficultyNormal)

	// Test basic AI functionality without unit creation for now
	// (Unit creation requires faction data which is complex to set up in tests)

	// Test that AI can execute strategies without crashing
	ai.executeExpansionStrategy(make(map[string]interface{}))
	ai.executeScoutingStrategy(make(map[string]interface{}))

	// Test passed if no panic occurred
	t.Log("AI command execution completed without errors")
}

// Helper function to create a test world
func createTestWorldForAI() *World {
	techTree := &data.TechTree{}
	assetMgr := &data.AssetManager{}
	settings := GameSettings{
		MaxPlayers: 4,
		GameSpeed:  1.0,
	}

	world, err := NewWorld(settings, techTree, assetMgr)
	if err != nil {
		panic("Failed to create test world")
	}

	return world
}