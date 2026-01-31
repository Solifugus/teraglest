package ui

import (
	"fmt"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

// ImGuiRenderer handles ImGui rendering with OpenGL
type ImGuiRenderer struct {
	shaderHandle      uint32
	vertHandle        uint32
	fragHandle        uint32
	attribLocationTex uint32
	attribLocationProjMtx uint32
	attribLocationVtxPos uint32
	attribLocationVtxUV uint32
	attribLocationVtxColor uint32
	vboHandle         uint32
	elementsHandle    uint32
	fontTexture       uint32

	// State backup
	lastActiveTexture        int32
	lastProgram             uint32
	lastTexture             uint32
	lastSampler             uint32
	lastArrayBuffer         uint32
	lastVertexArrayObject   uint32
	lastPolygonMode         [2]int32
	lastViewport            [4]int32
	lastScissorBox          [4]int32
	lastBlendSrcRgb        int32
	lastBlendDstRgb        int32
	lastBlendSrcAlpha      int32
	lastBlendDstAlpha      int32
	lastBlendEquationRgb   int32
	lastBlendEquationAlpha int32
	lastEnableBlend        bool
	lastEnableCullFace     bool
	lastEnableDepthTest    bool
	lastEnableScissorTest  bool
}

// NewImGuiRenderer creates a new ImGui OpenGL renderer
func NewImGuiRenderer() (*ImGuiRenderer, error) {
	renderer := &ImGuiRenderer{}

	err := renderer.createDeviceObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to create device objects: %w", err)
	}

	return renderer, nil
}

