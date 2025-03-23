package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"io"
	"math"
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

var opcao int
var (
	id        string
	latitude  float64
	longitude float64
	//bateria   float64
)
var veiculo modelo.Veiculo
var ticker *time.Ticker
var goroutineCriada bool
var postoRecomendado *modelo.Posto
var posicaoFila int

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
	//alterar loc do veículo
	if veiculo.ID != "" {
		ticker = time.NewTicker(2 * time.Second) // temporizador faz com que chame a função a cada dois segundos
		go func() {
			for range ticker.C {
				modelo.AtualizarLocalizacao(&veiculo)
				fmt.Println("aqui")
			}
		}()
	}

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

func encontrarPostoRecomendado() {

	postos := listarPostos()
	if postos == nil {
		return
	}

	//var posto_recomendado *modelo.Posto
	var menor_tempo time.Duration = -1
	

	for i := range postos {
		posto := &postos[i]
		//tempo ate o posto é a distancia entre o veiculo e o posto multiplicado por 15 segundos, 1 de distancia vezes 15 segundos
		tempo_ate_posto := time.Duration(math.Abs(veiculo.Latitude-posto.Latitude)+math.Abs(veiculo.Longitude-posto.Longitude)) * 15 * time.Second
		tempo_total, posicao := modelo.TempoEstimado(posto, tempo_ate_posto)
		if menor_tempo == -1 {
			menor_tempo = tempo_total
			postoRecomendado = posto
			posicaoFila = posicao
			
		} else if tempo_total < menor_tempo {
			menor_tempo = tempo_total
			postoRecomendado = posto
			posicaoFila = posicao
		}
	}

	if postoRecomendado != nil {
		fmt.Printf("Posto recomendado: %s\n", postoRecomendado.ID)
		if posicaoFila == -1 {
			fmt.Printf("Ultimo da fila")
		} else {
			fmt.Printf("Posição na fila: %d\n", posicaoFila)
		}
	} else {
		fmt.Println("Nenhum posto recomendado encontrado")
	}
}
