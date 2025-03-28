package main

import (
	//"bytes"
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	//"io"
	"net"
	//"net/http"
	"time"
)

type VeiculoJson struct {
	ID        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Bateria   float64 `json:"bateria"`
}

type RecomendadoResponse struct {
	ID_posto        string  `json:"id_posto"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Posicao_na_fila int     `json:"posicao_na_fila"`
}

type Requisicao struct {
	Comando string `json:"comando"`
	// ClienteID string          `json:"cliente_id"`
	Dados json.RawMessage `json:"dados"`
}

var opcao int
var (
	id        string
	latitude  float64
	longitude float64 //bateria   float64
)
var veiculo modelo.Veiculo
var ticker *time.Ticker
var goroutineCriada bool
var conexao net.Conn

func main() {
	var erro error
	conexao, erro = net.Dial("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao conectar ao servidor:", erro)
		return
	}

	defer conexao.Close()

	fmt.Println("Veículo conectado à porta:", conexao.RemoteAddr())

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
	case "listar-postos":
		return response.Dados
	case "encontrar-posto-recomendado":
		return response.Dados
	}

	return nil
}

func selecionarObjetivo() {
	for {
		if veiculo.ID != "" {
			if !goroutineCriada {
				ticker = time.NewTicker(5 * time.Second) // temporizador faz com que chame a função a cada dois segundos
				go func() {
					for range ticker.C {
						modelo.AtualizarLocalizacao(&veiculo)
					}
				}()
				goroutineCriada = true
			}
		}

		fmt.Println("veiculo id", veiculo.ID)

		fmt.Printf("Digite 0 para cadastrar seu veículo\n")
		fmt.Print("Digite 1 para encontrar posto recomendado\n")
		fmt.Printf("Digite 2 para reservar vaga em um posto\n")
		fmt.Printf("Digite 3 para listar todos os postos\n")
		fmt.Scanln(&opcao)
		switch {
		case opcao == 0:
			fmt.Println("Cadastrar veículo")
			cadastrarVeiculo()

		case opcao == 1:
			fmt.Println("Encontrar posto recomendado")
			encontrarPostoRecomendado()

		case opcao == 2:
			fmt.Println("Reservar vaga em um posto")
			reservarVaga()

		case opcao == 3:
			fmt.Println("Listar todos os postos")
			listarPostos()

		default:
			fmt.Println("Opção inválida")
		}
	}
}

func reservarVaga() {
	fmt.Println("Posto recomendado atualmente: ")
	encontrarPostoRecomendado()
	fmt.Println("A seguir a lista com todos os postos disponíveis: ")
	listaDosPosto := listarPostos()
	fmt.Println("Digite o ID do posto que deseja reservar: ")
	var idPosto string
	fmt.Scanln(&idPosto)

	var postoEncontrado bool = false
	//var pagamentoRealizado bool = false

	var posto_selecionado *modelo.Posto

	for i := range listaDosPosto {
		posto := &listaDosPosto[i]
		if posto.ID == idPosto {
			postoEncontrado = true
			posto_selecionado = posto
			break
		}
	}
	if !postoEncontrado {
		fmt.Println("Posto não encontrado")
		return
	}
	valorPraPagar := (100 - modelo.GetNivelBateriaAoChegarNoPosto(veiculo, posto_selecionado)) * 0.5 //0.5 reais por % nivel de bateria
	for {
		fmt.Println("É necessario realizar o pagamento para reservar a vaga")
		fmt.Printf("O valor a ser pago é de %.2f\n", valorPraPagar)
		fmt.Println("Deseja concluir o pagamento? (0 - sim, 1 - nao): ")
		var opcao int
		fmt.Scanf("%d", &opcao)
		if opcao == 0 {
			fmt.Printf("Pagamento realizado com sucesso!, o valor de %.2f foi cobrado da sua conta bancaria\n", valorPraPagar)
			break
		} else if opcao == 1 {
			fmt.Println("Pagamento não realizado")
			return
		} else {
			fmt.Println("Opção inválida")
			return
		}
	}
	// var pagamentoFeito modelo.PagamentoJson
	pagamentoFeito := modelo.PagamentoJson{
		ID_veiculo: veiculo.ID,
		Valor:      valorPraPagar,
		ID_posto:   posto_selecionado.ID,
	}
	req, err := json.Marshal(pagamentoFeito)
	if err != nil {
		fmt.Printf("Erro ao converter pagamento para JSON: %v\n", err)
		return
	}

	requisicao := Requisicao{
		Comando: "reservar-vaga",
		Dados:   req,
	}

	erro := enviarRequisicao(requisicao)

	if erro != nil {
		fmt.Println("erro ao enviar requisiçao")
	}


	resp := receberResposta()
	if resp == nil {
		fmt.Println("Erro ao listar postos")
		return
	}

	//to convertendo o JSON para um slice de postos
	var vagaFeita  RecomendadoResponse
	erroo := json.Unmarshal(resp, &vagaFeita)
	if erroo != nil {
		fmt.Println("Erro ao converter JSON da resposta:", erroo)
		return
	}

	fmt.Println("vaga reservada no posto: ", vagaFeita.ID_posto)
	fmt.Println("latitude: ", vagaFeita.Latitude)
	fmt.Println("longitude: ", vagaFeita.Longitude)
	fmt.Println("posicao na fila: ", vagaFeita.Posicao_na_fila)		

}

func cadastrarVeiculo() {
	fmt.Println("Digite o ID do veículo a ser cadastrado:")
	fmt.Scanln(&id)
	fmt.Println("Digite a latitude do veículo:")
	fmt.Scanln(&latitude)
	fmt.Println("Digite a longitude do veículo:")
	fmt.Scanln(&longitude)
	// fmt.Println("Digite a procetagem de bateria do veículo:")
	// fmt.Scanln(&bateria)

	veiculo = modelo.NovoVeiculo(id, longitude, latitude)

	veiculoJSON, erro := json.Marshal(veiculo)
	if erro != nil {
		fmt.Printf("Erro ao converter veículo para JSON: %v\n", erro)
		return
	}

	req := Requisicao{
		Comando: "cadastrar-veiculo",
		Dados:   veiculoJSON,
	}

	erro = enviarRequisicao(req)

	if erro == nil {
		fmt.Println("Veículo cadastrado com sucesso")
	}
}

func listarPostos() []modelo.Posto {
	//fiz a requisicao para listar os postos GET
	req := Requisicao{
		Comando: "listar-postos",
	}

	enviarRequisicao(req)

	resp := receberResposta()
	if resp == nil {
		fmt.Println("Erro ao listar postos")
		return nil
	}

	//to convertendo o JSON para um slice de postos
	var postos []modelo.Posto
	erro := json.Unmarshal(resp, &postos)
	if erro != nil {
		fmt.Println("Erro ao converter JSON:", erro)
		return nil
	}

	//printando as informacoes dos postos
	for i := range postos {
		posto := &postos[i]
		fmt.Printf("ID: %s\n", posto.ID)
		fmt.Printf("Latitude: %.2f\n", posto.Latitude)
		fmt.Printf("Longitude: %.2f\n", posto.Longitude)
		fmt.Printf("Quantidade de carros na fila: %d\n", len(posto.Fila))
		fmt.Printf("Bomba disponivel : %t\n", posto.BombaOcupada)
		fmt.Println("----------------------------------------")
	}

	return postos
}

// TESTANDO
func encontrarPostoRecomendado() {
	//var requisicao Requisicao
	req, err := json.Marshal(veiculo)
	if err != nil {
		fmt.Printf("Erro ao converter veículo para JSON: %v\n", err)
		return
	}
  
	requisicao := Requisicao{
		Comando: "encontrar-posto-recomendado",
		Dados:   req,
	}
  
	err = enviarRequisicao(requisicao)
	if err != nil {
		fmt.Println("Erro ao enviar requisição")
		return
	}

	resposta := receberResposta()
	if resposta == nil {
		fmt.Println("Erro ao receber resposta")
		return
	}

	// converte a resposta JSON para a estrutura RecomendadoResponse
	var recomendado modelo.RecomendadoResponse
	err = json.Unmarshal(resposta, &recomendado)
	if err != nil {
		fmt.Printf("Erro ao converter resposta JSON: %v\n", err)
		return
	}

	fmt.Println("*******************************************************")
	fmt.Printf("Posto recomendado: %s\n", recomendado.ID_posto)
	fmt.Printf("Latitude: %.4f\n", recomendado.Latitude)
	fmt.Printf("Longitude: %.4f\n", recomendado.Longitude)
	fmt.Printf("Posição na fila: %d\n", recomendado.Posicao_na_fila)
	fmt.Println("*******************************************************")
}
