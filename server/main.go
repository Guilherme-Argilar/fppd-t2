package main

import (
	"bufio"
	"fmt"
	"log"
	"jogo/common"
	"net"
	"net/rpc"
	"os"
	"sync"

	"github.com/nsf/termbox-go"
)

var (
	Parede    = common.Elemento{'▤', termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim, termbox.ColorDarkGray, true}
	Vegetacao = common.Elemento{'♣', termbox.ColorGreen, termbox.ColorDefault, false}
	Vazio     = common.Elemento{' ', termbox.ColorDefault, termbox.ColorDefault, false}
)

type GameServer struct {
	mutex            sync.Mutex
	state            common.GameState
	nextPlayerID     int
	lastProcessedCmd map[int]int64
}

func NewGameServer(mapFile string) *GameServer {
	server := &GameServer{
		state: common.GameState{
			Players: make(map[int]common.Player),
		},
		nextPlayerID:     1,
		lastProcessedCmd: make(map[int]int64),
	}
	if err := server.loadMap(mapFile); err != nil {
		log.Fatalf("Falha ao carregar o mapa: %v", err)
	}
	return server
}

func (s *GameServer) loadMap(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var row []common.Elemento
		for _, char := range scanner.Text() {
			var elem common.Elemento
			switch char {
			case Parede.Simbolo:
				elem = Parede
			case Vegetacao.Simbolo:
				elem = Vegetacao
			default:
				elem = Vazio
			}
			row = append(row, elem)
		}
		s.state.Mapa = append(s.state.Mapa, row)
	}
	return scanner.Err()
}

func (s *GameServer) Connect(args *common.ConnectArgs, reply *common.ConnectReply) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	playerID := s.nextPlayerID
	s.nextPlayerID++

	player := common.Player{
		ID:    playerID,
		X:     2,
		Y:     12,
		Icono: common.Elemento{'☺', termbox.ColorWhite, termbox.ColorDefault, true},
	}
	s.state.Players[playerID] = player
	s.lastProcessedCmd[playerID] = 0

	reply.PlayerID = playerID
	reply.State = s.state

	log.Printf("Jogador %d conectado.", playerID)
	return nil
}

func (s *GameServer) Move(args *common.MoveArgs, reply *common.MoveReply) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if args.SequenceNumber <= s.lastProcessedCmd[args.PlayerID] {
		reply.Success = true
		return nil
	}

	player, ok := s.state.Players[args.PlayerID]
	if !ok {
		reply.Success = false
		return fmt.Errorf("jogador com ID %d não encontrado", args.PlayerID)
	}

	dx, dy := 0, 0
	switch args.Direction {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}

	nx, ny := player.X+dx, player.Y+dy

	if s.canMoveTo(nx, ny) {
		player.X = nx
		player.Y = ny
		s.state.Players[args.PlayerID] = player
		s.lastProcessedCmd[args.PlayerID] = args.SequenceNumber
		reply.Success = true
	} else {
		reply.Success = false
	}

	return nil
}

func (s *GameServer) GetState(args *common.GetStateArgs, reply *common.GetStateReply) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	reply.State = s.state
	return nil
}

func (s *GameServer) canMoveTo(x, y int) bool {
	if y < 0 || y >= len(s.state.Mapa) || x < 0 || x >= len(s.state.Mapa[y]) {
		return false
	}
	if s.state.Mapa[y][x].Tangivel {
		return false
	}
	for _, p := range s.state.Players {
		if p.X == x && p.Y == y {
			return false
		}
	}
	return true
}

func main() {
	mapFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapFile = os.Args[1]
	}

	server := NewGameServer(mapFile)
	rpc.Register(server)

	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
	defer listener.Close()

	log.Println("Servidor de jogo iniciado na porta 12345")
	rpc.Accept(listener)
}