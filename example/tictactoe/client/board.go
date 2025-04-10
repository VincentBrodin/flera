package main

import rl "github.com/gen2brain/raylib-go/raylib"

type Board struct {
	State     [][]int
	mouseDown bool
}

func (b *Board) Draw() {
	xPad, yPad, size := b.GetScreenBounds()
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
			if b.State[x][y] == 1 {
				rl.DrawCircle(xPad+(step*int32(x))+int32(padding), yPad+(step*int32(y))+int32(padding), radius, rl.Blue)
			} else if b.State[x][y] == 2 {
				rl.DrawCircle(xPad+(step*int32(x))+int32(padding), yPad+(step*int32(y))+int32(padding), radius, rl.Red)
			}
		}
	}
}

func (b *Board) GetScreenBounds() (int32, int32, int32) {
	width := int32(rl.GetScreenWidth())
	height := int32(rl.GetScreenHeight())
	size := int32(float64(min(width, height)) * 0.9)
	xPad := (width - size) / 2
	yPad := (height - size) / 2
	return xPad, yPad, size
}

func (b *Board) Click() (int, int, bool) {
	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		if !b.mouseDown {
			b.mouseDown = true
			mousePos := rl.GetMousePosition()

			xPad, yPad, size := b.GetScreenBounds()
			if mousePos.X >= float32(xPad) && mousePos.X < float32(xPad+size) &&
				mousePos.Y >= float32(yPad) && mousePos.Y < float32(yPad+size) {
				step := size / 3

				x := int((mousePos.X - float32(xPad)) / float32(step))
				y := int((mousePos.Y - float32(yPad)) / float32(step))
				return x, y, true
			}
		}

	} else {
		b.mouseDown = false

	}
	return -1, -1, false
}
