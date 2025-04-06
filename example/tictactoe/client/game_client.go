package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "raylib [core] example - basic window")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	board := make([][]int, 3)
	for i := range board {
		board[i] = make([]int, 3)
	}

	mouseDown := false

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		width := int32(rl.GetScreenWidth())
		height := int32(rl.GetScreenHeight())

		size := int32(float64(min(width, height)) * 0.9)
		xPad := (width - size) / 2
		yPad := (height - size) / 2

		drawBoard(board, xPad, yPad, size)
		mousePos := rl.GetMousePosition()
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			if !mouseDown {
				if mousePos.X >= float32(xPad) && mousePos.X < float32(xPad+size) &&
					mousePos.Y >= float32(yPad) && mousePos.Y < float32(yPad+size) {
					step := size / 3

					indexX := int((mousePos.X - float32(xPad)) / float32(step))
					indexY := int((mousePos.Y - float32(yPad)) / float32(step))

					board[indexX][indexY] += 1
				}
			}
			mouseDown = true
		} else {
			mouseDown = false
		}
		rl.EndDrawing()
	}
}

func drawBoard(board [][]int, xPad, yPad, size int32) {
	rl.DrawRectangle(xPad, yPad, size, size, rl.Black)
	xPos := xPad
	yPos := yPad
	step := size / 3

	for range 2 {
		xPos += step
		rl.DrawRectangle(xPos-5, yPad, 10, size, rl.White)
		yPos += step
		rl.DrawRectangle(xPad, yPos-5, size, 10, rl.White)
	}

	padding := float32(step) * 0.5
	radius := padding * 0.9

	for x := range 3 {
		for y := range 3 {
			if board[x][y] == 1 {
				rl.DrawCircle(xPad+(step*int32(x))+int32(padding), yPad+(step*int32(y))+int32(padding), radius, rl.Blue)
			} else if board[x][y] == 2 {
				rl.DrawCircle(xPad+(step*int32(x))+int32(padding), yPad+(step*int32(y))+int32(padding), radius, rl.Red)
			}
		}
	}

}
