package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var idVeiculo int
var selecionado int

func main() {
	// capturar sinal em caso do cliente se desconectar
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	resp, erro := http.Get("http://localhost:8080/conectar")

	if erro != nil {
		fmt.Println("F TOTAL")
	}

	defer resp.Body.Close()
	fmt.Println("Conectado!")

	// lidar com a saída em caso de ctrl + c
	go func() {
		<-signalChan
		desconectarDoServidor()
		os.Exit(0)
	}()
	
	selecionarObjetivo()
}

func selecionarObjetivo() {
	fmt.Println("Digite o ID do veículo a ser importado:")
	fmt.Scanln(&idVeiculo)

	for {
		fmt.Println("Digite 1 se deseja reservar vaga em um posto")
		fmt.Println("Digite 2 para desconectar")
		fmt.Scanln(&selecionado)

		if selecionado == 1 {
			fmt.Println("Reserva enviada!")
		} else if selecionado == 2 {
			desconectarDoServidor()
			break
		}
	}
}

func desconectarDoServidor() {
	resp, erro := http.Get("http://localhost:8080/desconectar")

	if erro != nil {
		fmt.Println("F TOTAL")
		return
	}

	defer resp.Body.Close()
	fmt.Println("Desconectado!")
}
