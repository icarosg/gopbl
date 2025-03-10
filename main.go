package main

import (
	"fmt"
	"gopbl/modelo"
)

func main() {
	novoVeiculo := modelo.NovoVeiculo("tesla", 200, 240)
	
	novoPosto := modelo.NovoPosto("Ipiranga", 300, 500)

	fmt.Println("\nVeiculo criado: ", novoVeiculo)
	fmt.Println("Posto criado:", novoPosto)
}