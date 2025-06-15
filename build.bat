@echo off
REM build.bat - Script de build e configuracao inicial para Windows

ECHO.
ECHO =========================================
ECHO      INICIANDO BUILD DO JOGO MULTIPLAYER
ECHO =========================================
ECHO.

REM Inicializa o modulo go e baixa as dependencias, apenas se nao existir.
IF NOT EXIST go.mod (
    echo O arquivo go.mod nao foi encontrado.
    echo Inicializando o modulo 'jogo'...
    go mod init jogo
    ECHO.
    echo Baixando dependencias...
    go get -u github.com/nsf/termbox-go
) ELSE (
    echo O arquivo go.mod ja existe. Verificando dependencias...
    go mod tidy
)
ECHO.

REM Compila o servidor
echo Compilando o servidor...
go build -o servidor.exe ./server
IF %ERRORLEVEL% NEQ 0 (
    echo.
    echo *** FALHA AO COMPILAR O SERVIDOR! ***
    GOTO:EOF
)
echo Servidor compilado com sucesso como servidor.exe
ECHO.

REM Compila o cliente
echo Compilando o cliente...
go build -o cliente.exe ./client
IF %ERRORLEVEL% NEQ 0 (
    echo.
    echo *** FALHA AO COMPILAR O CLIENTE! ***
    GOTO:EOF
)
echo Cliente compilado com sucesso como cliente.exe
ECHO.

ECHO =========================================
ECHO      BUILD CONCLUIDO COM SUCESSO!
ECHO =========================================
ECHO.
echo Para iniciar, execute servidor.exe em um terminal e cliente.exe em outro.

