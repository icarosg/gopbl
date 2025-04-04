package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"net"
	"time"

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
	// posto := modelo.NovoPosto("teste", 10, 10)
	// posto_criado = &posto
	//retornarConexaoPosto()	
	//go verificarAlgoNoBuffer()
	selecionarObjetivo()
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
	// Configura um timeout de 100ms
    conexao.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	mutex.Lock()
	n, erro := conexao.Read(buffer)
	mutex.Unlock()

	if erro != nil {
		//fmt.Println("Erro ao receber a resposta")
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
	case "cadastrou":
		return &response
	default:
		return nil
	}

	//return nil
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
	//fmt.Println("verificando algo no buffer")
	for {
		//fmt.Println("to aki oh")
		resposta := receberResposta()
		//fmt.Println("sexo")
		if resposta != nil {
			fmt.Println(resposta.Comando)
		}

		if resposta != nil{
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
						Comando: "listar-postos",
						Dados:   postoJSON,
					}
					enviarRequisicao(req)
				// case "cadastrar-posto":
				// 	cadastrarPosto()
				}

			} else {
				fmt.Println("Você ainda não possui um posto.")
			}
			//return
		}
		
	}
}

func selecionarObjetivo() {
	retornarConexaoPosto()
	go verificarAlgoNoBuffer()

	for {
		//verificarAlgoNoBuffer()
		fmt.Printf("Digite 0 para cadastrar seu posto\n")
		fmt.Printf("Digite 1 para listar os postos e importar algum\n")

		fmt.Scanln(&opcao)
		switch {
		case opcao == 0:				
			cadastrarPosto()

		case opcao == 1:
			fmt.Println("Listar e importar posto")
			//listarEImportarVeiculo()

		default:
			fmt.Println("Opção inválida")
		}
		//verificarAlgoNoBuffer()
	}
}

func cadastrarPosto() {	
	
		fmt.Println("Cadastrar posto")
		fmt.Println("Digite o ID do posto:")
		fmt.Scanln(&id)
		fmt.Println("Digite a latitude do posto:")
		fmt.Scanln(&latitude)
		fmt.Println("Digite a longitude do posto:")
		fmt.Scanln(&longitude)	

		posto := modelo.NovoPosto(id, longitude, latitude)
		posto_criado = posto

		postoJSON, erro := json.Marshal(posto_criado)
		if erro != nil {
			fmt.Printf("Erro ao converter posto para JSON: %v\n", erro)
			return
		}

		req := Requisicao{
			Comando: "cadastrar-posto",
			Dados:   postoJSON,
		}
		
		erro = enviarRequisicao(req)		

		if erro != nil {
			fmt.Println("erro ao cadastrar posto")
			return
		}

		err := receberResposta()
		if err != nil {
			fmt.Println("erro ao receber respsta depois de tentar cadastrar posto")
			return
		}



	} 