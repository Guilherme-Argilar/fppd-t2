package comum

import "github.com/nsf/termbox-go"

type Cor = termbox.Attribute

type Elemento struct {
	Simbolo  rune
	Cor      Cor
	CorFundo Cor
	Tangivel bool
}

type Jogador struct {
	ID    int
	X, Y  int
	Icone Elemento
}

type EstadoJogo struct {
	Mapa      [][]Elemento
	Jogadores map[int]Jogador
}

type ArgsConexao struct{}

type RespostaConexao struct {
	IDJogador int
	Estado    EstadoJogo
}

type ArgsMovimento struct {
	IDJogador       int
	NumeroSequencia int64
	Direcao         rune
}

type RespostaMovimento struct {
	Sucesso bool
}

type ArgsEstado struct{}

type RespostaEstado struct {
	Estado EstadoJogo
}

type ArgsPing struct {
	IDJogador int
}

type RespostaPing struct{}

type ArgsDesconexao struct {
	IDJogador int
}

type RespostaDesconexao struct{}