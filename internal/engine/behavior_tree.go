package engine

import (
	"fmt"
	"time"
)

// NodeStatus represents the possible states of a behavior tree node execution
type NodeStatus int

const (
	StatusSuccess NodeStatus = iota // Node completed successfully
	StatusFailure                   // Node failed to complete
	StatusRunning                   // Node is still executing
	StatusInvalid                   // Node is in an invalid state
)

// String returns the string representation of NodeStatus
func (s NodeStatus) String() string {
	switch s {
	case StatusSuccess:
		return "Success"
	case StatusFailure:
		return "Failure"
	case StatusRunning:
		return "Running"
	case StatusInvalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

// BehaviorNode represents a single node in a behavior tree
type BehaviorNode interface {
	// Execute runs the behavior and returns its status
	Execute(context *BehaviorContext) NodeStatus

	// Reset resets the node's internal state
	Reset()

	// GetName returns a human-readable name for the node
	GetName() string

	// GetChildren returns child nodes (empty for leaf nodes)
	GetChildren() []BehaviorNode

	// AddChild adds a child node (only valid for composite nodes)
	AddChild(child BehaviorNode) error
}

// BehaviorContext contains the execution context for behavior tree nodes
type BehaviorContext struct {
	Unit      *GameUnit           // The unit executing the behavior
	World     *World              // World reference for game state
	Blackboard *Blackboard        // Shared data storage
	DeltaTime  time.Duration      // Time since last update
	StartTime  time.Time          // When the behavior started
	UserData   map[string]interface{} // Node-specific data storage
}

// NewBehaviorContext creates a new behavior context for a unit
func NewBehaviorContext(unit *GameUnit, world *World, deltaTime time.Duration) *BehaviorContext {
	return &BehaviorContext{
		Unit:      unit,
		World:     world,
		Blackboard: NewBlackboard(),
		DeltaTime: deltaTime,
		StartTime: time.Now(),
		UserData:  make(map[string]interface{}),
	}
}

// Blackboard provides shared memory for behavior tree execution
type Blackboard struct {
	data map[string]interface{}
}

// NewBlackboard creates a new blackboard instance
func NewBlackboard() *Blackboard {
	return &Blackboard{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the blackboard
func (bb *Blackboard) Set(key string, value interface{}) {
	bb.data[key] = value
}

// Get retrieves a value from the blackboard
func (bb *Blackboard) Get(key string) (interface{}, bool) {
	value, exists := bb.data[key]
	return value, exists
}

// GetString retrieves a string value from the blackboard
func (bb *Blackboard) GetString(key string) (string, bool) {
	if value, exists := bb.data[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetInt retrieves an integer value from the blackboard
func (bb *Blackboard) GetInt(key string) (int, bool) {
	if value, exists := bb.data[key]; exists {
		if i, ok := value.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// GetFloat retrieves a float value from the blackboard
func (bb *Blackboard) GetFloat(key string) (float64, bool) {
	if value, exists := bb.data[key]; exists {
		if f, ok := value.(float64); ok {
			return f, true
		}
	}
	return 0.0, false
}

// GetBool retrieves a boolean value from the blackboard
func (bb *Blackboard) GetBool(key string) (bool, bool) {
	if value, exists := bb.data[key]; exists {
		if b, ok := value.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// GetVector3 retrieves a Vector3 value from the blackboard
func (bb *Blackboard) GetVector3(key string) (Vector3, bool) {
	if value, exists := bb.data[key]; exists {
		if v, ok := value.(Vector3); ok {
			return v, true
		}
	}
	return Vector3{}, false
}

// GetUnit retrieves a GameUnit from the blackboard
func (bb *Blackboard) GetUnit(key string) (*GameUnit, bool) {
	if value, exists := bb.data[key]; exists {
		if unit, ok := value.(*GameUnit); ok {
			return unit, true
		}
	}
	return nil, false
}

// Has checks if a key exists in the blackboard
func (bb *Blackboard) Has(key string) bool {
	_, exists := bb.data[key]
	return exists
}

// Clear removes all data from the blackboard
func (bb *Blackboard) Clear() {
	bb.data = make(map[string]interface{})
}

// Remove removes a specific key from the blackboard
func (bb *Blackboard) Remove(key string) {
	delete(bb.data, key)
}

// GetKeys returns all keys in the blackboard
func (bb *Blackboard) GetKeys() []string {
	keys := make([]string, 0, len(bb.data))
	for key := range bb.data {
		keys = append(keys, key)
	}
	return keys
}

// BaseNode provides common functionality for behavior tree nodes
type BaseNode struct {
	name     string
	children []BehaviorNode
}

// GetName returns the node's name
func (bn *BaseNode) GetName() string {
	return bn.name
}

// GetChildren returns the node's children
func (bn *BaseNode) GetChildren() []BehaviorNode {
	return bn.children
}

// AddChild adds a child node (for composite nodes)
func (bn *BaseNode) AddChild(child BehaviorNode) error {
	bn.children = append(bn.children, child)
	return nil
}

// Reset resets the node's state (default implementation)
func (bn *BaseNode) Reset() {
	// Override in specific node types if needed
}

// CompositeNode base class for nodes that have children
type CompositeNode struct {
	BaseNode
	currentChild int // Index of currently executing child
}

// Reset resets the composite node's state
func (cn *CompositeNode) Reset() {
	cn.currentChild = 0
	for _, child := range cn.children {
		child.Reset()
	}
}

// SequenceNode executes children in order, succeeding only if all succeed
type SequenceNode struct {
	CompositeNode
}

// NewSequenceNode creates a new sequence node
func NewSequenceNode(name string) *SequenceNode {
	return &SequenceNode{
		CompositeNode: CompositeNode{
			BaseNode: BaseNode{
				name:     name,
				children: make([]BehaviorNode, 0),
			},
			currentChild: 0,
		},
	}
}

// Execute runs the sequence node logic
func (sn *SequenceNode) Execute(context *BehaviorContext) NodeStatus {
	for sn.currentChild < len(sn.children) {
		status := sn.children[sn.currentChild].Execute(context)

		switch status {
		case StatusSuccess:
			// Move to next child
			sn.currentChild++
			continue
		case StatusFailure:
			// Sequence fails if any child fails
			sn.Reset()
			return StatusFailure
		case StatusRunning:
			// Keep running current child
			return StatusRunning
		}
	}

	// All children succeeded
	sn.Reset()
	return StatusSuccess
}

// SelectorNode executes children in order, succeeding if any child succeeds
type SelectorNode struct {
	CompositeNode
}

// NewSelectorNode creates a new selector node
func NewSelectorNode(name string) *SelectorNode {
	return &SelectorNode{
		CompositeNode: CompositeNode{
			BaseNode: BaseNode{
				name:     name,
				children: make([]BehaviorNode, 0),
			},
			currentChild: 0,
		},
	}
}

// Execute runs the selector node logic
func (sn *SelectorNode) Execute(context *BehaviorContext) NodeStatus {
	for sn.currentChild < len(sn.children) {
		status := sn.children[sn.currentChild].Execute(context)

		switch status {
		case StatusSuccess:
			// Selector succeeds if any child succeeds
			sn.Reset()
			return StatusSuccess
		case StatusFailure:
			// Try next child
			sn.currentChild++
			continue
		case StatusRunning:
			// Keep running current child
			return StatusRunning
		}
	}

	// All children failed
	sn.Reset()
	return StatusFailure
}

// ParallelNode executes all children simultaneously
type ParallelNode struct {
	CompositeNode
	policy ParallelPolicy // How many children must succeed
}

// ParallelPolicy defines when a parallel node succeeds
type ParallelPolicy int

const (
	ParallelPolicyRequireOne ParallelPolicy = iota // Succeed if at least one child succeeds
	ParallelPolicyRequireAll                       // Succeed only if all children succeed
)

// NewParallelNode creates a new parallel node
func NewParallelNode(name string, policy ParallelPolicy) *ParallelNode {
	return &ParallelNode{
		CompositeNode: CompositeNode{
			BaseNode: BaseNode{
				name:     name,
				children: make([]BehaviorNode, 0),
			},
		},
		policy: policy,
	}
}

// Execute runs the parallel node logic
func (pn *ParallelNode) Execute(context *BehaviorContext) NodeStatus {
	if len(pn.children) == 0 {
		return StatusSuccess
	}

	var successCount, failureCount, runningCount int

	// Execute all children
	for _, child := range pn.children {
		status := child.Execute(context)
		switch status {
		case StatusSuccess:
			successCount++
		case StatusFailure:
			failureCount++
		case StatusRunning:
			runningCount++
		}
	}

	// Apply policy
	switch pn.policy {
	case ParallelPolicyRequireOne:
		if successCount > 0 {
			return StatusSuccess
		}
		if runningCount > 0 {
			return StatusRunning
		}
		return StatusFailure

	case ParallelPolicyRequireAll:
		if failureCount > 0 {
			return StatusFailure
		}
		if runningCount > 0 {
			return StatusRunning
		}
		return StatusSuccess
	}

	return StatusFailure
}

// DecoratorNode base class for nodes that modify child behavior
type DecoratorNode struct {
	BaseNode
	child BehaviorNode
}

// AddChild adds a single child to the decorator (replaces existing child)
func (dn *DecoratorNode) AddChild(child BehaviorNode) error {
	if dn.child != nil {
		return fmt.Errorf("decorator node %s already has a child", dn.name)
	}
	dn.child = child
	return nil
}

// GetChildren returns the decorator's child as a slice
func (dn *DecoratorNode) GetChildren() []BehaviorNode {
	if dn.child == nil {
		return []BehaviorNode{}
	}
	return []BehaviorNode{dn.child}
}

// Reset resets the decorator and its child
func (dn *DecoratorNode) Reset() {
	if dn.child != nil {
		dn.child.Reset()
	}
}

// InverterNode inverts the result of its child
type InverterNode struct {
	DecoratorNode
}

// NewInverterNode creates a new inverter node
func NewInverterNode(name string) *InverterNode {
	return &InverterNode{
		DecoratorNode: DecoratorNode{
			BaseNode: BaseNode{
				name: name,
			},
		},
	}
}

// Execute runs the inverter logic
func (in *InverterNode) Execute(context *BehaviorContext) NodeStatus {
	if in.child == nil {
		return StatusFailure
	}

	status := in.child.Execute(context)
	switch status {
	case StatusSuccess:
		return StatusFailure
	case StatusFailure:
		return StatusSuccess
	case StatusRunning:
		return StatusRunning
	default:
		return StatusInvalid
	}
}

// RepeaterNode repeats its child a specified number of times
type RepeaterNode struct {
	DecoratorNode
	maxRepeats    int // -1 for infinite repeats
	currentRepeat int
}

// NewRepeaterNode creates a new repeater node
func NewRepeaterNode(name string, maxRepeats int) *RepeaterNode {
	return &RepeaterNode{
		DecoratorNode: DecoratorNode{
			BaseNode: BaseNode{
				name: name,
			},
		},
		maxRepeats:    maxRepeats,
		currentRepeat: 0,
	}
}

// Execute runs the repeater logic
func (rn *RepeaterNode) Execute(context *BehaviorContext) NodeStatus {
	if rn.child == nil {
		return StatusFailure
	}

	for {
		// Check if we've reached max repeats
		if rn.maxRepeats != -1 && rn.currentRepeat >= rn.maxRepeats {
			rn.Reset()
			return StatusSuccess
		}

		status := rn.child.Execute(context)
		switch status {
		case StatusSuccess:
			rn.currentRepeat++
			rn.child.Reset() // Reset child for next iteration
			continue
		case StatusFailure:
			rn.Reset()
			return StatusFailure
		case StatusRunning:
			return StatusRunning
		default:
			return StatusInvalid
		}
	}
}

// Reset resets the repeater node
func (rn *RepeaterNode) Reset() {
	rn.currentRepeat = 0
	rn.DecoratorNode.Reset()
}

// SucceederNode always returns success regardless of child result
type SucceederNode struct {
	DecoratorNode
}

// NewSucceederNode creates a new succeeder node
func NewSucceederNode(name string) *SucceederNode {
	return &SucceederNode{
		DecoratorNode: DecoratorNode{
			BaseNode: BaseNode{
				name: name,
			},
		},
	}
}

// Execute runs the succeeder logic
func (sn *SucceederNode) Execute(context *BehaviorContext) NodeStatus {
	if sn.child == nil {
		return StatusSuccess
	}

	status := sn.child.Execute(context)
	if status == StatusRunning {
		return StatusRunning
	}

	return StatusSuccess
}

// BehaviorTree represents a complete behavior tree for a unit
type BehaviorTree struct {
	root     BehaviorNode      // Root node of the tree
	context  *BehaviorContext  // Execution context
	status   NodeStatus        // Current tree status
	isActive bool              // Whether the tree is currently active
}

// NewBehaviorTree creates a new behavior tree with the given root node
func NewBehaviorTree(root BehaviorNode) *BehaviorTree {
	return &BehaviorTree{
		root:     root,
		status:   StatusInvalid,
		isActive: false,
	}
}

// Start begins execution of the behavior tree
func (bt *BehaviorTree) Start(unit *GameUnit, world *World) {
	bt.context = NewBehaviorContext(unit, world, 0)
	bt.isActive = true
	bt.status = StatusRunning
}

// Stop stops execution of the behavior tree
func (bt *BehaviorTree) Stop() {
	bt.isActive = false
	bt.status = StatusInvalid
	if bt.root != nil {
		bt.root.Reset()
	}
}

// Update executes one tick of the behavior tree
func (bt *BehaviorTree) Update(deltaTime time.Duration) NodeStatus {
	if !bt.isActive || bt.root == nil {
		return StatusInvalid
	}

	// Update context timing
	bt.context.DeltaTime = deltaTime

	// Execute the root node
	bt.status = bt.root.Execute(bt.context)

	// If tree completed, stop it
	if bt.status == StatusSuccess || bt.status == StatusFailure {
		bt.Stop()
	}

	return bt.status
}

// GetStatus returns the current status of the behavior tree
func (bt *BehaviorTree) GetStatus() NodeStatus {
	return bt.status
}

// IsActive returns whether the behavior tree is currently running
func (bt *BehaviorTree) IsActive() bool {
	return bt.isActive
}

// GetContext returns the behavior tree's execution context
func (bt *BehaviorTree) GetContext() *BehaviorContext {
	return bt.context
}

// Reset resets the behavior tree to initial state
func (bt *BehaviorTree) Reset() {
	if bt.root != nil {
		bt.root.Reset()
	}
	bt.status = StatusInvalid
}

// BehaviorTreeManager manages behavior trees for multiple units
type BehaviorTreeManager struct {
	trees map[int]*BehaviorTree // Unit ID -> Behavior Tree mapping
	world *World               // World reference
}

// NewBehaviorTreeManager creates a new behavior tree manager
func NewBehaviorTreeManager(world *World) *BehaviorTreeManager {
	return &BehaviorTreeManager{
		trees: make(map[int]*BehaviorTree),
		world: world,
	}
}

// SetBehaviorTree assigns a behavior tree to a unit
func (btm *BehaviorTreeManager) SetBehaviorTree(unitID int, tree *BehaviorTree) error {
	unit := btm.world.ObjectManager.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}

	// Stop any existing tree for this unit
	if existingTree, exists := btm.trees[unitID]; exists {
		existingTree.Stop()
	}

	// Start the new tree
	tree.Start(unit, btm.world)
	btm.trees[unitID] = tree

	return nil
}

// RemoveBehaviorTree removes a unit's behavior tree
func (btm *BehaviorTreeManager) RemoveBehaviorTree(unitID int) {
	if tree, exists := btm.trees[unitID]; exists {
		tree.Stop()
		delete(btm.trees, unitID)
	}
}

// Update updates all active behavior trees
func (btm *BehaviorTreeManager) Update(deltaTime time.Duration) {
	for unitID, tree := range btm.trees {
		// Check if unit still exists
		unit := btm.world.ObjectManager.GetUnit(unitID)
		if unit == nil || !unit.IsAlive() {
			// Unit no longer exists, remove its tree
			tree.Stop()
			delete(btm.trees, unitID)
			continue
		}

		// Update the behavior tree
		tree.Update(deltaTime)

		// Remove completed trees
		if !tree.IsActive() {
			delete(btm.trees, unitID)
		}
	}
}

// GetBehaviorTree returns the behavior tree for a unit
func (btm *BehaviorTreeManager) GetBehaviorTree(unitID int) (*BehaviorTree, bool) {
	tree, exists := btm.trees[unitID]
	return tree, exists
}

// GetActiveTrees returns the number of currently active behavior trees
func (btm *BehaviorTreeManager) GetActiveTrees() int {
	return len(btm.trees)
}

// HasBehaviorTree checks if a unit has an active behavior tree
func (btm *BehaviorTreeManager) HasBehaviorTree(unitID int) bool {
	_, exists := btm.trees[unitID]
	return exists
}