// createDeviceObjects creates OpenGL objects needed for ImGui rendering
func (r *ImGuiRenderer) createDeviceObjects() error {
	// Vertex shader
	vertexShaderSource := `
#version 330
uniform mat4 ProjMtx;
in vec2 Position;
in vec2 UV;
in vec4 Color;
out vec2 Frag_UV;
out vec4 Frag_Color;

void main() {
    Frag_UV = UV;
    Frag_Color = Color;
    gl_Position = ProjMtx * vec4(Position.xy, 0, 1);
}
` + "\x00"

	// Fragment shader
	fragmentShaderSource := `
#version 330
uniform sampler2D Texture;
in vec2 Frag_UV;
in vec4 Frag_Color;
out vec4 Out_Color;

void main() {
    Out_Color = Frag_Color * texture(Texture, Frag_UV.st);
}
` + "\x00"

	// Create and compile vertex shader
	r.vertHandle = gl.CreateShader(gl.VERTEX_SHADER)
	vertexShaderSourcePtr, vertexShaderSourceFree := gl.Strs(vertexShaderSource)
	defer vertexShaderSourceFree()
	gl.ShaderSource(r.vertHandle, 1, vertexShaderSourcePtr, nil)
	gl.CompileShader(r.vertHandle)

	// Check vertex shader compilation
	var status int32
	gl.GetShaderiv(r.vertHandle, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(r.vertHandle, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetShaderInfoLog(r.vertHandle, logLength, nil, &log[0])
		return fmt.Errorf("vertex shader compilation failed: %s", string(log))
	}

	// Create and compile fragment shader
	r.fragHandle = gl.CreateShader(gl.FRAGMENT_SHADER)
	fragmentShaderSourcePtr, fragmentShaderSourceFree := gl.Strs(fragmentShaderSource)
	defer fragmentShaderSourceFree()
	gl.ShaderSource(r.fragHandle, 1, fragmentShaderSourcePtr, nil)
	gl.CompileShader(r.fragHandle)

	// Check fragment shader compilation
	gl.GetShaderiv(r.fragHandle, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(r.fragHandle, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetShaderInfoLog(r.fragHandle, logLength, nil, &log[0])
		return fmt.Errorf("fragment shader compilation failed: %s", string(log))
	}

	// Create shader program
	r.shaderHandle = gl.CreateProgram()
	gl.AttachShader(r.shaderHandle, r.vertHandle)
	gl.AttachShader(r.shaderHandle, r.fragHandle)
	gl.LinkProgram(r.shaderHandle)

	// Check program linking
	gl.GetProgramiv(r.shaderHandle, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(r.shaderHandle, gl.INFO_LOG_LENGTH, &logLength)
		log := make([]byte, logLength)
		gl.GetProgramInfoLog(r.shaderHandle, logLength, nil, &log[0])
		return fmt.Errorf("shader program linking failed: %s", string(log))
	}

	// Get uniform and attribute locations
	r.attribLocationTex = uint32(gl.GetUniformLocation(r.shaderHandle, gl.Str("Texture\x00")))
	r.attribLocationProjMtx = uint32(gl.GetUniformLocation(r.shaderHandle, gl.Str("ProjMtx\x00")))
	r.attribLocationVtxPos = uint32(gl.GetAttribLocation(r.shaderHandle, gl.Str("Position\x00")))
	r.attribLocationVtxUV = uint32(gl.GetAttribLocation(r.shaderHandle, gl.Str("UV\x00")))
	r.attribLocationVtxColor = uint32(gl.GetAttribLocation(r.shaderHandle, gl.Str("Color\x00")))

	// Create buffers
	gl.GenBuffers(1, &r.vboHandle)
	gl.GenBuffers(1, &r.elementsHandle)

	// Create font texture
	r.createFontsTexture()

	return nil
}

// createFontsTexture creates the font atlas texture
func (r *ImGuiRenderer) createFontsTexture() {
	io := imgui.CurrentIO()
	image := io.Fonts().TextureDataRGBA32()

	// Create OpenGL texture
	gl.GenTextures(1, &r.fontTexture)
	gl.BindTexture(gl.TEXTURE_2D, r.fontTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		int32(image.Width), int32(image.Height), 0,
		gl.RGBA, gl.UNSIGNED_BYTE, image.Pixels)

	// Store texture ID in ImGui
	io.Fonts().SetTextureID(imgui.TextureID(r.fontTexture))
}

// Render renders the ImGui draw data
func (r *ImGuiRenderer) Render(drawData imgui.DrawData) {
	displaySize := drawData.DisplaySize()
	fbScale := drawData.FrameBufferScale()
	displayWidth := displaySize.X
	displayHeight := displaySize.Y
	fbWidth := int(displayWidth * fbScale.X)
	fbHeight := int(displayHeight * fbScale.Y)

	if fbWidth <= 0 || fbHeight <= 0 {
		return
	}

	// Backup GL state
	r.backupGLState()

	// Setup render state
	r.setupRenderState(drawData, fbWidth, fbHeight, displayWidth, displayHeight)

	// Will project scissor/clipping rectangles into framebuffer space
	clipOffset := drawData.DisplayPos()
	clipScale := drawData.FrameBufferScale()

	// Render command lists
	for _, commandList := range drawData.CommandLists() {
		// Upload vertex/index buffers
		vertexBuffer, vertexBufferSize := commandList.VertexBuffer()
		indexBuffer, indexBufferSize := commandList.IndexBuffer()

		gl.BindBuffer(gl.ARRAY_BUFFER, r.vboHandle)
		gl.BufferData(gl.ARRAY_BUFFER, vertexBufferSize, vertexBuffer, gl.STREAM_DRAW)

		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.elementsHandle)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexBufferSize, indexBuffer, gl.STREAM_DRAW)

		// Process each command
		for _, command := range commandList.Commands() {
			if command.HasUserCallback() {
				command.CallUserCallback(commandList)
			} else {
				// Project scissor/clipping rectangles into framebuffer space
				clipRect := command.ClipRect()
				clipRect.X = (clipRect.X - clipOffset.X) * clipScale.X
				clipRect.Y = (clipRect.Y - clipOffset.Y) * clipScale.Y
				clipRect.Z = (clipRect.Z - clipOffset.X) * clipScale.X
				clipRect.W = (clipRect.W - clipOffset.Y) * clipScale.Y

				if clipRect.X < float32(fbWidth) && clipRect.Y < float32(fbHeight) && clipRect.Z >= 0.0 && clipRect.W >= 0.0 {
					// Apply scissor/clipping rectangle
					gl.Scissor(int32(clipRect.X), int32(float32(fbHeight)-clipRect.W),
						int32(clipRect.Z-clipRect.X), int32(clipRect.W-clipRect.Y))

					// Bind texture, draw
					gl.BindTexture(gl.TEXTURE_2D, uint32(command.TextureID()))

					gl.DrawElementsWithOffset(gl.TRIANGLES, int32(command.ElementCount()),
						gl.UNSIGNED_SHORT, uintptr(command.IndexOffset()*2))
				}
			}
		}
	}

	// Restore modified GL state
	r.restoreGLState()
}

