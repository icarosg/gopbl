package main

import (
	"fmt"

	"github.com/icarosg/gopbl/modelo"
)

func main() {
	novoVeiculo := modelo.NovoVeiculo("tesla", 200, 240);

	fmt.Println("Veiculo criado: ", novoVeiculo);
}