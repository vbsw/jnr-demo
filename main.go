/*
 *          Copyright 2020, Vitali Baumtrok.
 * Distributed under the Boost Software License, Version 1.0.
 *     (See accompanying file LICENSE or copy at
 *        http://www.boost.org/LICENSE_1_0.txt)
 */

package main

import (
	"errors"
	"fmt"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/vbsw/plainshader"
	"runtime"
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
				var vertexShader uint32
				vertexShader, err = loadShader(gl.VERTEX_SHADER, plainshader.VertexShader)

				if err == nil {
					var fragmentShader uint32
					fragmentShader, err = loadShader(gl.FRAGMENT_SHADER, plainshader.FragmentShader)

					if err == nil {
						var program uint32
						program, err = loadProgram(vertexShader, fragmentShader)

						if err == nil {
							defer gl.DeleteProgram(program)
							levelVBOs := newVBOs(2)
							defer gl.DeleteBuffers(int32(len(levelVBOs)), &levelVBOs[0])
							levelVAOs := newVAOs(2)
							defer gl.DeleteVertexArrays(int32(len(levelVAOs)), &levelVAOs[0])
							bindLevelObjects(levelVAOs, levelVBOs)
							gl.UseProgram(program)

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
									gl.DrawArrays(gl.TRIANGLES, 0, 3)
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

func loadProgram(vShader, fShader uint32) (uint32, error) {
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
	pointsA := newPoints(-0.5, 0.75, -0.25, 0.5, -0.75, 0.5)
	pointsB := newPoints(0.5, -0.25, 0.75, -0.5, 0.25, -0.5)
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

func newPoints(aX, aY, bX, bY, cX, cY float32) []float32 {
	points := make([]float32, 21)
	points[0] = aX
	points[1] = aY
	points[2] = 0.0
	points[3] = 1.0
	points[4] = 1.0
	points[5] = 1.0
	points[6] = 1.0
	points[7] = bX
	points[8] = bY
	points[9] = 0.0
	points[10] = 1.0
	points[11] = 1.0
	points[12] = 1.0
	points[13] = 1.0
	points[14] = cX
	points[15] = cY
	points[16] = 0.0
	points[17] = 1.0
	points[18] = 1.0
	points[19] = 1.0
	points[20] = 1.0
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
