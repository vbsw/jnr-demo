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
	"github.com/vbsw/shaders"
	"runtime"
)

const (
	canvasWidth      = 840
	canvasHeight     = 360
	playerWidth      = 16
	playerHeight     = 16
	gapWidth         = 150
	platformWidth    = canvasWidth - gapWidth
	platformHeight   = 150
	wallWidth        = 50
	jumpHeightA      = 100
	jumpHeightB      = 70
	speedX           = 4
	speedY           = 3
	breakY           = 2
	jumpAcceleration = 8
)

var (
	modelMatrix []float32
	moveLeft    bool
	moveRight   bool
	moveUp      bool
	moveDown    bool
	jump        bool
	jumpingA    bool
	jumpingB    bool
	wallLock    bool
	playerX     float32
	playerY     float32
	jumpY       float32
	jumpSpeed   float32
)

func init() {
	runtime.LockOSThread()
}

func main() {
	err := glfw.Init()

	if err == nil {
		var window *glfw.Window
		defer glfw.Terminate()
		window, err = glfw.CreateWindow(canvasWidth, canvasHeight, "OpenGL Example", nil, nil)

		if err == nil {
			defer window.Destroy()
			window.SetKeyCallback(onKey)
			window.SetSizeCallback(onResize)
			window.MakeContextCurrent()
			glfw.SwapInterval(1)
			err = gl.Init()

			if err == nil {
				shader := shaders.NewPrimitiveShader()
				err = initShaderProgram(shader)

				if err == nil {
					defer gl.DeleteShader(shader.VertexShaderID)
					defer gl.DeleteShader(shader.FragmentShaderID)
					defer gl.DeleteProgram(shader.ProgramID)
					vbos := newVBOs(3)
					defer gl.DeleteBuffers(int32(len(vbos)), &vbos[0])
					vaos := newVAOs(3)
					defer gl.DeleteVertexArrays(int32(len(vaos)), &vaos[0])
					modelMatrix = newModelMatrix()

					bindLevelObjects(shader, vaos[:2], vbos[:2])
					bindPlayerObjects(shader, vaos[2:], vbos[2:])
					gl.UseProgram(shader.ProgramID)
					setProjection(shader)
					resetPlayer()

					// transparancy
					// gl.Enable(gl.BLEND);
					// gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);

					// wireframe mode
					// gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)

					for !window.ShouldClose() {
						updateMovement()
						gl.ClearColor(0, 0, 0, 0)
						gl.Clear(gl.COLOR_BUFFER_BIT)

						// draw level
						setLevelModel(shader)
						for _, vao := range vaos[:2] {
							gl.BindVertexArray(vao)
							gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
						}

						// draw player
						setPlayerModel(shader)
						for _, vao := range vaos[2:] {
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
	if err != nil {
		fmt.Println(err.Error())
	}
}

func onKey(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyEscape:
			window.SetShouldClose(true)
		case glfw.KeyLeft:
			moveLeft = true
		case glfw.KeyJ:
			moveLeft = true
		case glfw.KeyRight:
			moveRight = true
		case glfw.KeyL:
			moveRight = true
		case glfw.KeyUp:
			moveUp = true
		case glfw.KeyI:
			moveUp = true
		case glfw.KeyDown:
			moveDown = true
		case glfw.KeyK:
			moveDown = true
		case glfw.KeyS:
			wallLock = !wallLock
			if wallLock {
				fmt.Println("wall lock ON")
			} else {
				fmt.Println("wall lock OFF")
			}
		case glfw.KeySpace:
			jump = true
			if playerY <= jumpY && playerX < platformWidth || (playerX == platformWidth || playerX == canvasWidth-wallWidth-playerWidth) {
				jumpingA = true
				jumpY = playerY
				jumpSpeed = jumpAcceleration
			}
		case glfw.KeyF:
			jump = true
			if playerY <= jumpY && playerX < platformWidth || (playerX == platformWidth || playerX == canvasWidth-wallWidth-playerWidth) {
				jumpingA = true
				jumpY = playerY
				jumpSpeed = jumpAcceleration
			}
		case glfw.KeyR:
			resetPlayer()
		}
	} else if action == glfw.Release {
		switch key {
		case glfw.KeyLeft:
			moveLeft = false
		case glfw.KeyJ:
			moveLeft = false
		case glfw.KeyRight:
			moveRight = false
		case glfw.KeyL:
			moveRight = false
		case glfw.KeyUp:
			moveUp = false
		case glfw.KeyI:
			moveUp = false
		case glfw.KeyDown:
			moveDown = false
		case glfw.KeyK:
			moveDown = false
		case glfw.KeySpace:
			jump = false
		case glfw.KeyF:
			jump = false
		}
	}
}

func onResize(w *glfw.Window, width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func initShaderProgram(shader *shaders.Shader) error {
	var err error
	shader.VertexShaderID, err = newShader(gl.VERTEX_SHADER, shader.VertexShader)

	if err == nil {
		shader.FragmentShaderID, err = newShader(gl.FRAGMENT_SHADER, shader.FragmentShader)

		if err == nil {
			shader.ProgramID, err = newProgram(shader)

			if err == nil {
				shader.PositionLocation = gl.GetAttribLocation(shader.ProgramID, shader.PositionAttribute)
				shader.ColorLocation = gl.GetAttribLocation(shader.ProgramID, shader.ColorAttribute)
				shader.ProjectionLocation = gl.GetUniformLocation(shader.ProgramID, shader.ProjectionUniform)
				shader.ModelLocation = gl.GetUniformLocation(shader.ProgramID, shader.ModelUniform)

			} else {
				gl.DeleteShader(shader.VertexShaderID)
				gl.DeleteShader(shader.FragmentShaderID)
			}
		} else {
			gl.DeleteShader(shader.VertexShaderID)
		}
	}
	return err
}

func newShader(shaderType uint32, shaderSource **uint8) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	gl.ShaderSource(shader, 1, shaderSource, nil)
	gl.CompileShader(shader)
	err := checkShader(shader, gl.COMPILE_STATUS)

	if err != nil {
		gl.DeleteShader(shader)
	}
	return shader, err
}

func newProgram(shader *shaders.Shader) (uint32, error) {
	program := gl.CreateProgram()
	gl.AttachShader(program, shader.VertexShaderID)
	gl.AttachShader(program, shader.FragmentShaderID)
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

func bindLevelObjects(shader *shaders.Shader, vaos, vbos []uint32) {
	pointsA := newPoints(0, 0, platformWidth, platformHeight)
	pointsB := newPoints(canvasWidth-wallWidth, 0, wallWidth, 340)
	bindObjects(vaos[0], vbos[0], pointsA, shader)
	bindObjects(vaos[1], vbos[1], pointsB, shader)
}

func bindPlayerObjects(shader *shaders.Shader, vaos, vbos []uint32) {
	pointsA := newPoints(0, 0, playerWidth, playerHeight)
	// green color
	pointsA[3] = 0.0
	pointsA[5] = 0.0
	pointsA[10] = 0.0
	pointsA[12] = 0.0
	pointsA[17] = 0.0
	pointsA[19] = 0.0
	pointsA[24] = 0.0
	pointsA[26] = 0.0
	bindObjects(vaos[0], vbos[0], pointsA, shader)
}

func bindObjects(vao, vbo uint32, points []float32, shader *shaders.Shader) {
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(uint32(shader.PositionLocation))
	gl.EnableVertexAttribArray(uint32(shader.ColorLocation))

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(points)*4, gl.Ptr(points), gl.STATIC_DRAW)
	// position
	gl.VertexAttribPointer(uint32(shader.PositionLocation), 3, gl.FLOAT, false, 7*4, gl.PtrOffset(0))
	// color
	gl.VertexAttribPointer(uint32(shader.ColorLocation), 4, gl.FLOAT, false, 7*4, gl.PtrOffset(3*4))
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

func newModelMatrix() []float32 {
	modelMatrix = make([]float32, 4*4)
	modelMatrix[0] = 1.0
	modelMatrix[5] = 1.0
	modelMatrix[10] = 1.0
	modelMatrix[15] = 1.0
	return modelMatrix
}

func setProjection(shader *shaders.Shader) {
	matrix := make([]float32, 4*4)
	matrix[0] = 2.0 / float32(canvasWidth)
	matrix[5] = 2.0 / float32(canvasHeight)
	matrix[12] = -1.0
	matrix[13] = -1.0
	matrix[15] = 1.0
	gl.UniformMatrix4fv(shader.ProjectionLocation, 1, false, &matrix[0])
}

func setLevelModel(shader *shaders.Shader) {
	modelMatrix[12] = 0.0
	modelMatrix[13] = 0.0
	gl.UniformMatrix4fv(shader.ModelLocation, 1, false, &modelMatrix[0])
}

func setPlayerModel(shader *shaders.Shader) {
	modelMatrix[12] = playerX
	modelMatrix[13] = playerY
	gl.UniformMatrix4fv(shader.ModelLocation, 1, false, &modelMatrix[0])
}

func updateMovement() {
	if jumpingA {
		if playerY-jumpY < jumpHeightA {
			playerY += jumpSpeed
			jumpSpeed -= 0.2
			if playerY-jumpY < jumpHeightB && !jump {
				jumpingA = false
				jumpingB = true
			}
		} else {
			jumpingA = false
			playerY += -speedY
		}
	} else if jumpingB {
		if playerY-jumpY < jumpHeightB {
			playerY += jumpSpeed
			jumpSpeed -= 0.2
		} else {
			jumpingB = false
			playerY += -speedY
		}
	} else if playerY > platformHeight || playerX >= platformWidth {
		playerY += -speedY
		if playerX < platformWidth && playerY < platformHeight {
			playerY = platformHeight
			jumpY = platformHeight
		}
	} else if playerX > platformWidth {
		playerY += -speedY
	}
	if moveLeft {
		if wallLock {
			if jump || playerX != canvasWidth-wallWidth-playerWidth || moveDown {
				playerX += -speedX
			} else {
				playerY += breakY
			}
		} else {
			playerX += -speedX
		}
		if playerX < platformWidth && playerY < platformHeight {
			playerX = platformWidth
			playerY += breakY
		}
	} else if moveRight {
		if wallLock {
			if jump || playerX != platformWidth || moveDown {
				playerX += speedX
			} else {
				playerY += breakY
			}
		} else {
			playerX += speedX
		}
		if playerX > canvasWidth-wallWidth-playerWidth {
			playerX = canvasWidth - wallWidth - playerWidth
			playerY += breakY
		}
	}
}

func resetPlayer() {
	playerX = (canvasWidth-150)/2 + playerWidth/2
	playerY = canvasHeight - canvasHeight/3
	jumpY = platformHeight
}
