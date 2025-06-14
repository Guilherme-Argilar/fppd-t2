package main

import (
	"log"
	"jogo/common"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

type GameClient struct {
	client         *rpc.Client
	playerID       int
	state          common.GameState
	mutex          sync.RWMutex
	sequenceNumber int64
}

func (c *GameClient) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.state.Mapa == nil {
		return
	}

	for y, row := range c.state.Mapa {
		for x, elem := range row {
			termbox.SetCell(x, y, elem.Simbolo, elem.Cor, elem.CorFundo)
		}
	}

	for _, player := range c.state.Players {
		termbox.SetCell(player.X, player.Y, player.Icono.Simbolo, player.Icono.Cor, player.Icono.CorFundo)
	}
	termbox.Flush()
}

func (c *GameClient) pollState() {
	for {
		var reply common.GetStateReply
		err := c.client.Call("GameServer.GetState", &common.GetStateArgs{}, &reply)
		if err != nil {
			log.Printf("Erro ao buscar estado: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		c.mutex.Lock()
		c.state = reply.State
		c.mutex.Unlock()
		c.draw()
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *GameClient) move(direction rune) {
	c.sequenceNumber++
	args := &common.MoveArgs{
		PlayerID:       c.playerID,
		SequenceNumber: c.sequenceNumber,
		Direction:      direction,
	}
	var reply common.MoveReply
	err := c.client.Call("GameServer.Move", args, &reply)
	if err != nil {
		log.Printf("Erro ao mover: %v", err)
	}
}

func main() {
	serverAddress := "localhost:12345"
	if len(os.Args) > 1 {
		serverAddress = os.Args[1]
	}

	client, err := rpc.Dial("tcp", serverAddress)
	if err != nil {
		log.Fatalf("Falha ao conectar ao servidor em %s: %v", serverAddress, err)
	}

	var connReply common.ConnectReply
	err = client.Call("GameServer.Connect", &common.ConnectArgs{}, &connReply)
	if err != nil {
		log.Fatalf("Falha ao registrar no servidor: %v", err)
	}

	gameClient := &GameClient{
		client:   client,
		playerID: connReply.PlayerID,
		state:    connReply.State,
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	go gameClient.pollState()

	gameClient.draw()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return
			}
			switch ev.Ch {
			case 'w', 'a', 's', 'd':
				gameClient.move(ev.Ch)
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}