package main

import (
	"bufio"
	"fmt"
	"jogo/comum"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	Parede    = comum.Elemento{Simbolo: '▤', Cor: termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim, CorFundo: termbox.ColorDarkGray, Tangivel: true}
	Vegetacao = comum.Elemento{Simbolo: '♣', Cor: termbox.ColorGreen, CorFundo: termbox.ColorDefault, Tangivel: false}
	Vazio     = comum.Elemento{Simbolo: ' ', Cor: termbox.ColorDefault, CorFundo: termbox.ColorDefault, Tangivel: false}
)


type ServidorJogo struct {
	mutex               sync.Mutex
	estado              comum.EstadoJogo
	proximoIDJogador    int
	ultimoCmdProcessado map[int]int64
	ultimoPing          map[int]time.Time
}


func NovoServidorJogo(arquivoMapa string) *ServidorJogo {
	servidor := &ServidorJogo{
		estado: comum.EstadoJogo{
			Jogadores: make(map[int]comum.Jogador),
		},
		proximoIDJogador:    1,
		ultimoCmdProcessado: make(map[int]int64),
		ultimoPing:          make(map[int]time.Time),
	}
	if err := servidor.carregarMapa(arquivoMapa); err != nil {
		log.Fatalf("Falha ao carregar o mapa: %v", err)
	}
	return servidor
}

func (s *ServidorJogo) carregarMapa(nomeArquivo string) error {
	arquivo, err := os.Open(nomeArquivo)
	if err != nil {
		return err
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	log.Printf("Carregando mapa do arquivo: %s", nomeArquivo)
	for scanner.Scan() {
		var linha []comum.Elemento
		for _, caractere := range scanner.Text() {
			var elemento comum.Elemento
			switch caractere {
			case Parede.Simbolo:
				elemento = Parede
			case Vegetacao.Simbolo:
				elemento = Vegetacao
			default:
				elemento = Vazio
			}
			linha = append(linha, elemento)
		}
		s.estado.Mapa = append(s.estado.Mapa, linha)
	}
	return scanner.Err()
}

func (s *ServidorJogo) Conectar(args *comum.ArgsConexao, resposta *comum.RespostaConexao) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	idJogador := s.proximoIDJogador
	s.proximoIDJogador++

	jogador := comum.Jogador{
		ID:    idJogador,
		X:     2,
		Y:     12,
		Icone: comum.Elemento{Simbolo: '☺', Cor: termbox.ColorWhite, CorFundo: termbox.ColorDefault, Tangivel: true},
	}
	s.estado.Jogadores[idJogador] = jogador
	s.ultimoCmdProcessado[idJogador] = 0
	s.ultimoPing[idJogador] = time.Now()

	resposta.IDJogador = idJogador
	resposta.Estado = s.estado

	log.Printf("EVENTO: Jogador %d conectou.", idJogador)
	return nil
}

func (s *ServidorJogo) Desconectar(args *comum.ArgsDesconexao, resposta *comum.RespostaDesconexao) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	id := args.IDJogador

	if _, ok := s.estado.Jogadores[id]; ok {
		log.Printf("EVENTO: Jogador %d desconectou corretamente.", id)
		delete(s.estado.Jogadores, id)
		delete(s.ultimoPing, id)
		delete(s.ultimoCmdProcessado, id)
	}

	return nil
}


func (s *ServidorJogo) Mover(args *comum.ArgsMovimento, resposta *comum.RespostaMovimento) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if args.NumeroSequencia <= s.ultimoCmdProcessado[args.IDJogador] {
		log.Printf("AÇÃO: Comando duplicado ou antigo do Jogador %d (seq: %d) ignorado.", args.IDJogador, args.NumeroSequencia)
		resposta.Sucesso = true
		return nil
	}

	jogador, ok := s.estado.Jogadores[args.IDJogador]
	if !ok {
		resposta.Sucesso = false
		return fmt.Errorf("jogador com ID %d não encontrado", args.IDJogador)
	}

	dx, dy := 0, 0
	switch args.Direcao {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}

	nx, ny := jogador.X+dx, jogador.Y+dy

	if s.podeMoverPara(nx, ny) {
		log.Printf("AÇÃO: Jogador %d moveu-se para (%d, %d).", args.IDJogador, nx, ny)
		jogador.X = nx
		jogador.Y = ny
		s.estado.Jogadores[args.IDJogador] = jogador
		s.ultimoCmdProcessado[args.IDJogador] = args.NumeroSequencia
		resposta.Sucesso = true
	} else {
		log.Printf("AÇÃO: Movimento do Jogador %d para (%d, %d) bloqueado.", args.IDJogador, nx, ny)
		resposta.Sucesso = false
	}

	return nil
}

func (s *ServidorJogo) ObterEstado(args *comum.ArgsEstado, resposta *comum.RespostaEstado) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	resposta.Estado = s.estado
	return nil
}

func (s *ServidorJogo) Ping(args *comum.ArgsPing, resposta *comum.RespostaPing) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.estado.Jogadores[args.IDJogador]; ok {
		s.ultimoPing[args.IDJogador] = time.Now()
	}
	return nil
}

func (s *ServidorJogo) podeMoverPara(x, y int) bool {
	if y < 0 || y >= len(s.estado.Mapa) || x < 0 || x >= len(s.estado.Mapa[y]) {
		return false
	}
	if s.estado.Mapa[y][x].Tangivel {
		return false
	}
	for _, j := range s.estado.Jogadores {
		if j.X == x && j.Y == y {
			return false
		}
	}
	return true
}

func (s *ServidorJogo) verificarJogadoresAtivos() {
	for {
		time.Sleep(5 * time.Second)

		s.mutex.Lock()
		for id, ultimoPing := range s.ultimoPing {
			if time.Since(ultimoPing) > 15*time.Second {
				log.Printf("EVENTO: Jogador %d desconectado por inatividade.", id)
				delete(s.estado.Jogadores, id)
				delete(s.ultimoPing, id)
				delete(s.ultimoCmdProcessado, id)
			}
		}
		s.mutex.Unlock()
	}
}

func main() {
	arquivoMapa := "mapa.txt"
	if len(os.Args) > 1 {
		arquivoMapa = os.Args[1]
	}

	servidor := NovoServidorJogo(arquivoMapa)
	rpc.Register(servidor)

	go servidor.verificarJogadoresAtivos()

	escuta, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
	defer escuta.Close()

	log.Println("Servidor de jogo iniciado na porta 12345")
	rpc.Accept(escuta)
}