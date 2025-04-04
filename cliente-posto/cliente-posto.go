package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"net"
	"sync"
	"time"
)

type Requisicao struct {
	Comando string          `json:"comando"`
	Dados   json.RawMessage `json:"dados"`
}

var (
	id        string
	latitude  float64
	longitude float64 //bateria   float64
)

var opcao int

var posto_criado modelo.Posto
var conexao net.Conn
var mutex sync.Mutex
var buffer = make([]byte, 4096)

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

	if erro != nil {
		fmt.Println("Erro ao enviar a requisição")
		return erro
	}

	fmt.Println("enviado")

	return nil
}

func receberResposta() *Requisicao {
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
	case "reservar-vaga":
		return &response
	case "atualizar-posicao-veiculo":
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
						Comando: "listar-postos",
						Dados:   postoJSON,
					}
					enviarRequisicao(req)

				case "reservar-vaga":
					reservarVaga(resposta.Dados)

				case "atualizar-posicao-veiculo":
					atualizarPosicaoFila(resposta.Dados)
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

func reservarVaga(r json.RawMessage) {
	var dados modelo.ReservarVagaJson
	erro := json.Unmarshal(r, &dados)
	if erro != nil {
		fmt.Println("Erro ao decodificar JSON", erro)
		return
	}

	fmt.Println("dados do veículo", dados.Veiculo)
	modelo.ReservarVaga(&posto_criado, &dados.Veiculo)

	//envia a requisição para o servidor para enviar a resposta para o veículo
	veiculoConexao := modelo.RetornarVagaJson{
		Posto: posto_criado,
	}

	req, err := json.Marshal(veiculoConexao)
	if err != nil {
		fmt.Printf("Erro ao converter veículo e conexão para JSON: %v\n", err)
		return
	}

	res := Requisicao{
		Comando: "reservar-vaga-retornoPosto",
		Dados:   req,
	}
	enviarRequisicao(res)

	time.Sleep(1 * time.Second) // aguarda por 1 segundo
}

func atualizarPosicaoFila(r json.RawMessage) {
	var dados modelo.ReservarVagaJson
	erro := json.Unmarshal(r, &dados)
	if erro != nil {
		fmt.Println("Erro ao decodificar JSON", erro)
		return
	}

	//modelo.ReservarVaga(&posto_criado, &dados.Veiculo)
	modelo.ArrumarPosicaoFila(&posto_criado)

	var veiculoDadosAtualizados modelo.Veiculo

	fmt.Printf("\n\nFila atual do posto:\n")
	for i := range posto_criado.Fila {
		fmt.Printf("Posição %d: ID VEÍCULO: %s", i, posto_criado.Fila[i].ID)

		if &dados.Veiculo.ID == &posto_criado.Fila[i].ID {
			veiculoDadosAtualizados = *posto_criado.Fila[i]
		}
	}

	//envia a requisição para o servidor para enviar a resposta para o veículo
	data := modelo.RetornarAtualizarPosicaoFila{
		Veiculo: veiculoDadosAtualizados,
		Posto:   posto_criado,
	}

	req, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Erro ao converter veículo e para JSON: %v\n", err)
		return
	}

	res := Requisicao{
		Comando: "atualizar-posicao-veiculo",
		Dados:   req,
	}
	enviarRequisicao(res)
}
