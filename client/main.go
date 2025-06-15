package main

import (
	"jogo/comum"
	"log"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// ClienteJogo gerencia todo o estado e a lógica do cliente.
type ClienteJogo struct {
	cliente         *rpc.Client
	idJogador       int
	estado          comum.EstadoJogo
	mutex           sync.RWMutex
	numeroSequencia int64
}

func (c *ClienteJogo) desenhar() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.estado.Mapa == nil {
		return
	}

	for y, linha := range c.estado.Mapa {
		for x, elemento := range linha {
			termbox.SetCell(x, y, elemento.Simbolo, elemento.Cor, elemento.CorFundo)
		}
	}

	for _, jogador := range c.estado.Jogadores {
		termbox.SetCell(jogador.X, jogador.Y, jogador.Icone.Simbolo, jogador.Icone.Cor, jogador.Icone.CorFundo)
	}
	termbox.Flush()
}

func (c *ClienteJogo) buscarEstadoPeriodicamente() {
	for {
		var resposta comum.RespostaEstado
		err := c.cliente.Call("ServidorJogo.ObterEstado", &comum.ArgsEstado{}, &resposta)
		if err != nil {
			log.Printf("Erro ao buscar estado: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		c.mutex.Lock()
		c.estado = resposta.Estado
		c.mutex.Unlock()
		c.desenhar()
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *ClienteJogo) enviarPingPeriodicamente() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		args := &comum.ArgsPing{IDJogador: c.idJogador}
		var resposta comum.RespostaPing
		err := c.cliente.Call("ServidorJogo.Ping", args, &resposta)
		if err != nil {
			log.Println("Falha ao enviar ping, a conexão com o servidor pode ter sido perdida.")
			return
		}
	}
}

func (c *ClienteJogo) mover(direcao rune) {
	c.numeroSequencia++
	args := &comum.ArgsMovimento{
		IDJogador:       c.idJogador,
		NumeroSequencia: c.numeroSequencia,
		Direcao:         direcao,
	}
	var resposta comum.RespostaMovimento
	err := c.cliente.Call("ServidorJogo.Mover", args, &resposta)
	if err != nil {
		log.Printf("Erro ao mover: %v", err)
	}
}

func main() {
	enderecoServidor := "localhost:12345"
	if len(os.Args) > 1 {
		enderecoServidor = os.Args[1]
	}

	clienteRPC, err := rpc.Dial("tcp", enderecoServidor)
	if err != nil {
		log.Fatalf("Falha ao conectar ao servidor em %s: %v", enderecoServidor, err)
	}

	var respostaConexao comum.RespostaConexao
	err = clienteRPC.Call("ServidorJogo.Conectar", &comum.ArgsConexao{}, &respostaConexao)
	if err != nil {
		log.Fatalf("Falha ao registrar no servidor: %v", err)
	}

	clienteJogo := &ClienteJogo{
		cliente:   clienteRPC,
		idJogador: respostaConexao.IDJogador,
		estado:    respostaConexao.Estado,
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	go clienteJogo.buscarEstadoPeriodicamente()
	go clienteJogo.enviarPingPeriodicamente()

	clienteJogo.desenhar()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc {
				return
			}
			switch ev.Ch {
			case 'w', 'a', 's', 'd':
				clienteJogo.mover(ev.Ch)
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}