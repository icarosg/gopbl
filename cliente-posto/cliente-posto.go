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
	longitude float64 
)

var opcao int

var posto_criado modelo.Posto
var conexao net.Conn
var mutex sync.Mutex
var respostas = make(chan *Requisicao, 10) // canal com buffer
var buffer = make([]byte, 4096)

func main() {
	var erro error
	conexao, erro = net.Dial("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao conectar ao servidor:", erro)
		return
	}

	defer conexao.Close()
	fmt.Println("****************************************")
	fmt.Println("posto conectado à porta:", conexao.RemoteAddr())
	//posto_criado = modelo.NovoPosto("teste", 10, 10)

	go verificarAlgoNoBuffer()
	selecionarObjetivo()
	//go verificarAlgoNoBuffer()
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

	//fmt.Println("enviado")

	return nil
}

func receberResposta() *Requisicao {
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
	case "reservar-vaga":
		return &response
	case "atualizar-posicao-veiculo":
		return &response
	case "postos-salvos":
		return &response
	case "cadastrou":
		return &response
	}

	return nil
}


func retornarConexaoPosto() {
	for {
		//resp := receberResposta()
		var resp *Requisicao
		timeout := time.After(2 * time.Second)
		select {
		case resp = <-respostas:
			// tudo certo
		case <-timeout:
			fmt.Println("timeout esperando resposta")
			return
		}
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
		//mutex.Lock()
		//fmt.Println("to aki oh")
		resposta := receberResposta()
		if resposta != nil {
			respostas <- resposta

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
			//return
		}
		//mutex.Unlock()
		//time.Sleep(1 * time.Second) // Espera 1 segundo antes de verificar novamente
	}
}

func selecionarObjetivo() {
	retornarConexaoPosto()
	//go verificarAlgoNoBuffer()

	for {
		fmt.Println("****************************************")
		fmt.Printf("Digite 0 para cadastrar seu posto\n")
		fmt.Printf("Digite 1 para listar os postos e importar algum\n")
		fmt.Println("*****************************************")

		opcao = -1

		fmt.Scanln(&opcao)
		switch {
		case opcao == 0:
			fmt.Println("Cadastrar posto")
			cadastrarPosto()

		case opcao == 1:
			fmt.Println("Listar e importar posto")
			listarEImportarPosto()

		default:
			fmt.Println("Opção inválida")
		}
	}
}

func cadastrarPosto() {
	fmt.Println("*****************************************")
	fmt.Println("Digite o ID do posto:")
	fmt.Scanln(&id)
	fmt.Println("Digite a latitude do posto:")
	fmt.Scanln(&latitude)
	fmt.Println("Digite a longitude do posto:")
	fmt.Scanln(&longitude)
	fmt.Println("*****************************************")
	posto_criado = modelo.NovoPosto(id, longitude, latitude)
	// posto_criado = &posto

	fmt.Println("Posto cadastrado com sucesso")

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
	var err *Requisicao
	timeout := time.After(2 * time.Second)

	select {
	case err = <-respostas:
		// tudo certo
	case <-timeout:
		fmt.Println("timeout esperando resposta")
		return
	}

	//err := receberResposta()
	if err != nil {
		fmt.Println("erro ao receber respsta depois de tentar cadastrar posto")
		return
	}

}

func listarEImportarPosto() {
	req := Requisicao{
		Comando: "postos-salvos",
		Dados:   json.RawMessage(`"postos-salvos"`),
	}
	//mutex.Lock()
	erro := enviarRequisicao(req)
	if erro != nil {
		fmt.Println("erro ao enviar requisiçao para listar os postos salvos no arquivo")
		return
	}
	var resposta *Requisicao
	//time.Sleep(2 * time.Second) // Espera 1 segundo para garantir que a resposta seja recebida
	timeout := time.After(2 * time.Second)
	//resposta := receberResposta()
	//mutex.Unlock()
	//mutex.Unlock()
	// fmt.Println("resposta: ", resposta)
	// if resposta == nil {
	// 	fmt.Println("erro ao receber resposta depois de tentar listar os postos salvos no arquivo")
	// 	return
	// }
	select {
	case resposta = <-respostas:
		// tudo certo
	case <-timeout:
		fmt.Println("timeout esperando resposta")
		return
	}
	mapPosto := map[string]modelo.Posto{}
	var postosNoArquivo *[]modelo.Posto
	err := json.Unmarshal(resposta.Dados, &postosNoArquivo)
	if err != nil {
		fmt.Println("erro ao converter resposta dos postos no arquivo para JSON")
	}
	for i := range *postosNoArquivo {
		mapPosto[(*postosNoArquivo)[i].ID] = (*postosNoArquivo)[i]
		fmt.Println("*****************************************")
		fmt.Printf("posto disponivel para importação ID: %s, Latitude: %.2f, Longitude: %.2f\n", (*postosNoArquivo)[i].ID, (*postosNoArquivo)[i].Latitude, (*postosNoArquivo)[i].Longitude)
		fmt.Println("*****************************************")
	}
	var op string
	fmt.Println("Digite o ID do posto que deseja importar:")
	fmt.Scanln(&op)
	postoDesejado, encontrado := mapPosto[op]
	if postoDesejado.ID == "" || !encontrado {
		fmt.Println("Posto não encontrado")
		return
	}
	posto_criado = postoDesejado
	fmt.Println("Posto importado com sucesso!")
	fmt.Println("ID:", posto_criado.ID)	
	// requi := Requisicao{
	// 	Comando: "add-fila",
	// }
	// ej := enviarRequisicao(requi)
	// if ej != nil {
	// 	fmt.Println("erro ao enviar requisição para adicionar o posto na fila")
	// 	return
	// }
}

func reservarVaga(r json.RawMessage) {
	var dados modelo.ReservarVagaJson
	erro := json.Unmarshal(r, &dados)
	if erro != nil {
		fmt.Println("Erro ao decodificar JSON", erro)
		return
	}

	//fmt.Println("dados do veículo", dados.Veiculo)
	modelo.ReservarVaga(&posto_criado, &dados.Veiculo)

	//envia a requisição para o servidor para enviar a resposta para o veículo
	veiculoConexao := modelo.RetornarVagaJson{
		Posto:      posto_criado,
		ID_veiculo: dados.Veiculo.ID,
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
	var veiculoDadosAtualizados modelo.Veiculo
	veiculoEncontrado := -1
	fmt.Println("*****************************************")
	fmt.Printf("\n\nFila atual do posto:\n")
	for i := range posto_criado.Fila {
		if dados.Veiculo.ID == posto_criado.Fila[i].ID {
			posto_criado.Fila[i] = &dados.Veiculo
			veiculoEncontrado = i
		}

		fmt.Printf("Posição %d: ID VEÍCULO: %s LONGITUDE E LATITUDE: %.4f %.4f\n\n", i, posto_criado.Fila[i].ID, posto_criado.Fila[i].Longitude, posto_criado.Fila[i].Latitude)
		fmt.Println("*****************************************")
	}

	if veiculoEncontrado != -1 {
		veiculoDadosAtualizados = *posto_criado.Fila[veiculoEncontrado]
	} else {
		veiculoDadosAtualizados = dados.Veiculo
		veiculoDadosAtualizados.Bateria = 100
		veiculoDadosAtualizados.IsCarregando = false
		veiculoDadosAtualizados.IsDeslocandoAoPosto = false
	}

	modelo.ArrumarPosicaoFila(&posto_criado)

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
