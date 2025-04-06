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

// var goroutineCriada bool
// var ticker *time.Ticker

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

	// if postos != nil {
	// 	if !goroutineCriada {
	// 		ticker = time.NewTicker(5 * time.Second) // temporizador faz com que chame a função a cada 5 segundos
	// 		go func() {
	// 			for range ticker.C {
	// 				atualizarFilas()
	// 			}
	// 		}()
	// 		goroutineCriada = true
	// 	}
	// }

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

		tipo := tipoDeCliente(conexao)
		if tipo == "posto" {
			conexoes_postos = append(conexoes_postos, conexao)
			go posto(conexao)
		} else {
			conexoes_clientes = append(conexoes_clientes, conexao)
			go cliente(conexao)
		}

		//go atualizarFilas()
		//go menu()
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
		case "postos-salvos":
			enviarPostosSalvos(conexao)					

		case "reservar-vaga-retornoPosto":
			//envia a resposta para o veículo
			var dados *modelo.RetornarVagaJson
			erro := json.Unmarshal(req.Dados, &dados)
			if erro != nil {
				fmt.Println("Erro ao decodificar JSON", erro)
				return
			}

			resposta := modelo.RecomendadoResponse{
				ID_posto:  dados.Posto.ID,
				Latitude:  dados.Posto.Latitude,
				Longitude: dados.Posto.Longitude,
				// Posicao_na_fila: modelo.GetPosFila(*veiculo, postoRecebido),
			}

			respostaJSON, err := json.Marshal(resposta)
			if err != nil {
				fmt.Println("Erro ao codificar resposta")
				return
			}
			respostaRequisicao := Requisicao{
				Comando: "new-vaga",
				Dados:   respostaJSON,
			}

			fmt.Println("enviando a resposta do reservar vaga")
			fmt.Println("conex clientes", conexoes_clientes)
			enviarResposta(conexoes_clientes[0], respostaRequisicao)

		case "atualizar-posicao-veiculo":
			//envia a resposta para o veículo
			var dados modelo.RetornarAtualizarPosicaoFila
			erro := json.Unmarshal(req.Dados, &dados)
			if erro != nil {
				fmt.Println("Erro ao decodificar JSON", erro)
				return
			}

			resposta := modelo.RetornarAtualizarPosicaoFila{
				Veiculo: dados.Veiculo,
				Posto:   dados.Posto,
				// Posicao_na_fila: modelo.GetPosFila(*veiculo, postoRecebido),
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

			fmt.Println("enviando a resposta do atualizar fila")
			fmt.Println("conex clientes", conexoes_clientes[0])
			enviarResposta(conexoes_clientes[0], respostaRequisicao)
		}
	}
}

func enviarPostosSalvos(conexao net.Conn) {
	postosMutex.Lock()
	defer postosMutex.Unlock()
	
	nomeArquivo := "postos.json"
	//mapPosto := map[string]modelo.Posto{}

	arquivo, err := os.Open(nomeArquivo)
    if err != nil {
		fmt.Println("erro ao abrir arquivo pra carregar os postos")
        return
    }
    defer arquivo.Close()

	var postosArquivo []*modelo.Posto
	var postosExistentesArquivo []*modelo.Posto
    decoder := json.NewDecoder(arquivo)
    if err := decoder.Decode(&postosArquivo); err != nil {        
		fmt.Println("erro ao decodificar JSON do arquivo dos postos")
		return
    }

	for i := range postosArquivo {
		posto := postosArquivo[i]
		if posto != nil{
			postosExistentesArquivo = append(postosExistentesArquivo, posto)
		}
	}
	postosJSON, err := json.Marshal(postosExistentesArquivo)
	if err != nil {
		fmt.Println("erro ao codificar os postos salvos no arquivo para JSON antes de enviar para o posto")
		return
	}
	req := Requisicao{
		Comando: "postos-salvos",
		Dados:   postosJSON,
	}
	//fmt.Println(postosExistentesArquivo)
	//time.Sleep(1 * time.Second) // Espera 1 segundo para garantir que a resposta seja recebida
	teste := enviarResposta(conexao, req)
	if teste != nil {
		fmt.Println("deu pau akl oh")
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
	armazenarPostosNaLista()

	// decodifica o JSON do body da req
	var pagamentoJson modelo.PagamentoJson
	erro := json.Unmarshal(requisicao.Dados, &pagamentoJson)
	if erro != nil {
		fmt.Println("Erro ao decodificar JSON")
		return
	}

	// procura o posto pelo ID
	var conexaoPosto net.Conn
	for i := range postos {
		if postos[i].ID == pagamentoJson.ID_posto {
			conexaoPosto = conexoes_postos[i]
			break
		}
	}

	if conexaoPosto == nil {
		fmt.Println("Posto não encontrado")
		return
	}

	veiculoConexao := modelo.ReservarVagaJson{
		Veiculo: pagamentoJson.Veiculo,
	}

	req, err := json.Marshal(veiculoConexao)
	if err != nil {
		fmt.Printf("Erro ao converter veículo e conexão para JSON: %v\n", err)
		return
	}

	res := Requisicao{
		Comando: "reservar-vaga",
		Dados:   req,
	}

	enviarResposta(conexaoPosto, res)
}

func atualizarPosicaoVeiculoNaFila(conexao net.Conn, requisicao Requisicao) {
	armazenarPostosNaLista()

	// decodifica o JSON do body da req
	var attPos modelo.AtualizarPosicaoNaFila
	erro := json.Unmarshal(requisicao.Dados, &attPos)
	if erro != nil {
		fmt.Println("Erro ao decodificar JSON")
		return
	}

	// procura o posto pelo ID
	var conexaoPosto net.Conn
	for i := range postos {
		if postos[i].ID == attPos.ID_posto {
			conexaoPosto = conexoes_postos[i]
			break
		}
	}

	if conexaoPosto == nil {
		fmt.Println("Posto não encontrado")
		return
	}

	veiculoConexao := modelo.ReservarVagaJson{
		Veiculo: attPos.Veiculo,
	}

	req, erro := json.Marshal(veiculoConexao)
	if erro != nil {
		fmt.Printf("Erro ao converter veículo para JSON: %v\n", erro)
		return
	}

	// reservado := modelo.ReservarVaga(posto, veiculo)
	// if reservado {
	// 	fmt.Printf("(DESLOCANDO AO POSTO) Posição do veículo %s atualizado para a latitude e longitude: %v, %v\n", veiculo.ID, veiculo.Latitude, veiculo.Longitude)
	// } else {
	// 	fmt.Printf("ALGUM PROBLEMA DEU")
	// }
	// resposta := modelo.RecomendadoResponse{
	// 	ID_posto:        posto.ID,
	// 	Latitude:        posto.Latitude,
	// 	Longitude:       posto.Longitude,
	// 	Posicao_na_fila: modelo.GetPosFila(*veiculo, posto),
	// }

	// respostaJSON, err := json.Marshal(resposta)
	// if err != nil {
	// 	fmt.Println("Erro ao codificar resposta")
	// 	return
	// }
	respostaRequisicao := Requisicao{
		Comando: "atualizar-posicao-veiculo",
		Dados:   req,
	}
	enviarResposta(conexaoPosto, respostaRequisicao)
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

	armazenarPostosNaLista()

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
		//Posicao_na_fila: posicaoFila,
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

func tipoDeCliente(conexao net.Conn) string {
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
	//fmt.Println(dados)
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
	case "postos-salvos":
		return response.Dados
	case "add-fila":
		return response.Dados
	}

	return nil
}