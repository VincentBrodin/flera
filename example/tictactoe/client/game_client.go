package main

import (
	"bytes"
	"encoding/binary"
	"flera/client"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SET_TEAM     uint32 = 1
	UPDATE_STATE uint32 = 2
	MOUSE_POS    uint32 = 3
)

var playerTurn bool = false
var team int = 0
var mouseDown bool = false
var board *Board
var winner int = 0
var oppX int32
var oppY int32

func main() {
	board = new(Board)
	board.State = make([][]int, 3)
	for i := range board.State {
		board.State[i] = make([]int, 3)
	}
	// client setup
	c := client.New()
	c.Register(SET_TEAM, SetTeam)
	c.Register(UPDATE_STATE, UpdateState)
	c.Register(MOUSE_POS, MousePos)
	// rl setup
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Client")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	triedToConnect := false

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if !c.Connected() {
			rl.DrawText("Connecting", 0, 0, 20, rl.Black)
			if !triedToConnect {
				triedToConnect = true
				go c.Connect(":2489")
			}
		} else {
			board.Draw()

			if winner != 0 {
				if winner == team {
					rl.DrawText("You win", 0, 0, 20, rl.Black)
				} else {
					rl.DrawText("You lose", 0, 0, 20, rl.Black)
				}
			} else if playerTurn {
				rl.DrawText("Your turn", 0, 0, 20, rl.Black)
				x, y, ok := board.Click()

				if ok && board.State[x][y] == 0 {
					data := []byte{uint8(x), uint8(y)}
					playerTurn = false
					board.State[x][y] = team
					c.SendSafe(UPDATE_STATE, data)
				}
			}

			mousePos := rl.GetMousePosition()
			if team == 1 {
				rl.DrawCircle(int32(mousePos.X), int32(mousePos.Y), 10, rl.Blue)
				rl.DrawCircle(oppX, oppY, 5, rl.Red)
			} else {
				rl.DrawCircle(int32(mousePos.X), int32(mousePos.Y), 10, rl.Red)
				rl.DrawCircle(oppX, oppY, 5, rl.Blue)
			}

			xPad, yPad, size := board.GetScreenBounds()
			x := (mousePos.X - float32(xPad)) / float32(size)
			xBuf := new(bytes.Buffer)
			if err := binary.Write(xBuf, binary.BigEndian, x); err != nil {
				fmt.Println(err)
				continue
			}

			y := (mousePos.Y - float32(yPad)) / float32(size)
			yBuf := new(bytes.Buffer)
			if err := binary.Write(yBuf, binary.BigEndian, y); err != nil {
				fmt.Println(err)
				continue
			}
			data := append(xBuf.Bytes(), yBuf.Bytes()...)
			if err := c.SendFast(MOUSE_POS, data); err != nil {
				fmt.Println(err)
			}

			// fmt.Println(x, y)
		}

		rl.EndDrawing()
	}
}

func SetTeam(c *client.Client, data []byte) error {
	team = int(data[0])
	if team == 1 {
		playerTurn = true
	}
	for x := range board.State {
		for y := range board.State[x] {
			board.State[x][y] = 0
		}
	}
	return nil
}

func UpdateState(c *client.Client, data []byte) error {
	fmt.Println(data)
	if int(data[0]) == team {
		playerTurn = true
	} else {
		playerTurn = false
	}

	winner = int(data[1])

	state := data[2:]
	for i := range state {
		x := i % 3
		y := i / 3
		fmt.Printf("x: %d y: %d\n", x, y)
		board.State[x][y] = int(state[i])
	}
	return nil
}

func MousePos(c *client.Client, data []byte) error {
	var id uint32
	if err := binary.Read(bytes.NewReader(data[:4]), binary.BigEndian, &id); err != nil {
		return err
	}

	if id == c.Id {
		return nil
	}
	var x float32
	if err := binary.Read(bytes.NewReader(data[4:8]), binary.BigEndian, &x); err != nil {
		return err
	}

	var y float32
	if err := binary.Read(bytes.NewReader(data[8:]), binary.BigEndian, &y); err != nil {
		return err
	}

	xPad, yPad, size := board.GetScreenBounds()

	oppX = int32(float32(xPad) + x*float32(size))
	oppY = int32(float32(yPad) + y*float32(size))

	return nil
}
