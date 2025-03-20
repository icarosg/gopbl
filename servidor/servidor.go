package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"net/http"
	"log"
)

var (
	qtdClientesConectados int
	mutex                 sync.Mutex
)

func main() {
	// Inicializa os postos
	inicializar()

	// Configura as rotas HTTP
	http.HandleFunc("/posto", handler)
	http.HandleFunc("/listar", ListarPostos)
	http.HandleFunc("/cadastrar-veiculo", cadastrarVeiculo)

	// Cria um listener TCP na porta 8080
	listener, erro := net.Listen("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor iniciado em localhost:8080")

	// Inicia o servidor HTTP no listener TCP
	go func() {
		log.Fatal(http.Serve(listener, nil))
	}()

	// Mantém o servidor principal em execução
	for {
		conexao, erro := listener.Accept()
		if erro != nil {
			fmt.Println("Erro ao conectar o cliente", erro)
			continue
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
	}()

	buffer := make([]byte, 1024)
	for {
		_, erro := conexao.Read(buffer)
		if erro != nil {
			if erro == io.EOF {
				fmt.Printf("O cliente %s fechou a conexão\n", conexao.RemoteAddr())
			}
			break
		}
	}
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