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
	postosMutex       sync.Mutex
	postos            []*modelo.Posto // Slice para armazenar todos os postos
	veiculos          []*modelo.Veiculo
	conexoes_postos   []net.Conn
	conexoes_clientes []net.Conn
	//pagamentos  []PagamentoJson
)

var (
	qtdClientesConectados int
	mutex                 sync.Mutex
)

var opcao int
var (
	id          string
	latitude    float64
	longitude   float64
	selecionado *modelo.Posto
)
var goroutineCriada bool
var ticker *time.Ticker

var arquivoPostosCriados bool = false

func main() {
	// cria um listener TCP na porta 8080
	listener, erro := net.Listen("tcp", "localhost:8080")
	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor iniciado em localhost:8080")
	inicializar()

	if postos != nil {
		if !goroutineCriada {
			ticker = time.NewTicker(5 * time.Second) // temporizador faz com que chame a função a cada 5 segundos
			go func() {
				for range ticker.C {
					AtualizarFilas()
				}
			}()
			goroutineCriada = true
		}
	}

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

		tipo := TipoDeCliente(conexao)
		if tipo == "posto" {
			conexoes_postos = append(conexoes_postos, conexao)
			go posto(conexao)
		} else {
			conexoes_clientes = append(conexoes_postos, conexao)
			go cliente(conexao)
		}

		go AtualizarFilas()
		go menu()
	}

}

func menu() {
	fmt.Printf("Digite 0 para cadastrar um posto\n")
	fmt.Printf("Digite 1 para listar os postos\n")
	fmt.Printf("Digite 2 para selecionar um posto e exibir a fila de veiculos\n")
	fmt.Printf("Digite 3 para listar os veiculos\n")
	fmt.Scanln(&opcao)
	switch {
	case opcao == 0:
		fmt.Println("Cadastrar posto")
		cadastrarPosto()
	case opcao == 1:
		fmt.Println("Listar postos")
		listarPostosServidor()

	case opcao == 2:
		fmt.Println("Digite o id do posto que deseja selecionar:")
		fmt.Scanln(&id)
		for i := range postos {
			if postos[i].ID == id {
				selecionado = postos[i]
				exibirFilaPosto(selecionado)
				break
			}
		}

	case opcao == 3:
		fmt.Println("listar veiculos")
		listarVeiculosServidor()

	default:
		fmt.Println("Opção inválida")
	}
}

func cliente(conexao net.Conn) {
	defer func() {
		decrementar()
		for i := range conexoes_clientes {
			c := conexoes_clientes[i]
			if conexao == c {
				conexoes_clientes = append(conexoes_clientes[:i], conexoes_clientes[i+1:]...)
				fmt.Println("cliente desconectado, conexoes de postos restantes: ", conexoes_clientes)
			}
		}
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

		case "listar-veiculos":
			listarVeiculos(conexao)

		case "listar-postos":
			listarPostos(conexao)

		case "encontrar-posto-recomendado":
			postoRecomendado(conexao, req)

		case "reservar-vaga":
			reservarVagaPosto(conexao, req)

		case "atualizar-posicao-veiculo":
			atualizarPosicaoVeiculoNaFila(conexao, req)
		}
	}
}

