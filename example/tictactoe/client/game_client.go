package main

import (
	"flera/pkg/client"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SET_TEAM     uint32 = 1
	UPDATE_STATE uint32 = 2
)

var playerTurn bool = false
var team int = 0
var mouseDown bool = false
var board *Board
var winner int = 0

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
	// rl setup
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Client")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	triedToConnect := false

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if !c.Connected {
			rl.DrawText("Connecting", 0, 0, 20, rl.Black)
			if !triedToConnect {
				triedToConnect = true
				go c.Connect(":5050")
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
					c.Send(UPDATE_STATE, data)
				}
			}
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
