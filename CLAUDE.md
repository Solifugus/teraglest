# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Teraglest is a real-time strategy game project. This repository structure and commands will be populated once the source code is available.

## Build Commands

```bash
# Build the main application
go build ./cmd/teraglest

# Run the application
./teraglest

# Build and run in one command
go build ./cmd/teraglest && ./teraglest
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/data
```

## Code Architecture

### Core Components
- **Data Layer** (`internal/data/`) - XML parsing for tech trees, factions, units, resources ✅ COMPLETE
- **Game Engine** (`internal/engine/`) - Game logic, world state, game loop ✅ FOUNDATION COMPLETE
- **Graphics System** (`internal/graphics/`) - 3D rendering, OpenGL (planned)
- **Audio System** (`internal/audio/`) - Sound effects, music (planned)
- **Networking** (`internal/network/`) - Multiplayer support (planned)
- **AI System** (`internal/ai/`) - Computer player AI (planned)

### Key Directories
- `cmd/teraglest/` - Main executable entry point
- `internal/data/` - XML data parsing (tech trees, resources, factions, units) ✅ COMPLETE
- `internal/engine/` - Game engine foundation (game state, world management) ✅ FOUNDATION COMPLETE
- `pkg/formats/` - File format parsers (G3D models, etc.)
- `megaglest-source/` - Original Megaglest source code for reference
- `assets/` - Test data and game assets

### Core System Files (Phase 1.1-2.1 Complete)

**Data Layer (Phase 1.1-1.6):**
- `internal/data/techtree.go` + tests - Tech tree XML parsing ✅
- `internal/data/resource.go` + tests - Resource XML parsing ✅
- `internal/data/faction.go` + tests - Faction XML parsing ✅
- `internal/data/unit.go` + tests - Unit XML parsing ✅
- `pkg/formats/g3d.go` + tests - G3D binary model format parsing ✅
- `internal/data/cache.go` + tests - Thread-safe asset caching ✅
- `internal/data/assets.go` + tests - Asset management system ✅
- `internal/data/validator.go` + tests - Data validation and error handling ✅

**Game Engine (Phase 2.1):**
- `internal/engine/game.go` + tests - Game controller and state management ✅ NEW
- `internal/engine/world.go` + tests - World state and player management ✅ NEW

## Development Workflow

### Development Progress

**Phase 1: Data Layer (Complete)**
- ✅ Phase 1.1: Project structure and Go module setup
- ✅ Phase 1.2: XML parsing for tech trees and resources
- ✅ Phase 1.3: Faction and unit type data structures
- ✅ Phase 1.4: G3D model format parser
- ✅ Phase 1.5: Asset management and caching system
- ✅ Phase 1.6: Data validation and error handling

**Phase 2: Core Game Engine**
- ✅ Phase 2.1: Core game engine foundation - COMPLETED ✅ NEW
- [ ] Phase 2.2: Game object management - NEXT 🎯
- [ ] Phase 2.3: Turn/time management system

### Current Implementation Status
- **Tech Tree Parsing**: Fully functional - loads attack types, armor types, damage multipliers
- **Resource Parsing**: Fully functional - loads all resource definitions with properties
- **Faction Parsing**: Fully functional - loads starting resources, units, AI behavior
- **Unit Parsing**: Fully functional - loads parameters, skills, commands with complex nested structures
- **G3D Model Parsing**: Fully functional - loads binary 3D models with mesh/vertex counts and texture names
- **Asset Management**: Fully functional - thread-safe caching with LRU eviction and memory limits
- **Data Validation**: Fully functional - comprehensive validation with clear error messages
- **Game Engine Foundation**: Fully functional - game state management, world state, event system ✅ NEW
- **Test Coverage**: Comprehensive unit tests across all data parsing, asset management, validation, and game engine systems

### Common Development Tasks

```bash
# Run development iteration
go test ./... && go build ./cmd/teraglest && ./teraglest

# Add new dependencies
go mod tidy

# View module dependencies
go mod graph
```

### Configuration
- `go.mod` - Go module definition and dependencies
- Phase tracking in `DEVELOPMENT_PLAN.md`

## Notes for Claude Code

- Project uses Go 1.22.2 with standard library XML parsing
- All Megaglest XML data formats are successfully parsed and validated
- Phase 2.1 COMPLETE: Game engine foundation implemented (game state, world management, events)
- Current focus: Game object management (Phase 2.2) - unit/building lifecycle management
- Next major milestone: Complete game mechanics implementation (Phase 2)
- Full asset compatibility maintained with original Megaglest data

### Latest Achievements (Phase 2.1)
- Complete game state management system with thread-safe operations ✅
- World initialization with player creation and starting units ✅
- Real-time game loop with configurable frame rate (60 FPS target) ✅
- Event system for game notifications and state changes ✅
- Player management with human and AI player support ✅
- Resource generation and management framework ✅
- Performance-optimized updates (sub-millisecond world updates) ✅
- Successfully demonstrated complete game session with 2 players and 18 units ✅