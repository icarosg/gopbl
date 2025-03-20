package main

import (
	"encoding/json"
	//"fmt"
	"gopbl/modelo"

	//"fmt"
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

func inicializar() {
	posto1 := &modelo.Posto{
		ID:           "posto1",
		Latitude:     800,
		Longitude:    100,
		QtdFila:      10,
		BombaOcupada: true,
	}

	posto2 := &modelo.Posto{
		ID:           "posto2",
		Latitude:     500,
		Longitude:    100,
		QtdFila:      10,
		BombaOcupada: true,
	}

	// Adiciona os postos ao slice
	postosMutex.Lock()
	postos = append(postos, posto1, posto2)
	postosMutex.Unlock()

	// Salva os postos em um arquivo JSON (opcional)
	salvarPostosNoArquivo()
}

func salvarPostosNoArquivo() {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	// Converter a lista de postos para JSON
	postosJSON, err := json.MarshalIndent(postos, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao converter postos para JSON: %s", err)
	}

	// Escrever o JSON em um arquivo usando os.Create
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

	// Escrever o JSON em um arquivo usando os.Create
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

	// Decodificar o JSON do corpo da requisição
	var veiculo modelo.Veiculo
	err := json.NewDecoder(r.Body).Decode(&veiculo)

	if err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Adicionar o veículo ao slice de veículos
	veiculos = append(veiculos, &veiculo)

	salvarVeiculosNoArquivo()

}

func ListarPostos(w http.ResponseWriter, r *http.Request) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	// Converter a lista de postos para JSON
	postosJSON, err := json.Marshal(postos)
	if err != nil {
		http.Error(w, "Erro ao converter postos para JSON", http.StatusInternalServerError)
		return
	}

	// Definir o cabeçalho da resposta como JSON
	w.Header().Set("Content-Type", "application/json")
	// Escrever o JSON na resposta
	w.Write(postosJSON)

	//fmt.Printf("Postos listados: %s\n", string(postosJSON))
}

// func main() {
// 	inicializar()

// 	// Criar um servidor HTTP para servir o arquivo JSON
// 	http.HandleFunc("/posto", handler)
// 	http.HandleFunc("/listar", ListarPostos) // Rota para listar postos
// 	fmt.Println("Servidor iniciado em http://localhost:8050")
// 	log.Fatal(http.ListenAndServe(":8050", nil))
// }
