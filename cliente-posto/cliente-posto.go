package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"io"
	"net"
	//"time"
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

var posto_criado *modelo.Posto
var conexao net.Conn

func main() {
	var erro error
	conexao, erro = net.Dial("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao conectar ao servidor:", erro)
		return
	}

	defer conexao.Close()

	fmt.Println("posto conectado à porta:", conexao.RemoteAddr())
	posto := modelo.NovoPosto("teste", 10, 10)
	posto_criado = &posto
	selecionarObjetivo()
}

func enviarRequisicao(req Requisicao) error {
	dados, erro := json.Marshal(req)
	if erro != nil {
		fmt.Println("Erro ao codificar a requisição")
		return erro
	}

	_, erro = conexao.Write(dados)

	if erro != nil {
		fmt.Println("Erro ao enviar a requisição")
		return erro
	}

	return nil
}

func receberResposta() json.RawMessage {
	
	buffer := make([]byte, 4096)

	n, erro := conexao.Read(buffer)
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
		return response.Dados
	case "get-posto":
		return response.Dados
	}

	return nil
}

func retornarConexaoPosto(){
	for {
		resp := receberResposta()
		if resp == nil {
			fmt.Println("Erro ao listar veículos")
			continue
		}
			
		var tipo string
		erro := json.Unmarshal(resp, &tipo)
		if erro != nil {
			fmt.Println("Erro ao converter JSON:", erro)
			continue
		}
		if tipo == "tipo-cliente"{
			req := Requisicao{
				Comando: "adicionar-conexao",
				Dados: json.RawMessage(`"posto"`),
			}
			enviarRequisicao(req)
			break
		}
	}
}

func selecionarObjetivo() {
	retornarConexaoPosto()
	buffer := make([]byte, 4096)
	for {		
		n, erro := conexao.Read(buffer)
		if erro != nil {
			if erro == io.EOF {
				fmt.Printf("O cliente %s fechou a conexão\n", conexao.RemoteAddr())
			}
			break
		}
		var req Requisicao
		var resposta Requisicao
		erro = json.Unmarshal(buffer[:n], &req)

		if erro != nil {
			fmt.Println("Erro ao decodificar a requisição")
			continue
		}

		switch req.Comando {
		case "get-posto":
			postoJSON,e := json.Marshal(posto_criado)
			if e != nil{
				fmt.Println("erro")
				return
			}
			fmt.Println("teste1")
			resposta = Requisicao{
				Comando: "get-posto",
				Dados: postoJSON,
			}
			fmt.Println("teste2")
			enviarRequisicao(resposta)
			
		}
		if posto_criado.ID != "" {
			fmt.Printf("O id do posto É: %s. \nLONGITUDE E LATITUDE: %v, %v.\n\n\n", posto_criado.ID, posto_criado.Longitude, posto_criado.Latitude)
		} else {
			fmt.Println("Você ainda não possui um posto.")
		}



		fmt.Printf("Digite 0 para cadastrar seu posto\n")
		fmt.Printf("Digite 1 para listar os postos e importar algum\n")
		
		fmt.Scanln(&opcao)
		switch {
		case opcao == 0:
			fmt.Println("Cadastrar posto")
			cadastrarVeiculo()

		case opcao == 1:
			fmt.Println("Listar e importar posto")
			//listarEImportarVeiculo()	

		default:
			fmt.Println("Opção inválida")
		}
	}
}


func cadastrarPosto() {
	if posto_criado != nil {
		fmt.Println("Digite o ID do posto:")
		fmt.Scanln(&id)
		fmt.Println("Digite a latitude do posto:")
		fmt.Scanln(&latitude)
		fmt.Println("Digite a longitude do posto:")
		fmt.Scanln(&longitude)	

		posto := modelo.NovoPosto(id, longitude, latitude)
		posto_criado = &posto

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

		if erro == nil {
			fmt.Println("Posto cadastrado com sucesso")
		}
	} else {
		fmt.Println("voce ja tem um posto cadastrado")
	}
	
}