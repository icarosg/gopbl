package api

// definir rotas para: consulta de pontos de recarga disponíveis;
// reserva de pontos de recarga;
// registro de recargas realizadas.

import (
	"fmt"
	"net"
)

func Route() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conn.Close()

	message := "Olá, servidor!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Erro ao enviar mensagem:", err)
		return
	}
	fmt.Println("Mensagem enviada:", message)
}
