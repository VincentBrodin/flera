package main

import (
	"bytes"
	"encoding/binary"
	"flera/pkg/server"
	"fmt"
	"sync"
)

const (
	SET_TEAM     uint32 = 1
	UPDATE_STATE uint32 = 2
	MOUSE_POS    uint32 = 3
)

type Client struct {
	Id     uint32
	TeamId int
}

var teamA *Client
var teamB *Client
var curTeam *Client
var state [][]int
var mu sync.Mutex

func main() {
	state = make([][]int, 3)
	for i := range state {
		state[i] = make([]int, 3)
	}

	s := server.New()
	s.OnConn = OnConn
	s.OnDisConn = OnDisConn

	s.Register(UPDATE_STATE, UpdateState)
	s.Register(MOUSE_POS, MousePos)
	fmt.Println(s.Start(":2489"))
}

func MousePos(s *server.Server, connId uint32, data []byte) error {
	// var x float32
	// if err := binary.Read(bytes.NewReader(data[:4]), binary.BigEndian, &x); err != nil {
	// 	return err
	// }
	// var y float32
	// if err := binary.Read(bytes.NewReader(data[4:]), binary.BigEndian, &y); err != nil {
	// 	return err
	// }

	idBuf := new(bytes.Buffer)
	if err := binary.Write(idBuf, binary.BigEndian, connId); err != nil {
		return err
	}
	packet := append(idBuf.Bytes(), data...)
	// fmt.Println("TICK")
	return s.BroadcastFast(MOUSE_POS, packet)
}

func UpdateState(s *server.Server, connId uint32, data []byte) error {
	mu.Lock()
	defer mu.Unlock()
	if curTeam.Id != connId {
		return nil
	}

	x := int(data[0])
	y := int(data[1])

	if state[x][y] == 0 {
		state[x][y] = int(curTeam.TeamId)
	} else {
		stateData := make([]byte, 10)
		stateData[0] = uint8(curTeam.TeamId)
		for i := 1; i < len(stateData); i++ {
			x = (i - 1) % 3
			y = (i - 1) / 3
			stateData[i] = uint8(state[x][y])
		}
		return s.BroadcastSafe(UPDATE_STATE, stateData)
	}

	if curTeam == teamA {
		curTeam = teamB
	} else {
		curTeam = teamA
	}

	stateData := make([]byte, 11)
	stateData[0] = uint8(curTeam.TeamId)
	stateData[1] = uint8(checkWin(state))
	for i := 2; i < len(stateData); i++ {
		x = (i - 2) % 3
		y = (i - 2) / 3
		stateData[i] = uint8(state[x][y])
	}

	s.BroadcastFast(4, []byte("hello, world"))
	return s.BroadcastSafe(UPDATE_STATE, stateData)
}

func checkWin(board [][]int) int {
	for i := range 3 {
		// Check row
		if board[i][0] != 0 && board[i][0] == board[i][1] && board[i][0] == board[i][2] {
			return board[i][0]
		}
		// Check column
		if board[0][i] != 0 && board[0][i] == board[1][i] && board[0][i] == board[2][i] {
			return board[0][i]
		}
	}

	// Check diagonals
	if board[0][0] != 0 && board[0][0] == board[1][1] && board[0][0] == board[2][2] {
		return board[0][0]
	}
	if board[0][2] != 0 && board[0][2] == board[1][1] && board[0][2] == board[2][0] {
		return board[0][2]
	}

	// No winner
	return 0
}

func OnConn(s *server.Server, connId uint32) {
	mu.Lock()
	defer mu.Unlock()
	if teamA == nil {
		teamA = &Client{
			Id:     connId,
			TeamId: 1,
		}
	} else if teamB == nil {
		teamB = &Client{
			Id:     connId,
			TeamId: 2,
		}
	} else {
		// Cick player
		return
	}

	if teamA != nil && teamB != nil {
		for x := range state {
			for y := range state[x] {
				state[x][y] = 0
			}
		}

		dataA := []byte{byte(teamA.TeamId)} // true
		_ = s.SendToClientSafe(teamA.Id, SET_TEAM, dataA)

		dataB := []byte{byte(teamB.TeamId)} // true
		_ = s.SendToClientSafe(teamB.Id, SET_TEAM, dataB)

		curTeam = teamA
	}
}

func OnDisConn(s *server.Server, connId uint32) {
	mu.Lock()
	if teamA.Id == connId {
		teamA = nil
	} else if teamB.Id == connId {
		teamB = nil
	}
	mu.Unlock()
}