func posto(conexao net.Conn) {
	defer func() {
		decrementar()
		for i := range conexoes_postos {
			p := conexoes_postos[i]
			if conexao == p {
				conexoes_postos = append(conexoes_postos[:i], conexoes_postos[i+1:]...)
				fmt.Println("posto desconectado, conexoes de postos restantes: ", conexoes_postos)
			}
		}
		fmt.Println("Cliente desconectado. Total de clientes conectados:", getQtdClientes())
		conexao.Close()
	}() // decrementa após a conexão ser encerrada

	buffer := make([]byte, 4096)
	for {
		n, erro := conexao.Read(buffer)
		if erro != nil {
			if erro == io.EOF {
				fmt.Printf("O posto %s fechou a conexão\n", conexao.RemoteAddr())
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
		case "listar-postos":
			var postoRecebido *modelo.Posto
			erro := json.Unmarshal(req.Dados, &postoRecebido)
			if erro != nil {
				fmt.Println("erro ao decodificar resposta")
			}
			postos = append(postos, postoRecebido)
		case "cadastrar-posto":
			testSalvarPosto(req,conexao)			
		}
	}
}

func testSalvarPosto(req Requisicao, conexao net.Conn){
	postosMutex.Lock()
	defer postosMutex.Unlock()

	nomeArquivo := "postos.json"

	var postoRecebido *modelo.Posto
	fmt.Println("tentando decodificar")
	erro := json.Unmarshal(req.Dados, &postoRecebido)
	if erro != nil {
		fmt.Println("erro ao decodificar resposta")
	}		
	fmt.Println("Posto recebido aki:", postoRecebido)	
	postos = append(postos, postoRecebido)
	salvarPostoNoArquivo(nomeArquivo)

	reqi := Requisicao{
		Comando: "cadastrou",
	}
	e := enviarResposta(conexao, reqi)
	if e != nil{
		fmt.Println("erro ao enviar resposta")
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
	if !arquivoPostosCriados {
		salvarPostoNoArquivo("postos.json")
		arquivoPostosCriados = true
	}
}


// func salvarPostosNoArquivo() {
// 	postosMutex.Lock()
// 	defer postosMutex.Unlock()

// 	armazenarPostosNaLista()

// 	// converte a lista de postos para JSON
// 	postosJSON, err := json.MarshalIndent(postos, "", "    ")
// 	if err != nil {
// 		log.Fatalf("Erro ao converter postos para JSON: %s", err)
// 	}

// 	// escreve o JSON em um arquivo
// 	file, err := os.Create("postos.json")

// 	if err != nil {
// 		log.Fatalf("Erro ao criar o arquivo: %s", err)
// 	}
// 	defer file.Close()

// 	_, err = file.Write(postosJSON)
// 	if err != nil {
// 		log.Fatalf("Erro ao escrever no arquivo: %s", err)
// 	}

// 	log.Println("Postos salvos em postos.json")
// }

func salvarPostoNoArquivo(nome string) {
	// postosMutex.Lock()
	// defer postosMutex.Unlock()
	//armazenarPostosNaLista()
	nomeArquivo := nome

	postosExistentes := make(map[string]modelo.Posto)

	// verifica se o arquivo já existe
	if _, err := os.Stat(nomeArquivo); err == nil {
		arquivo, err := os.Open(nomeArquivo)
		if err != nil {
			log.Fatalf("Erro ao abrir o arquivo: %s", err)
		}
		defer arquivo.Close()

		var postosSalvos []modelo.Posto
		if err := json.NewDecoder(arquivo).Decode(&postosSalvos); err != nil {
			log.Printf("Erro ao ler JSON existente. Criando novo arquivo: %s", err)
		}

		// add a lista
		for _, v := range postosSalvos {
			postosExistentes[v.ID] = v
		}
	}

	// add os novos veiculos
	for _, v := range postos {
		postosExistentes[v.ID] = *v
	}

	postosAtualizados := make([]modelo.Posto, 0, len(postosExistentes))
	for _, v := range postosExistentes {
		postosAtualizados = append(postosAtualizados, v)
	}

	postosJSON, err := json.MarshalIndent(postosAtualizados, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao converter postos para JSON: %s", err)
	}

	// abre o arquivo para escrita (substitui o conteúdo antigo)
	arquivo, err := os.Create(nomeArquivo)
	if err != nil {
		log.Fatalf("Erro ao criar o arquivo: %s", err)
	}
	defer arquivo.Close()

	_, err = arquivo.Write(postosJSON)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo: %s", err)
	}

	//log.Println("Veículos salvos em", nomeArquivo)
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

func cadastrarPosto() {
	fmt.Println("Cadastrar posto")
	fmt.Println("Digite o ID do posto:")
	fmt.Scanln(&id)
	fmt.Println("Digite a latitude do posto:")
	fmt.Scanln(&latitude)
	fmt.Println("Digite a longitude do posto:")
	fmt.Scanln(&longitude)

	novoPosto := modelo.NovoPosto(id, latitude, longitude)

	postosMutex.Lock()
	postos = append(postos, &novoPosto)
	postosMutex.Unlock()

	salvarPostoNoArquivo("postos.json")

	fmt.Println("Posto cadastrado com sucesso!")
}

func listarPostosServidor() {
	for i := range postos {
		posto := postos[i]
		fmt.Printf("ID: %s\n", posto.ID)
		fmt.Printf("Latitude: %.2f\n", posto.Latitude)
		fmt.Printf("Longitude: %.2f\n", posto.Longitude)
		fmt.Printf("Quantidade de carros na fila: %d\n", len(posto.Fila))
		fmt.Printf("Bomba disponivel : %t\n", posto.BombaOcupada)
		fmt.Println("----------------------------------------")
	}
}

func listarVeiculosServidor() {
	for i := range veiculos {
		veiculo := veiculos[i]
		fmt.Printf("ID: %s\n", veiculo.ID)
		fmt.Printf("Latitude: %.2f\n", veiculo.Latitude)
		fmt.Printf("Longitude: %.2f\n", veiculo.Longitude)
		fmt.Println("----------------------------------------")
	}
}

func exibirFilaPosto(posto *modelo.Posto) {
	fmt.Printf("Fila do posto %s:\n", posto.ID)
	for i := range posto.Fila {
		veiculo := posto.Fila[i]
		fmt.Printf("ID: %s\n", veiculo.ID)
		fmt.Printf("Latitude: %.2f\n", veiculo.Latitude)
		fmt.Printf("Longitude: %.2f\n", veiculo.Longitude)
		tempoEstimado, _ := modelo.TempoEstimado(posto, 0)
		fmt.Printf("Tempo estimado para o carregamento desse veiculo: %s\n", tempoEstimado)
		fmt.Printf("Posição na fila: %d\n", modelo.GetPosFila(*veiculo, posto))
		fmt.Println("----------------------------------------")
	}
}

func AtualizarFilas() {
	for i := range postos {
		p := postos[i]
		go modelo.ArrumarPosicaoFila(p)
	}
}

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

func listarVeiculos(conexao net.Conn) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	nomeArquivo := "veiculos.json"
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
			log.Printf("Erro ao ler JSON existente.")
		}

		// add a lista
		for _, v := range veiculosSalvos {
			veiculosExistentes[v.ID] = v
		}

		veiculosAtualizados := make([]modelo.Veiculo, 0, len(veiculosExistentes))
		for _, v := range veiculosExistentes {
			veiculosAtualizados = append(veiculosAtualizados, v)
		}

		veiculosJSON, erro := json.Marshal(veiculosAtualizados)
		if erro != nil {
			fmt.Println("Erro ao codificar os veículos")
			return
		}

		response := Requisicao{
			Comando: "listar-veiculos",
			Dados:   veiculosJSON,
		}

		enviarResposta(conexao, response)
	} else {
		fmt.Println("Arquivo de veículos inexistente.")
	}
}

