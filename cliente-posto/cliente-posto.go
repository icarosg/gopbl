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
var respostas = make(chan *Requisicao, 10) // canal com buffer

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
	go verificarAlgoNoBuffer()
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
	case "postos-salvos":
		//fmt.Println("chegou aki")
		return &response
	default:
		fmt.Println("Comando desconhecido:", response.Comando)
		return nil
	}

	//return nil
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
					// case "cadastrar-posto":
					// 	cadastrarPosto()				
				}


			} else {
				fmt.Println("Você ainda não possui um posto.")
			}
			//return
		}
		//mutex.Unlock()
		time.Sleep(1 * time.Second) // Espera 1 segundo antes de verificar novamente
	}
}

func selecionarObjetivo() {
	retornarConexaoPosto()
	//go verificarAlgoNoBuffer()
	
		for {		
			// if posto_criado.ID == "" {
				fmt.Printf("Digite 0 para cadastrar seu posto\n")
				fmt.Printf("Digite 1 para listar os postos e importar algum\n")
				fmt.Scanln(&opcao)

				switch {
				case opcao == 0:
					cadastrarPosto()

				case opcao == 1:
					fmt.Println("Listar e importar posto")
					listarEImportarPosto()

				default:
					fmt.Println("Opção inválida")
				}
			// } else {
			// 	go verificarAlgoNoBuffer()
			// }
			
		
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
		fmt.Printf("posto disponivel para importação ID: %s, Latitude: %f, Longitude: %f\n", (*postosNoArquivo)[i].ID, (*postosNoArquivo)[i].Latitude, (*postosNoArquivo)[i].Longitude)
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