// setupRenderState sets up OpenGL state for ImGui rendering
func (r *ImGuiRenderer) setupRenderState(drawData imgui.DrawData, fbWidth, fbHeight int, displayWidth, displayHeight float32) {
	// Setup blend equation and function
	gl.Enable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.SCISSOR_TEST)

	// Setup viewport, orthographic projection matrix
	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
	orthoProjection := [4][4]float32{
		{2.0 / displayWidth, 0.0, 0.0, 0.0},
		{0.0, 2.0 / -displayHeight, 0.0, 0.0},
		{0.0, 0.0, -1.0, 0.0},
		{-1.0, 1.0, 0.0, 1.0},
	}

	gl.UseProgram(r.shaderHandle)
	gl.Uniform1i(int32(r.attribLocationTex), 0)
	gl.UniformMatrix4fv(int32(r.attribLocationProjMtx), 1, false, &orthoProjection[0][0])

	// Setup vertex attributes
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vboHandle)
	gl.EnableVertexAttribArray(r.attribLocationVtxPos)
	gl.EnableVertexAttribArray(r.attribLocationVtxUV)
	gl.EnableVertexAttribArray(r.attribLocationVtxColor)

	vertexSize, _, _, _ := imgui.VertexBufferLayout()
	vertexSizeInt32 := int32(vertexSize)
	gl.VertexAttribPointer(r.attribLocationVtxPos, 2, gl.FLOAT, false, vertexSizeInt32, gl.PtrOffset(0))
	gl.VertexAttribPointer(r.attribLocationVtxUV, 2, gl.FLOAT, false, vertexSizeInt32, gl.PtrOffset(8))
	gl.VertexAttribPointer(r.attribLocationVtxColor, 4, gl.UNSIGNED_BYTE, true, vertexSizeInt32, gl.PtrOffset(16))
}

// backupGLState backs up current OpenGL state
func (r *ImGuiRenderer) backupGLState() {
	gl.GetIntegerv(gl.ACTIVE_TEXTURE, &r.lastActiveTexture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GetIntegerv(gl.CURRENT_PROGRAM, (*int32)(unsafe.Pointer(&r.lastProgram)))
	gl.GetIntegerv(gl.TEXTURE_BINDING_2D, (*int32)(unsafe.Pointer(&r.lastTexture)))
	gl.GetIntegerv(gl.SAMPLER_BINDING, (*int32)(unsafe.Pointer(&r.lastSampler)))
	gl.GetIntegerv(gl.ARRAY_BUFFER_BINDING, (*int32)(unsafe.Pointer(&r.lastArrayBuffer)))
	gl.GetIntegerv(gl.VERTEX_ARRAY_BINDING, (*int32)(unsafe.Pointer(&r.lastVertexArrayObject)))
	gl.GetIntegerv(gl.POLYGON_MODE, &r.lastPolygonMode[0])
	gl.GetIntegerv(gl.VIEWPORT, &r.lastViewport[0])
	gl.GetIntegerv(gl.SCISSOR_BOX, &r.lastScissorBox[0])
	gl.GetIntegerv(gl.BLEND_SRC_RGB, &r.lastBlendSrcRgb)
	gl.GetIntegerv(gl.BLEND_DST_RGB, &r.lastBlendDstRgb)
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &r.lastBlendSrcAlpha)
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &r.lastBlendDstAlpha)
	gl.GetIntegerv(gl.BLEND_EQUATION_RGB, &r.lastBlendEquationRgb)
	gl.GetIntegerv(gl.BLEND_EQUATION_ALPHA, &r.lastBlendEquationAlpha)
	r.lastEnableBlend = gl.IsEnabled(gl.BLEND)
	r.lastEnableCullFace = gl.IsEnabled(gl.CULL_FACE)
	r.lastEnableDepthTest = gl.IsEnabled(gl.DEPTH_TEST)
	r.lastEnableScissorTest = gl.IsEnabled(gl.SCISSOR_TEST)
}

