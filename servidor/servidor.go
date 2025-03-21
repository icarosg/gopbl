package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	qtdClientesConectados int
	mutex                 sync.Mutex
)

func main() {
	http.HandleFunc("/conectar", conexao)
	http.HandleFunc("/desconectar", desconectar)

	fmt.Println("Servidor HTTP iniciado em http://localhost:8080")
	erro := http.ListenAndServe("localhost:8080", nil)
	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
}

func conexao(w http.ResponseWriter, r *http.Request) {
	incrementar()

	// exibe em qual porta o cliente foi conectado
	fmt.Println("Cliente conectado:", r.RemoteAddr)
	fmt.Println("Total de clientes conectados:", getQtdClientes())

	// responde o cliente
	_, erro := fmt.Fprintf(w, "Conectado ao servidor! Total de clientes conectados: %d", getQtdClientes())
	if erro != nil {
		fmt.Println("Erro ao responder ao cliente:", erro)
	}
}

func desconectar(w http.ResponseWriter, r *http.Request) {
	decrementar()

	fmt.Println("Cliente desconectado:", r.RemoteAddr)
	fmt.Println("Total de clientes conectados:", getQtdClientes())

	// responde o cliente
	_, erro := fmt.Fprintf(w, "Desconectado do servidor! Total de clientes conectados: %d", getQtdClientes())
	if erro != nil {
		fmt.Println("Erro ao responder ao cliente:", erro)
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
