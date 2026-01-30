package renderer

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// ShaderManager manages GLSL shader programs
type ShaderManager struct {
	programs map[string]uint32             // Shader program name -> OpenGL program ID
	uniforms map[string]map[string]int32   // Program name -> uniform name -> location
}

// NewShaderManager creates a new shader manager
func NewShaderManager() *ShaderManager {
	return &ShaderManager{
		programs: make(map[string]uint32),
		uniforms: make(map[string]map[string]int32),
	}
}

// LoadShader loads and compiles a shader program from vertex and fragment shader files
func (sm *ShaderManager) LoadShader(name, vertexPath, fragmentPath string) error {
	// Read vertex shader source
	vertexSource, err := ioutil.ReadFile(vertexPath)
	if err != nil {
		return fmt.Errorf("failed to read vertex shader %s: %w", vertexPath, err)
	}

	// Read fragment shader source
	fragmentSource, err := ioutil.ReadFile(fragmentPath)
	if err != nil {
		return fmt.Errorf("failed to read fragment shader %s: %w", fragmentPath, err)
	}

	// Compile vertex shader
	vertexShader, err := sm.compileShader(string(vertexSource), gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile vertex shader %s: %w", vertexPath, err)
	}
	defer gl.DeleteShader(vertexShader)

	// Compile fragment shader
	fragmentShader, err := sm.compileShader(string(fragmentSource), gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile fragment shader %s: %w", fragmentPath, err)
	}
	defer gl.DeleteShader(fragmentShader)

	// Link shader program
	program, err := sm.linkProgram(vertexShader, fragmentShader)
	if err != nil {
		return fmt.Errorf("failed to link shader program %s: %w", name, err)
	}

	// Store the program
	sm.programs[name] = program

	// Initialize uniform location cache for this program
	sm.uniforms[name] = make(map[string]int32)

	log.Printf("Loaded shader program: %s (ID=%d)", name, program)
	return nil
}

// LoadShaderFromSource loads and compiles a shader program from source strings
func (sm *ShaderManager) LoadShaderFromSource(name, vertexSource, fragmentSource string) error {
	// Compile vertex shader
	vertexShader, err := sm.compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile vertex shader for %s: %w", name, err)
	}
	defer gl.DeleteShader(vertexShader)

	// Compile fragment shader
	fragmentShader, err := sm.compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("failed to compile fragment shader for %s: %w", name, err)
	}
	defer gl.DeleteShader(fragmentShader)

	// Link shader program
	program, err := sm.linkProgram(vertexShader, fragmentShader)
	if err != nil {
		return fmt.Errorf("failed to link shader program %s: %w", name, err)
	}

	// Store the program
	sm.programs[name] = program

	// Initialize uniform location cache for this program
	sm.uniforms[name] = make(map[string]int32)

	log.Printf("Loaded shader program from source: %s (ID=%d)", name, program)
	return nil
}

// compileShader compiles a single shader from source
func (sm *ShaderManager) compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	// Set shader source
	csource, free := gl.Strs(source + "\x00")
	defer free()
	gl.ShaderSource(shader, 1, csource, nil)

	// Compile shader
	gl.CompileShader(shader)

	// Check compilation status
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		gl.DeleteShader(shader)
		return 0, fmt.Errorf("shader compilation failed: %v", log)
	}

	return shader, nil
}

// linkProgram links vertex and fragment shaders into a program
func (sm *ShaderManager) linkProgram(vertexShader, fragmentShader uint32) (uint32, error) {
	program := gl.CreateProgram()

	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Check linking status
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		gl.DeleteProgram(program)
		return 0, fmt.Errorf("program linking failed: %v", log)
	}

	return program, nil
}

// UseShader activates a shader program
func (sm *ShaderManager) UseShader(name string) error {
	program, exists := sm.programs[name]
	if !exists {
		return fmt.Errorf("shader program %s not found", name)
	}

	gl.UseProgram(program)
	return nil
}

// getUniformLocation gets a uniform location and caches it
func (sm *ShaderManager) getUniformLocation(shaderName, uniformName string) (int32, error) {
	program, exists := sm.programs[shaderName]
	if !exists {
		return -1, fmt.Errorf("shader program %s not found", shaderName)
	}

	// Check cache first
	if location, cached := sm.uniforms[shaderName][uniformName]; cached {
		return location, nil
	}

	// Get uniform location
	location := gl.GetUniformLocation(program, gl.Str(uniformName+"\x00"))
	if location == -1 {
		return -1, fmt.Errorf("uniform %s not found in shader %s", uniformName, shaderName)
	}

	// Cache the location
	sm.uniforms[shaderName][uniformName] = location
	return location, nil
}

// SetUniformMat4 sets a mat4 uniform value
func (sm *ShaderManager) SetUniformMat4(shaderName, uniformName string, matrix mgl32.Mat4) error {
	location, err := sm.getUniformLocation(shaderName, uniformName)
	if err != nil {
		return err
	}

	gl.UniformMatrix4fv(location, 1, false, &matrix[0])
	return nil
}

// SetUniformVec3 sets a vec3 uniform value
func (sm *ShaderManager) SetUniformVec3(shaderName, uniformName string, vector mgl32.Vec3) error {
	location, err := sm.getUniformLocation(shaderName, uniformName)
	if err != nil {
		return err
	}

	gl.Uniform3fv(location, 1, &vector[0])
	return nil
}

// SetUniformFloat sets a float uniform value
func (sm *ShaderManager) SetUniformFloat(shaderName, uniformName string, value float32) error {
	location, err := sm.getUniformLocation(shaderName, uniformName)
	if err != nil {
		return err
	}

	gl.Uniform1f(location, value)
	return nil
}

// SetUniformInt sets an int uniform value
func (sm *ShaderManager) SetUniformInt(shaderName, uniformName string, value int32) error {
	location, err := sm.getUniformLocation(shaderName, uniformName)
	if err != nil {
		return err
	}

	gl.Uniform1i(location, value)
	return nil
}

// GetProgramID returns the OpenGL program ID for a shader
func (sm *ShaderManager) GetProgramID(name string) (uint32, bool) {
	program, exists := sm.programs[name]
	return program, exists
}

// ListShaders returns a list of all loaded shader names
func (sm *ShaderManager) ListShaders() []string {
	names := make([]string, 0, len(sm.programs))
	for name := range sm.programs {
		names = append(names, name)
	}
	return names
}

// Destroy cleans up all shader programs
func (sm *ShaderManager) Destroy() {
	for name, program := range sm.programs {
		gl.DeleteProgram(program)
		log.Printf("Deleted shader program: %s", name)
	}
	sm.programs = make(map[string]uint32)
	sm.uniforms = make(map[string]map[string]int32)
}