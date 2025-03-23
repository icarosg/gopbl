package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"log"
	"net/http"
	"os"
	"sync"
)

type PostoJson struct {
	ID              string  `json:"id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	QuantidadeFila  int     `json:"quantidade de carros na fila"`
	Disponibilidade bool    `json:"bomba disponivel"`
}

type PagamentoJson struct {
	ID_veiculo string  `json:"id_veiculo"`
	Valor      float64 `json:"valor"`
	Posto      string  `json:"id_posto"`
}

var (
	postosMutex sync.Mutex
	postos      []*modelo.Posto // Slice para armazenar todos os postos
	veiculos    []*modelo.Veiculo
	//pagamentos  []PagamentoJson
)

var (
	qtdClientesConectados int
	mutex                 sync.Mutex
)

func main() {
	http.HandleFunc("/conectar", conexao)
	http.HandleFunc("/desconectar", desconectar)
	http.HandleFunc("/posto", handler)
	http.HandleFunc("/listar", listarPostos)
	http.HandleFunc("/cadastrar-veiculo", cadastrarVeiculo)

	fmt.Println("Servidor HTTP iniciado em http://localhost:8080")
	erro := http.ListenAndServe("localhost:8080", nil)
	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
}

func conexao(w http.ResponseWriter, r *http.Request) {
	incrementar()
	inicializar()

	// exibe em qual porta o cliente foi conectado
	fmt.Println("Cliente conectado:", r.RemoteAddr)
	fmt.Println("Total de clientes conectados:", getQtdClientes())

	// responde o cliente
	_, erro := fmt.Fprintf(w, "Conectado ao servidor! Total de clientes conectados: %d", getQtdClientes())
	if erro != nil {
		fmt.Println("Erro ao responder ao cliente:", erro)
	}
}

func desconectar(w http.ResponseWriter, r *http.Request) {
	decrementar()

	fmt.Println("Cliente desconectado:", r.RemoteAddr)
	fmt.Println("Total de clientes conectados:", getQtdClientes())

	// responde o cliente
	_, erro := fmt.Fprintf(w, "Desconectado do servidor! Total de clientes conectados: %d", getQtdClientes())
	if erro != nil {
		fmt.Println("Erro ao responder ao cliente:", erro)
	}
}

func incrementar() {
	mutex.Lock()
	qtdClientesConectados++
	mutex.Unlock()
}

func decrementar() {
	mutex.Lock()
	qtdClientesConectados--
	mutex.Unlock()
}

func getQtdClientes() int {
	mutex.Lock()
	defer mutex.Unlock()
	return qtdClientesConectados
}

func inicializar() {
	posto1 := &modelo.Posto{
		ID:           "posto1",
		Latitude:     800,
		Longitude:    100,
		Fila:         make([]*modelo.Veiculo, 0),
		QtdFila:      0,
		BombaOcupada: true,
	}

	posto2 := &modelo.Posto{
		ID:           "posto2",
		Latitude:     10,
		Longitude:    10,
		Fila:         make([]*modelo.Veiculo, 0),
		QtdFila:      0,
		BombaOcupada: true,
	}

	// Adiciona um veículo à fila do posto2 com coordenadas mais realistas
	veiculo1 := &modelo.Veiculo{
		ID:           "veiculo1",
		Latitude:     10,
		Longitude:    10,
		Bateria:      20,
		IsCarregando: false,
	}

	// Adiciona o veículo apenas ao posto2
	posto2.Fila = append(posto2.Fila, veiculo1)
	posto2.QtdFila = 1

	posto1.Fila = append(posto1.Fila, veiculo1)
	posto1.QtdFila = 1

	// adiciona os postos ao slice
	postosMutex.Lock()
	postos = append(postos, posto1, posto2)
	postosMutex.Unlock()

	salvarPostosNoArquivo()
}

func salvarPostosNoArquivo() {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	// converte a lista de postos para JSON
	postosJSON, err := json.MarshalIndent(postos, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao converter postos para JSON: %s", err)
	}

	// escreve o JSON em um arquivo
	file, err := os.Create("postos.json")

	if err != nil {
		log.Fatalf("Erro ao criar o arquivo: %s", err)
	}
	defer file.Close()

	_, err = file.Write(postosJSON)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %s", err)
	}

	log.Println("Postos salvos em postos.json")
}

func salvarVeiculosNoArquivo() {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	veiculoJSON, err := json.MarshalIndent(veiculos, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao converter veiculos para JSON: %s", err)
	}

	// escrever o JSON em um arquivo
	file, err := os.Create("veiculos.json")
	if err != nil {
		log.Fatalf("Erro ao criar o arquivo: %s", err)
	}
	defer file.Close()

	_, err = file.Write(veiculoJSON)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %s", err)
	}

	log.Println("Veiculos salvos em veiculos.json")

}

func handler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "postos.json")
}

func cadastrarVeiculo(w http.ResponseWriter, r *http.Request) {
	//aki tava travando o sistema
	// postosMutex.Lock()
	// defer postosMutex.Unlock()

	// decodifica o JSON do body da req
	var veiculo modelo.Veiculo
	err := json.NewDecoder(r.Body).Decode(&veiculo)

	if err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	veiculos = append(veiculos, &veiculo)

	salvarVeiculosNoArquivo()

}

func listarPostos(w http.ResponseWriter, r *http.Request) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	// converte a lista de postos para JSON
	postosJSON, err := json.Marshal(postos)
	if err != nil {
		http.Error(w, "Erro ao converter postos para JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // define o cabeçalho da resposta como JSON
	w.Write(postosJSON)                                // escrever o JSON na resposta

	//fmt.Printf("Postos listados: %s\n", string(postosJSON))
}