func reservarVagaPosto(conexao net.Conn, requisicao Requisicao) {
	// postosMutex.Lock()
	// defer postosMutex.Unlock()

	// decodifica o JSON do body da req
	var pagamentoJson modelo.PagamentoJson
	err := json.Unmarshal(requisicao.Dados, &pagamentoJson)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON")
		return
	}

	var veiculo *modelo.Veiculo
	for i := range veiculos {
		if veiculos[i].ID == pagamentoJson.ID_veiculo {
			veiculo = veiculos[i]
			break
		}
	}

	// procura o posto pelo ID
	var posto *modelo.Posto
	for i := range postos {
		if postos[i].ID == pagamentoJson.ID_posto {
			posto = postos[i]
			break
		}
	}

	if posto == nil {
		fmt.Println("Posto não encontrado")
		return
	}

	if veiculo == nil {
		fmt.Println("Posto não encontrado")
		return
	}

	reservado := modelo.ReservarVaga(posto, veiculo)
	if reservado {
		fmt.Printf("Vaga reservada com sucesso para o veículo %s no posto %s\n", veiculo.ID, posto.ID)
	} else {
		fmt.Printf("Veículo %s adicionado à fila do posto %s\n", veiculo.ID, posto.ID)
	}
	resposta := modelo.RecomendadoResponse{
		ID_posto:        posto.ID,
		Latitude:        posto.Latitude,
		Longitude:       posto.Longitude,
		Posicao_na_fila: modelo.GetPosFila(*veiculo, posto),
	}

	respostaJSON, err := json.Marshal(resposta)
	if err != nil {
		fmt.Println("Erro ao codificar resposta")
		return
	}
	respostaRequisicao := Requisicao{
		Comando: "reservar-vaga",
		Dados:   respostaJSON,
	}
	enviarResposta(conexao, respostaRequisicao)
}

