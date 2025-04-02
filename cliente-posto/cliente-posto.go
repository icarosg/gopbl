package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"net"
	//"time"
	"sync"
)

type Requisicao struct {
	Comando string          `json:"comando"`
	Dados   json.RawMessage `json:"dados"`
}

var opcao int
var (
	id        string
	latitude  float64
	longitude float64 //bateria   float64
)

var posto_criado modelo.Posto
var conexao net.Conn
var mutex sync.Mutex

func main() {
	var erro error
	conexao, erro = net.Dial("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao conectar ao servidor:", erro)
		return
	}

	defer conexao.Close()

	fmt.Println("posto conectado à porta:", conexao.RemoteAddr())
	posto_criado = modelo.NovoPosto("teste", 10, 10)

	selecionarObjetivo()
	// 	go verificarAlgoNoBuffer()
}

func enviarRequisicao(req Requisicao) error {
	dados, erro := json.Marshal(req)
	if erro != nil {
		fmt.Println("Erro ao codificar a requisição")
		return erro
	}

	mutex.Lock()
	_, erro = conexao.Write(dados)
	mutex.Unlock()

	fmt.Println("enviado")

	if erro != nil {
		fmt.Println("Erro ao enviar a requisição")
		return erro
	}

	return nil
}

func receberResposta() *Requisicao {
	buffer := make([]byte, 4096)

	mutex.Lock()
	n, erro := conexao.Read(buffer)
	mutex.Unlock()

	if erro != nil {
		fmt.Println("Erro ao receber a resposta")
		return nil
	}

	var response Requisicao
	erro = json.Unmarshal(buffer[:n], &response)
	if erro != nil {
		fmt.Println("Erro ao decodificar a resposta")
	}

	switch response.Comando {
	case "tipo-cliente":
		return &response
	case "get-posto":
		return &response
	}

	return nil
}

func retornarConexaoPosto() {
	for {
		resp := receberResposta()
		if resp == nil {
			fmt.Println("Erro ao retornar conexão do posto")
			continue
		}

		var tipo string
		erro := json.Unmarshal(resp.Dados, &tipo)
		if erro != nil {
			fmt.Println("Erro ao converter JSON:", erro)
			continue
		}
		if tipo == "tipo-cliente" {
			req := Requisicao{
				Comando: "adicionar-conexao",
				Dados:   json.RawMessage(`"posto"`),
			}
			enviarRequisicao(req)
			break
		}
	}
}

func verificarAlgoNoBuffer() {
	for {
		resposta := receberResposta()

		if (resposta) != nil {
			if posto_criado.ID != "" {
				var req Requisicao

				switch resposta.Comando {
				case "get-posto":
					postoJSON, erro := json.Marshal(posto_criado)
					if erro != nil {
						fmt.Println("erro")
						continue
					}
					req = Requisicao{
						Comando: "get-posto",
						Dados:   postoJSON,
					}
					enviarRequisicao(req)
				}

			} else {
				fmt.Println("Você ainda não possui um posto.")
			}
		}
	}
}

func selecionarObjetivo() {
	retornarConexaoPosto()
	go verificarAlgoNoBuffer()

	for {
		fmt.Printf("Digite 0 para cadastrar seu posto\n")
		fmt.Printf("Digite 1 para listar os postos e importar algum\n")

		fmt.Scanln(&opcao)
		switch {
		case opcao == 0:
			fmt.Println("Cadastrar posto")
			cadastrarPosto()

		case opcao == 1:
			fmt.Println("Listar e importar posto")
			//listarEImportarVeiculo()

		default:
			fmt.Println("Opção inválida")
		}
	}
}

func cadastrarPosto() {
	fmt.Println("Digite o ID do posto:")
	fmt.Scanln(&id)
	fmt.Println("Digite a latitude do posto:")
	fmt.Scanln(&latitude)
	fmt.Println("Digite a longitude do posto:")
	fmt.Scanln(&longitude)

	posto_criado = modelo.NovoPosto(id, longitude, latitude)
	// posto_criado = &posto

	fmt.Println("Posto cadastrado com sucesso")
}
