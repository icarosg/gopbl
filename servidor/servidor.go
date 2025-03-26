package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"io"
	"log"
	"math"
	"net"

	//"net/http"
	"os"
	"sync"
	"time"
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

type Requisicao struct {
	Comando string          `json:"comando"`
	Dados   json.RawMessage `json:"dados"`
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

var arquivoPostosCriados bool = false

func main() {
	// http.HandleFunc("/posto", handler)
	// http.HandleFunc("/listar", listarPostos)
	// http.HandleFunc("/cadastrar-veiculo", cadastrarVeiculo)
	// http.HandleFunc("/posto-recomendado", postoRecomendado)
	// http.HandleFunc("/reservar-vaga", reservarVagaPosto)
	// http.HandleFunc("/pagamento", reservarVagaPosto)

	// cria um listener TCP na porta 8080
	listener, erro := net.Listen("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor iniciado em localhost:8080")
	inicializar()

	for {
		conexao, erro := listener.Accept()

		if erro != nil {
			fmt.Println("Erro ao conectao o cliente", erro)
			fmt.Println("Erro ao conectar o cliente", erro)
			continue // continua aguardando outras conexões
		}

		incrementar()

		fmt.Println("Cliente conectado à porta:", conexao.RemoteAddr())
		fmt.Println("Total de clientes conectados:", getQtdClientes())

		go cliente(conexao)
	}

}

func cliente(conexao net.Conn) {
	defer func() {
		decrementar()
		fmt.Println("Cliente desconectado. Total de clientes conectados:", getQtdClientes())
		conexao.Close()
	}() // decrementa após a conexão ser encerrada

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

		erro = json.Unmarshal(buffer[:n], &req)

		if erro != nil {
			fmt.Println("Erro ao decodificar a requisição")
			continue
		}

		switch req.Comando {
		case "cadastrar-veiculo":
			cadastrarVeiculo(req)

		case "listar-postos":
			listarPostos(conexao)

		case "encontrar-posto-recomendado":
			postoRecomendado(conexao, req)

		}
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
	posto1 := modelo.NovoPosto("posto1", 10, 15)
	posto2 := modelo.NovoPosto("posto2", 50, 50)

	// Adiciona um veículo à fila do posto2 com coordenadas mais realistas
	veiculo1 := modelo.NovoVeiculo("veiculo1", 30, 30)

	// Adiciona o veículo apenas ao posto2
	posto2.Fila = append(posto2.Fila, &veiculo1)
	posto2.QtdFila = 1

	// posto1.Fila = append(posto1.Fila, veiculo1)
	// posto1.QtdFila = 1

	// adiciona os postos ao slice
	postosMutex.Lock()
	postos = append(postos, &posto1, &posto2)
	postosMutex.Unlock()

	if !arquivoPostosCriados {
		salvarPostosNoArquivo()
		arquivoPostosCriados = true
	}
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

func salvarNoArquivo(nome string) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	nomeArquivo := nome

	veiculosExistentes := make(map[string]modelo.Veiculo)

	// verifica se o arquivo já existe
	if _, err := os.Stat(nomeArquivo); err == nil {
		arquivo, err := os.Open(nomeArquivo)
		if err != nil {
			log.Fatalf("Erro ao abrir o arquivo: %s", err)
		}
		defer arquivo.Close()

		var veiculosSalvos []modelo.Veiculo
		if err := json.NewDecoder(arquivo).Decode(&veiculosSalvos); err != nil {
			log.Printf("Erro ao ler JSON existente. Criando novo arquivo: %s", err)
		}

		// add a lista
		for _, v := range veiculosSalvos {
			veiculosExistentes[v.ID] = v
		}
	}

	// add os novos veiculos
	for _, v := range veiculos {
		veiculosExistentes[v.ID] = *v
	}

	veiculosAtualizados := make([]modelo.Veiculo, 0, len(veiculosExistentes))
	for _, v := range veiculosExistentes {
		veiculosAtualizados = append(veiculosAtualizados, v)
	}

	veiculoJSON, err := json.MarshalIndent(veiculosAtualizados, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao converter veículos para JSON: %s", err)
	}

	// abre o arquivo para escrita (substitui o conteúdo antigo)
	arquivo, err := os.Create(nomeArquivo)
	if err != nil {
		log.Fatalf("Erro ao criar o arquivo: %s", err)
	}
	defer arquivo.Close()

	_, err = arquivo.Write(veiculoJSON)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %s", err)
	}

	//log.Println("Veículos salvos em", nomeArquivo)
}