func atualizarPosicaoVeiculoNaFila(conexao net.Conn, requisicao Requisicao) {
	// decodifica o JSON do body da req
	var attPosicaoVeiculoJson modelo.AtualizarPosicaoNaFila
	err := json.Unmarshal(requisicao.Dados, &attPosicaoVeiculoJson)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON")
		return
	}

	var veiculo *modelo.Veiculo
	for i := range veiculos {
		if veiculos[i].ID == attPosicaoVeiculoJson.Veiculo.ID {
			veiculo = &attPosicaoVeiculoJson.Veiculo
			break
		}
	}

	// procura o posto pelo ID
	var posto *modelo.Posto
	for i := range postos {
		if postos[i].ID == attPosicaoVeiculoJson.ID_posto {
			posto = postos[i]
			break
		}
	}

	if posto == nil {
		fmt.Println("Posto não encontrado para atualização da posição do veículo")
		return
	}

	if veiculo == nil {
		fmt.Println("Veiculo não encontrado para atualização da posição do veículo")
		return
	}

	reservado := modelo.ReservarVaga(posto, veiculo)
	if reservado {
		fmt.Printf("(DESLOCANDO AO POSTO) Posição do veículo %s atualizado para a latitude e longitude: %v, %v\n", veiculo.ID, veiculo.Latitude, veiculo.Longitude)
	} else {
		fmt.Printf("ALGUM PROBLEMA DEU")
	}
	resposta := modelo.RecomendadoResponse{
		ID_posto:        posto.ID,
		Latitude:        posto.Latitude,
		Longitude:       posto.Longitude,
		Posicao_na_fila: modelo.GetPosFila(*veiculo, posto),
	}

	respostaJSON, err := json.Marshal(resposta)
	if err != nil {
		fmt.Println("Erro ao codificar resposta")
		return
	}
	respostaRequisicao := Requisicao{
		Comando: "atualizar-posicao-veiculo",
		Dados:   respostaJSON,
	}
	enviarResposta(conexao, respostaRequisicao)
}

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
		Comando: "encontrar-posto-recomendado",
		Dados:   respostaJSON,
	}

	enviarResposta(conexao, response)
}

func TipoDeCliente(conexao net.Conn) string {
	pergunta := Requisicao{
		Comando: "tipo-cliente",
		Dados:   json.RawMessage(`"tipo-cliente"`),
	}
	err := enviarResposta(conexao, pergunta)
	if err != nil {
		fmt.Println("erro ao enviar a requisicao para essa conexao")
		return "erro"
	}
	resposta := receberResposta(conexao)
	if resposta == nil {
		fmt.Println("erro ao receber a resposta dessa conexao")
		return "erro"
	}
	var tipo_de_cliente string
	fail := json.Unmarshal(resposta, &tipo_de_cliente)
	if fail != nil {
		fmt.Println("erro ao descodificar a resposta")
		return "erro"
	}
	return tipo_de_cliente
}

func armazenarPostosNaLista() {
    postos = nil //limpa os postos

    //percorre a lista das conexoes que sao postos
    for i := range conexoes_postos {
        conexaoPosto := conexoes_postos[i]
        req := Requisicao{
            Comando: "get-posto",
        }
        //envia a requisiçao pra conexao de um posto para ele responder com o posto cadastrado ou importado
        err := enviarResposta(conexaoPosto, req)
        if err != nil {
            fmt.Println("erro ao enviar requisição pra esse posto")
            return
        }
    }

    time.Sleep(1 * time.Second) // aguarda por 1 segundo
}

func listarPostos(conexao net.Conn) {		
	armazenarPostosNaLista()

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

	//envia a lista de postos para o cliente-veiculo que pediu
	enviarResposta(conexao, response)
}

func enviarResposta(conexao net.Conn, resposta Requisicao) error {
	dados, erro := json.Marshal(resposta)
	if erro != nil {
		fmt.Printf("Erro ao codificar resposta: %v", erro)
		return erro
	}

	_, erro = conexao.Write(dados)
	if erro != nil {
		log.Printf("Erro ao enviar resposta: %v", erro)
		return erro
	}

	return nil
}

func receberResposta(conexao net.Conn) json.RawMessage {
	buffer := make([]byte, 4096)

	mutex.Lock()
	n, erro := conexao.Read(buffer)
	mutex.Unlock()
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
	case "listar-veiculos":
		return response.Dados
	case "tipo-de-cliente":
		return response.Dados
	case "adicionar-conexao":
		return response.Dados
	case "get-posto":
		return response.Dados
	}

	return nil
}