package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

var opcao int
var ( 
	id string; latitude float64; longitude float64; //bateria   float64
)
var veiculo modelo.Veiculo
var ticker *time.Ticker
var goroutineCriada bool

func main() {
	// captura sinal em caso do cliente se desconectar
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	resp, erro := http.Get("http://localhost:8080/conectar")

	if erro != nil {
		fmt.Println("F TOTAL")
	}

	defer resp.Body.Close()
	fmt.Println("Conectado!")

	// lidar com a saída em caso de ctrl + c
	go func() {
		<-signalChan
		desconectarDoServidor()
		os.Exit(0)
	}()

	selecionarObjetivo()
}

func selecionarObjetivo() {
	for {
		if veiculo.ID != "" {
			if !goroutineCriada {
				ticker = time.NewTicker(2 * time.Second) // temporizador faz com que chame a função a cada dois segundos
				go func() {
					for range ticker.C {
						modelo.AtualizarLocalizacao(&veiculo)
						fmt.Println("aqui")
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
	valorPraPagar := (100 - modelo.GetNivelBateriaAoChegarNoPosto(veiculo,posto_selecionado)) * 0.5 //0.5 reais por % nivel de bateria
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
		ID_posto:      posto_selecionado.ID,
	}
	req, err := json.Marshal(pagamentoFeito)
	if err != nil {
		fmt.Printf("Erro ao converter pagamento para JSON: %v\n", err)
		return
	}

	// Faz a requisição POST para o servidor
	resp, err := http.Post("http://localhost:8080/pagamento", "application/json", bytes.NewBuffer(req))
	if err != nil {
		fmt.Printf("Erro ao enviar requisição: %v\n", err)
		return
	}

	defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	println("erro ao ler resposta do servidor")
	// 	return
	// }

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

	//fmt.Println("Veículo cadastrado com sucesso!")
	veiculo = modelo.NovoVeiculo(id, longitude, latitude)

	//converto o veiculo pra JSON
	req, err := json.Marshal(veiculo)
	if err != nil {
		fmt.Printf("Erro ao converter veículo para JSON: %v\n", err)
		return
	}

	//faço a requisiçao de POST pro servidor
	post, err := http.Post("http://localhost:8080/cadastrar-veiculo", "application/json", bytes.NewBuffer(req))
	if err != nil {
		fmt.Printf("Erro ao cadastrar veículo: %v\n", err)
		return
	}

	defer post.Body.Close()
}

func listarPostos() []modelo.Posto {
	//fiz a requisicao para listar os postos GET
	resp, erro := http.Get("http://localhost:8080/listar")
	if erro != nil {
		fmt.Println("Erro ao listar postos:", erro)
		return nil
	}

	//to lendo o corpo da resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Erro ao ler resposta:", err)
		return nil
	}

	//to convertendo o JSON para um slice de postos
	var postos []modelo.Posto
	err = json.Unmarshal(body, &postos)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return nil
	}

	//printando as informacoes dos postos
	for i := range postos {
		posto := &postos[i]
		fmt.Printf("ID: %s\n", posto.ID)
		fmt.Printf("Latitude: %.2f\n", posto.Latitude)
		fmt.Printf("Longitude: %.2f\n", posto.Longitude)
		fmt.Printf("Quantidade de carros na fila: %d\n", posto.QtdFila)
		fmt.Printf("Bomba disponivel : %t\n", posto.BombaOcupada)
		fmt.Println("----------------------------------------")
	}

	return postos
}

func desconectarDoServidor() {
	resp, erro := http.Get("http://localhost:8080/desconectar")

	if erro != nil {
		fmt.Println("F TOTAL")
		return
	}

	defer resp.Body.Close()
	fmt.Println("Desconectado!")
}
//TESTANDO
func encontrarPostoRecomendado() {
	// Converte o veículo para JSON
	req, err := json.Marshal(veiculo)
	if err != nil {
		fmt.Printf("Erro ao converter veículo para JSON: %v\n", err)
		return
	}

	// Faz a requisição POST para o servidor
	resp, err := http.Post("http://localhost:8080/posto-recomendado", "application/json", bytes.NewBuffer(req))
	if err != nil {
		fmt.Printf("Erro ao enviar requisição: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Lê a resposta do servidor
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Erro ao ler resposta: %v\n", err)
		return
	}

	// Converte a resposta JSON para a estrutura RecomendadoResponse
	var recomendado modelo.RecomendadoResponse
	err = json.Unmarshal(body, &recomendado)
	if err != nil {
		fmt.Printf("Erro ao converter resposta JSON: %v\n", err)
		return
	}

	// Exibe as informações do posto recomendado
	fmt.Println("*******************************************************")
	fmt.Printf("Posto recomendado: %s\n", recomendado.ID_posto)
	fmt.Printf("Latitude: %.4f\n", recomendado.Latitude)
	fmt.Printf("Longitude: %.4f\n", recomendado.Longitude)
	fmt.Printf("Posição na fila: %d\n", recomendado.Posicao_na_fila)
	fmt.Println("*******************************************************")
}
