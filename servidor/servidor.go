package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	qtdClientesConectados int
	mutex                 sync.Mutex
)

func main() {
	// cria um listener TCP na porta 8080
	listener, erro := net.Listen("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor iniciado em localhost:8080")

	for {
		conexao, erro := listener.Accept()

		if erro != nil {
			fmt.Println("Erro ao conectao o cliente", erro)
			continue // continua aguardando outras conexões
		}

		incrementar()

		fmt.Println("Cliente conectado à porta:", conexao.RemoteAddr())
		fmt.Println("Total de clientes conectados:", getQtdClientes())

		go cliente(conexao)
	}

}

func cliente(conexao net.Conn) {
	defer func() {

		decrementar()

		fmt.Println("Cliente desconectado. Total de clientes conectados:", getQtdClientes())

		conexao.Close()
	}() // decrementa após a conexão ser encerrada
}

func incrementar() {
	mutex.Lock()
	qtdClientesConectados++
	mutex.Unlock()
}

func decrementar() {
	mutex.Lock()
	qtdClientesConectados--
	mutex.Unlock()
}

func getQtdClientes() int {
	mutex.Lock()
	defer mutex.Unlock()
	return qtdClientesConectados
}