// func handler(w http.ResponseWriter, r *http.Request) {
// 	http.ServeFile(w, r, "postos.json")
// }

func cadastrarVeiculo(req Requisicao) {
	//aki tava travando o sistema
	// postosMutex.Lock()
	// defer postosMutex.Unlock()

	// decodifica o JSON do body da req
	var veiculo modelo.Veiculo
	erro := json.Unmarshal(req.Dados, &veiculo)

	if erro != nil {
		fmt.Println("Erro ao cadastrar o veículo")
		return
	}

	veiculos = append(veiculos, &veiculo)

	salvarNoArquivo("veiculos.json")

	fmt.Println("Veículo cadastrado")

}

// func reservarVagaPosto(w http.ResponseWriter, r *http.Request) {
// 	// postosMutex.Lock()
// 	// defer postosMutex.Unlock()

// 	// decodifica o JSON do body da req
// 	var pagamentoJson modelo.PagamentoJson
// 	err := json.NewDecoder(r.Body).Decode(&pagamentoJson)
// 	if err != nil {
// 		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
// 		return
// 	}

// 	var veiculo *modelo.Veiculo
// 	for i := range veiculos {
// 		if veiculos[i].ID == pagamentoJson.ID_veiculo {
// 			veiculo = veiculos[i]
// 			break
// 		}
// 	}

// 	// procura o posto pelo ID
// 	var posto *modelo.Posto
// 	for i := range postos {
// 		if postos[i].ID == pagamentoJson.ID_posto {
// 			posto = postos[i]
// 			break
// 		}
// 	}

// 	if posto == nil {
// 		http.Error(w, "Posto não encontrado", http.StatusNotFound)
// 		return
// 	}

// 	if veiculo == nil {
// 		http.Error(w, "Veículo não encontrado", http.StatusNotFound)
// 		return
// 	}

// 	reservado := modelo.ReservarVaga(posto, veiculo)
// 	if reservado {
// 		fmt.Printf("Vaga reservada com sucesso para o veículo %s no posto %s\n", veiculo.ID, posto.ID)
// 		w.WriteHeader(http.StatusOK)
// 	} else {
// 		fmt.Printf("Veículo %s adicionado à fila do posto %s\n", veiculo.ID, posto.ID)
// 		w.WriteHeader(http.StatusAccepted)
// 	}
// }

func postoRecomendado(conexao net.Conn, req Requisicao) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	var veiculo modelo.Veiculo
	err := json.Unmarshal(req.Dados, &veiculo)

	if err != nil {
		fmt.Println("Erro ao achar posto recomendado")
		return
	}

	var menor_tempo time.Duration = -1
	var postoRecomendado *modelo.Posto
	var posicaoFila int

	for i := range postos {
		posto := postos[i]
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
		fmt.Printf("Posição na fila: %d\n", posicaoFila)
	} else {
		fmt.Println("Nenhum posto recomendado encontrado")
	}

	recomendadoResponse := modelo.RecomendadoResponse{
		ID_posto:        postoRecomendado.ID,
		Latitude:        postoRecomendado.Latitude,
		Longitude:       postoRecomendado.Longitude,
		Posicao_na_fila: posicaoFila,
	}

	respostaJSON, err := json.Marshal(recomendadoResponse)
	if err != nil {
		fmt.Println("Erro ao codificar posto recomendado")
		return
	}

	response := Requisicao{
		Comando: "posto-recomendado",
		Dados:   respostaJSON,
	}

	enviarResposta(conexao, response)
}

func listarPostos(conexao net.Conn) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	// converte a lista de postos para JSON
	postosJSON, erro := json.Marshal(postos)
	if erro != nil {
		fmt.Println("Erro ao codificar postos")
		return
	}

	response := Requisicao{
		Comando: "listar-postos",
		Dados:   postosJSON,
	}

	enviarResposta(conexao, response)

	//fmt.Printf("Postos listados: %s\n", string(postosJSON))
}

func enviarResposta(conexao net.Conn, resposta Requisicao) {
	dados, erro := json.Marshal(resposta)
	if erro != nil {
		fmt.Printf("Erro ao codificar resposta: %v", erro)
		return
	}

	_, erro = conexao.Write(dados)
	if erro != nil {
		log.Printf("Erro ao enviar resposta: %v", erro)
	}
}