// restoreGLState restores previous OpenGL state
func (r *ImGuiRenderer) restoreGLState() {
	gl.UseProgram(r.lastProgram)
	gl.BindTexture(gl.TEXTURE_2D, r.lastTexture)
	gl.BindSampler(0, r.lastSampler)
	gl.ActiveTexture(uint32(r.lastActiveTexture))
	gl.BindVertexArray(r.lastVertexArrayObject)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.lastArrayBuffer)
	gl.BlendEquationSeparate(uint32(r.lastBlendEquationRgb), uint32(r.lastBlendEquationAlpha))
	gl.BlendFuncSeparate(uint32(r.lastBlendSrcRgb), uint32(r.lastBlendDstRgb),
		uint32(r.lastBlendSrcAlpha), uint32(r.lastBlendDstAlpha))

	if r.lastEnableBlend {
		gl.Enable(gl.BLEND)
	} else {
		gl.Disable(gl.BLEND)
	}

	if r.lastEnableCullFace {
		gl.Enable(gl.CULL_FACE)
	} else {
		gl.Disable(gl.CULL_FACE)
	}

	if r.lastEnableDepthTest {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}

	if r.lastEnableScissorTest {
		gl.Enable(gl.SCISSOR_TEST)
	} else {
		gl.Disable(gl.SCISSOR_TEST)
	}

	gl.PolygonMode(gl.FRONT_AND_BACK, uint32(r.lastPolygonMode[0]))
	gl.Viewport(r.lastViewport[0], r.lastViewport[1], r.lastViewport[2], r.lastViewport[3])
	gl.Scissor(r.lastScissorBox[0], r.lastScissorBox[1], r.lastScissorBox[2], r.lastScissorBox[3])
}

// Cleanup destroys ImGui renderer resources
func (r *ImGuiRenderer) Cleanup() {
	if r.vboHandle != 0 {
		gl.DeleteBuffers(1, &r.vboHandle)
		r.vboHandle = 0
	}
	if r.elementsHandle != 0 {
		gl.DeleteBuffers(1, &r.elementsHandle)
		r.elementsHandle = 0
	}
	if r.shaderHandle != 0 && r.vertHandle != 0 {
		gl.DetachShader(r.shaderHandle, r.vertHandle)
	}
	if r.shaderHandle != 0 && r.fragHandle != 0 {
		gl.DetachShader(r.shaderHandle, r.fragHandle)
	}
	if r.vertHandle != 0 {
		gl.DeleteShader(r.vertHandle)
		r.vertHandle = 0
	}
	if r.fragHandle != 0 {
		gl.DeleteShader(r.fragHandle)
		r.fragHandle = 0
	}
	if r.shaderHandle != 0 {
		gl.DeleteProgram(r.shaderHandle)
		r.shaderHandle = 0
	}
	if r.fontTexture != 0 {
		gl.DeleteTextures(1, &r.fontTexture)
		imgui.CurrentIO().Fonts().SetTextureID(0)
		r.fontTexture = 0
	}
}