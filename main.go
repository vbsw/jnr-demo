/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

/*
#include <string.h>

char *vertex_shader = "#version 130\n\nin vec3 vertexPosition;\nin vec4 vertexColor;\nout vec4 fragementColor;\n\nuniform mat4 projection = mat4(1.0);\n\nvoid main() {\n\tgl_Position = projection * vec4(vertexPosition, 1.0f);\n\tfragementColor = vertexColor;\n}";
char *fragment_shader = "#version 130\n\nin vec4 fragementColor;\nout vec4 color;\n\nvoid main() {\n\tcolor = fragementColor;\n}";
char *projection_name = "projection";
*/
import "C"
import (
	"errors"
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"runtime"
	"unsafe"
)

var (
	// vertexShaderSrc contains the following program:
	//
	//   #version 130
	//
	//   in vec3 vertexPosition;
	//   in vec4 vertexColor;
	//   out vec4 fragementColor;
	//
	//   uniform mat4 projection = mat4(1.0);
	//
	//   void main() {
	//     gl_Position = projection * vec4(vertexPosition, 1.0f);
	//     fragementColor = vertexColor;
	//   }
	vertexShaderSrc = (**uint8)(unsafe.Pointer(&C.vertex_shader))

	// fragmentShaderSrc contains the following program:
	//
	//   #version 130
	//
	//   in vec4 fragementColor;
	//   out vec4 color;
	//
	//   void main () {
	//     color = texture(ourTexture, TexCoord);
	//   }
	fragmentShaderSrc = (**uint8)(unsafe.Pointer(&C.fragment_shader))

	projectionName = (*uint8)(unsafe.Pointer(C.projection_name))
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()

	if err == nil {
		var window *glfw.Window
		defer glfw.Terminate()
		window, err = glfw.CreateWindow(840, 360, "OpenGL Example", nil, nil)

		if err == nil {
			defer window.Destroy()
			window.SetKeyCallback(onKey)
			window.SetSizeCallback(onResize)
			window.MakeContextCurrent()
			err = gl.Init()

			if err == nil {
				var vShader uint32
				vShader, err = loadShader(gl.VERTEX_SHADER, vertexShaderSrc)

				if err == nil {
					var fShader uint32
					fShader, err = loadShader(gl.FRAGMENT_SHADER, fragmentShaderSrc)

					if err == nil {
						var program uint32
						program, err = newProgram(vShader, fShader)

						if err == nil {
							defer gl.DeleteProgram(program)
							levelVBOs := newVBOs(2)
							defer gl.DeleteBuffers(int32(len(levelVBOs)), &levelVBOs[0])
							levelVAOs := newVAOs(2)
							defer gl.DeleteVertexArrays(int32(len(levelVAOs)), &levelVAOs[0])

							bindLevelObjects(levelVAOs, levelVBOs)
							gl.UseProgram(program)
							setProjection(program, 840, 360)

							// transparancy
							// gl.Enable(gl.BLEND);
							// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

							// wireframe mode
							// gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

							for !window.ShouldClose() {
								gl.ClearColor(0, 0, 0, 0)
								gl.Clear(gl.COLOR_BUFFER_BIT)

								for _, vao := range levelVAOs {
									gl.BindVertexArray(vao)
									gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
								}
								window.SwapBuffers()
								glfw.PollEvents()
							}
						}
					}
				}
			}
		}
	}
	if err != nil {
		fmt.Println(err.Error())
	}
}

func onKey(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
}

func onResize(w *glfw.Window, width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func loadShader(shaderType uint32, shaderSource **uint8) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, 1, shaderSource, nil)
	gl.CompileShader(shader)
	err := checkShader(shader, gl.COMPILE_STATUS)

	if err != nil {
		gl.DeleteShader(shader)
	}
	return shader, err
}

func newProgram(vShader, fShader uint32) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, vShader)
	gl.AttachShader(program, fShader)
	gl.LinkProgram(program)
	err := checkProgram(program, gl.LINK_STATUS)

	if err == nil {
		gl.ValidateProgram(program)
		err = checkProgram(program, gl.VALIDATE_STATUS)
	}
	if err != nil {
		gl.DeleteProgram(program)
	}
	return program, err
}

func checkShader(shader, statusType uint32) error {
	var status int32
	var err error

	gl.GetShaderiv(shader, statusType, &status)

	if status == gl.FALSE {
		var length int32
		var infoLog string

		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &length)

		if length > 0 {
			infoLogBytes := make([]byte, length)
			gl.GetShaderInfoLog(shader, length, nil, &infoLogBytes[0])
			infoLog = string(infoLogBytes)
		}
		err = errors.New("shader " + infoLog)
	}
	return err
}

func checkProgram(program, statusType uint32) error {
	var status int32
	var err error

	gl.GetProgramiv(program, statusType, &status)

	if status == gl.FALSE {
		var length int32
		var infoLog string

		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &length)

		if length > 0 {
			infoLogBytes := make([]byte, length)
			gl.GetProgramInfoLog(program, length, nil, &infoLogBytes[0])
			infoLog = string(infoLogBytes)
		}
		err = errors.New("program " + infoLog)
	}
	return err
}

func newVBOs(n int) []uint32 {
	vbos := make([]uint32, n)
	gl.GenBuffers(int32(len(vbos)), &vbos[0])
	return vbos
}

func newVAOs(n int) []uint32 {
	vaos := make([]uint32, n)
	gl.GenVertexArrays(int32(len(vaos)), &vaos[0])
	return vaos
}

func bindLevelObjects(vaos, vbos []uint32) {
	pointsA := newPoints(200, 250, 30, 30)
	pointsB := newPoints(600, 70, 30, 30)
	bindObjects(vaos[0], vbos[0], pointsA)
	bindObjects(vaos[1], vbos[1], pointsB)
}

func bindObjects(vao, vbo uint32, points []float32) {
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*4, gl.Ptr(points), gl.STATIC_DRAW)
	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 7*4, gl.PtrOffset(0))
	// color
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 7*4, gl.PtrOffset(3*4))
}

func newPoints(aX, aY, width, height float32) []float32 {
	points := make([]float32, 28)
	points[0] = aX + width
	points[1] = aY + height
	points[2] = 0.0
	points[3] = 1.0
	points[4] = 1.0
	points[5] = 1.0
	points[6] = 1.0
	points[7] = aX + width
	points[8] = aY
	points[9] = 0.0
	points[10] = 1.0
	points[11] = 1.0
	points[12] = 1.0
	points[13] = 1.0
	points[14] = aX
	points[15] = aY + height
	points[16] = 0.0
	points[17] = 1.0
	points[18] = 1.0
	points[19] = 1.0
	points[20] = 1.0
	points[21] = aX
	points[22] = aY
	points[23] = 0.0
	points[24] = 1.0
	points[25] = 1.0
	points[26] = 1.0
	points[27] = 1.0
	return points
}

func bindBuffer(index int, buffers []uint32, points []float32) {
	gl.BindBuffer(gl.ARRAY_BUFFER, buffers[index])
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*4, gl.Ptr(points), gl.STATIC_DRAW)

	// position
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 7*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// color
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 7*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)
}

func setProjection(shader uint32, width, height int) {
	location := gl.GetUniformLocation(shader, projectionName)
	projection := make([]float32, 4*4)
	projection[0] = 2.0 / float32(width)
	projection[5] = 2.0 / float32(height)
	projection[12] = -1.0
	projection[13] = -1.0
	projection[15] = 1.0
	gl.UniformMatrix4fv(location, 1, false, &projection[0])
}

