package main

import (
	"fmt"
	"net"
)

var idVeiculo int = 0
var selecionado int

func main() {
	conexao, erro := net.Dial("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao conectar ao servidor:", erro)
		return
	}

	defer conexao.Close()

	fmt.Println("Veículo conectado à porta:", conexao.RemoteAddr())

	for {
		if idVeiculo == 0 {
			fmt.Println("Digite o ID do veículo a ser importado")
			fmt.Scanln(&idVeiculo)
		} else {
			fmt.Println("Digite 1 se deseja reservar vaga em um posto")
			fmt.Scanln(&selecionado)

			if selecionado == 1 {
				fmt.Println("Ok")
			}
		}
	}
}
