Com certeza! Aqui está uma versão atualizada do arquivo `README.md` que reflete a nova estrutura cliente-servidor do projeto.

---

# Jogo de Terminal Multiplayer em Go

Este projeto é um jogo multiplayer que roda no terminal, desenvolvido em Go. Ele utiliza um servidor central para gerenciar o estado do jogo e múltiplos clientes que se conectam para interagir no mesmo mapa. A comunicação é feita via RPC (Remote Procedure Call) e a interface do cliente usa a biblioteca [termbox-go](https://github.com/nsf/termbox-go).

## Arquitetura

O projeto foi reestruturado em um modelo cliente-servidor para suportar múltiplos jogadores.

- **Servidor (`/server`):** Uma aplicação de console responsável por gerenciar o estado do jogo, como a posição de todos os jogadores e as regras do mapa. Ele não possui interface gráfica.
- **Cliente (`/client`):** A aplicação que o jogador executa. Ela renderiza o jogo, captura os comandos do teclado e se comunica com o servidor para enviar ações e receber atualizações.
- **Comum (`/common`):** Um pacote compartilhado que contém as estruturas de dados usadas na comunicação entre o cliente e o servidor.

## Como funciona

- O servidor é iniciado e carrega um mapa a partir de um arquivo `.txt`.
- Múltiplos clientes podem se conectar ao endereço do servidor.
- O personagem se move com as teclas **W**, **A**, **S**, **D**. Os movimentos são enviados ao servidor e o estado atualizado é recebido por todos os clientes conectados.
- Pressione **ESC** para sair do jogo.

### Controles do Cliente

| Tecla | Ação |
| :--- | :--- |
| **W** | Mover para cima |
| **A** | Mover para a esquerda |
| **S** | Mover para baixo |
| **D** | Mover para a direita |
| **ESC** | Sair do jogo |

## Como Compilar e Rodar

1.  **Pré-requisitos:**
    * Instale o [Go](https://go.dev/doc/install).
    * Clone este repositório para a sua máquina.

2.  **Instalar Dependências:**
    ```bash
    go mod init jogo
    ```
    ```bash
    go get -u github.com/nsf/termbox-go
    ```

3.  **Compilar o Servidor e o Cliente:**
    Execute os seguintes comandos na raiz do projeto para criar os executáveis.
    ```bash
    # Linux
    # Compila o servidor
    go build -o servidor ./server

    # Compila o cliente
    go build -o cliente ./client

    # Windows
    # Compila o servidor
    go build -o servidor.exe ./server
    # Compila o cliente
    go build -o cliente.exe ./client
    ```

4.  **Executar o Jogo:**
    Você precisará de pelo menos dois terminais abertos.

    * **Terminal 1: Inicie o Servidor**
        ```bash
        ./servidor
        ```
        O servidor começará a rodar e aguardar por conexões na porta `12345`.

    * **Terminal 2 (e outros): Inicie o Cliente**
        Para conectar em um servidor rodando na sua própria máquina:
        ```bash
        ./cliente
        ```
        Para conectar em um servidor rodando em outra máquina na rede (substitua pelo IP do servidor):
        ```bash
        ./cliente <ip_do_servidor>:12345
        ```

## Estrutura do Projeto

- **`server/main.go`**: Ponto de entrada e lógica principal do servidor do jogo.
- **`client/main.go`**: Ponto de entrada e lógica principal do cliente (UI, input, etc).
- **`common/types.go`**: Contém as estruturas de dados compartilhadas via RPC.
- **`go.mod`**: Arquivo de definição do módulo e suas dependências.
- **`mapa.txt`**: Arquivo de exemplo do mapa do jogo.