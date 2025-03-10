package main

import (
	"fmt"
	"gopbl/modelo"
)

func main() {
	novoVeiculo := modelo.NovoVeiculo("tesla", 200, 240);

	fmt.Println("Veiculo criado: ", novoVeiculo);
}