package common

import "github.com/nsf/termbox-go"

type Cor = termbox.Attribute

type Elemento struct {
	Simbolo  rune
	Cor      Cor
	CorFundo Cor
	Tangivel bool
}

type Player struct {
	ID    int
	X, Y  int
	Icono Elemento
}

type GameState struct {
	Mapa    [][]Elemento
	Players map[int]Player
}

type ConnectArgs struct{}

type ConnectReply struct {
	PlayerID int
	State    GameState
}

type MoveArgs struct {
	PlayerID       int
	SequenceNumber int64
	Direction      rune
}

type MoveReply struct {
	Success bool
}

type GetStateArgs struct{}

type GetStateReply struct {
	State GameState
}