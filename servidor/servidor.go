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
	postosMutex                sync.Mutex
	postos                     []*modelo.Posto // Slice para armazenar todos os postos
	veiculos                   []*modelo.Veiculo
	postosParaArquivo          []*modelo.Posto
	conexoes_postos            []net.Conn
	conexoes_clientes          []net.Conn
	dicionarioConexoesClientes = make(map[string]net.Conn)
	//pagamentos  []PagamentoJson
)

var (
	qtdClientesConectados int
	mutex                 sync.Mutex
)

func main() {
	// cria um listener TCP na porta 9090, ouvindo em todas as interfaces

	listener, erro := net.Listen("tcp", "localhost:8080")

	//listener, erro := net.Listen("tcp", "0.0.0.0:9090")

	if erro != nil {
		fmt.Println("Erro ao iniciar o servidor:", erro)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor iniciado em localhost:8080")
	// inicializar()

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
	}

}

func cliente(conexao net.Conn) {
	defer func() {
		decrementar()
		for i := range conexoes_clientes {
			c := conexoes_clientes[i]
			if conexao == c {
				salvarNoArquivo("veiculos.json")
				conexoes_clientes = append(conexoes_clientes[:i], conexoes_clientes[i+1:]...)
				fmt.Println("cliente desconectado, conexoes de postos restantes: ", conexoes_clientes)
			}
		}

		for chave, valor := range dicionarioConexoesClientes { //remove do dicionário
			if valor == conexao {
				delete(dicionarioConexoesClientes, chave)
				break
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
			cadastrarVeiculo(req, conexao)

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

		case "reservar-vaga-retornoPosto":
			//envia a resposta para o veículo
			var dados *modelo.RetornarVagaJson
			erro := json.Unmarshal(req.Dados, &dados)
			if erro != nil {
				fmt.Println("Erro ao decodificar JSON", erro)
				return
			}

			_, existe := dicionarioConexoesClientes[dados.ID_veiculo] //verifica se o id do veículo existe no dicionário

			if existe {
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

				// fmt.Println("enviando a resposta do reservar vaga")
				// fmt.Println("conex clientes", conexoes_clientes)
				// enviarResposta(conexoes_clientes[0], respostaRequisicao)
				enviarResposta(dicionarioConexoesClientes[dados.ID_veiculo], respostaRequisicao)
			} else {
				fmt.Println("O veículo está desconectado! Não foi possível enviar a resposta!")
			}

		case "atualizar-posicao-veiculo":
			//envia a resposta para o veículo
			var dados modelo.RetornarAtualizarPosicaoFila
			erro := json.Unmarshal(req.Dados, &dados)
			if erro != nil {
				fmt.Println("Erro ao decodificar JSON", erro)
				return
			}

			//fmt.Println("dados", dados)
			_, existe := dicionarioConexoesClientes[dados.Veiculo.ID] //verifica se o id do veículo existe no dicionário

			if existe {
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

				// fmt.Println("enviando a resposta do atualizar fila")
				// fmt.Println("conex clientes", conexoes_clientes[0])
				// enviarResposta(conexoes_clientes[0], respostaRequisicao)
				enviarResposta(dicionarioConexoesClientes[dados.Veiculo.ID], respostaRequisicao)
			} else {
				fmt.Println("O veículo está desconectado! Não foi possível enviar a resposta!")
			}

		case "listarPostosDoArquivo":
			listarPostosDoArquivo(conexao)

		case "cadastrar-posto":
			fmt.Println("teste")
			cadastrarPosto(req)
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

func postoNoArquivo(nome string) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

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
		for _, p := range postosSalvos {
			postosExistentes[p.ID] = p
		}
	}

	// add os novos veiculos
	for _, p := range postosParaArquivo {
		postosExistentes[p.ID] = *p
	}

	postosAtualizados := make([]modelo.Posto, 0, len(postosExistentes))
	for _, p := range postosExistentes {
		postosAtualizados = append(postosAtualizados, p)
	}

	postosJSON, err := json.MarshalIndent(postosAtualizados, "", "    ")
	if err != nil {
		log.Fatalf("Erro ao converter veículos para JSON: %s", err)
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

func cadastrarVeiculo(req Requisicao, conexao net.Conn) {
	// decodifica o JSON do body da req
	var veiculo modelo.Veiculo
	erro := json.Unmarshal(req.Dados, &veiculo)

	if erro != nil {
		fmt.Println("Erro ao cadastrar o veículo")
		return
	}

	dicionarioConexoesClientes[veiculo.ID] = conexao
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

func listarPostosDoArquivo(conexao net.Conn) {
	postosMutex.Lock()
	defer postosMutex.Unlock()

	nomeArquivo := "postos.json"
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
			log.Printf("Erro ao ler JSON existente.")
		}

		// add a lista
		for _, p := range postosSalvos {
			postosExistentes[p.ID] = p
		}

		postosAtualizados := make([]modelo.Posto, 0, len(postosExistentes))
		for _, p := range postosExistentes {
			postosAtualizados = append(postosAtualizados, p)
		}

		postosJSON, erro := json.Marshal(postosAtualizados)
		if erro != nil {
			fmt.Println("Erro ao codificar os veículos")
			return
		}

		response := Requisicao{
			Comando: "listar-postosNoArquivo",
			Dados:   postosJSON,
		}

		enviarResposta(conexao, response)
	} else {
		fmt.Println("Arquivo de postos inexistente.")
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
	pag := modelo.Pagamento{
		Veiculo:  pagamentoJson.Veiculo.ID,
		Valor:    pagamentoJson.Valor,
		ID_posto: pagamentoJson.ID_posto,
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
		res := Requisicao{
			Comando: "reservar-vaga",
			Dados:   nil,
		}

		enviarResposta(conexao, res)
		return
	} else {
		veiculoConexao := modelo.ReservarVagaJson{
			Veiculo: pagamentoJson.Veiculo,
		}

		req, err := json.Marshal(veiculoConexao)
		if err != nil {
			fmt.Printf("Erro ao converter veículo e conexão para JSON: %v\n", err)
			return
		}

		for i := range veiculos {
			if veiculos[i].ID == pagamentoJson.Veiculo.ID {
				if veiculos[i].Pagamentos == nil {
					veiculos[i].Pagamentos = []modelo.Pagamento{}
				}
				veiculos[i].Pagamentos = append(veiculos[i].Pagamentos, pag)
			}
		}

		res := Requisicao{
			Comando: "reservar-vaga",
			Dados:   req,
		}

		enviarResposta(conexaoPosto, res)
	}
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

	fmt.Println("--------------------------------------------")
	fmt.Printf("veiculo recebido: \nID: %s \nLatitude: %.2f \nLongitude %.2f \nBateria em %.2f \nEsta carregando: %t \nEsta se deslocando: %t\n\n", attPos.Veiculo.ID, attPos.Veiculo.Latitude, attPos.Veiculo.Longitude, attPos.Veiculo.Bateria, attPos.Veiculo.IsCarregando, attPos.Veiculo.IsDeslocandoAoPosto)
	fmt.Println("--------------------------------------------")

	// procura o posto pelo ID
	var conexaoPosto net.Conn
	for i := range postos {
		if postos[i].ID == attPos.ID_posto {
			conexaoPosto = conexoes_postos[i]
			break
		}
	}

	veiculoConexao := modelo.ReservarVagaJson{
		Veiculo: attPos.Veiculo,
	}

	req, erro := json.Marshal(veiculoConexao)
	if erro != nil {
		fmt.Printf("Erro ao converter veículo para JSON: %v\n", erro)
		return
	}

	if conexaoPosto == nil {
		fmt.Println("Posto não encontrado")

		respostaRequisicao := Requisicao{
			Comando: "atualizar-posicao-veiculo",
			Dados:   req,
		}
		enviarResposta(conexao, respostaRequisicao)
		return
	} else {
		respostaRequisicao := Requisicao{
			Comando: "atualizar-posicao-veiculo",
			Dados:   req,
		}
		enviarResposta(conexaoPosto, respostaRequisicao)
	}
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
		fmt.Println("****************************")
		fmt.Printf("Posto recomendado: %s\n", postoRecomendado.ID)
		fmt.Printf("Posição na fila: %d\n", posicaoFila)
		fmt.Println("****************************")
		recomendadoResponse := modelo.RecomendadoResponse{
			ID_posto:  postoRecomendado.ID,
			Latitude:  postoRecomendado.Latitude,
			Longitude: postoRecomendado.Longitude,
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
	} else {
		fmt.Println("Nenhum posto recomendado encontrado")
		response := Requisicao{
			Comando: "encontrar-posto-recomendado",
		}

		enviarResposta(conexao, response)
	}

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

func cadastrarPosto(req Requisicao) {
	// decodifica o JSON do body da req
	var posto modelo.Posto
	erro := json.Unmarshal(req.Dados, &posto)

	if erro != nil {
		fmt.Println("Erro ao cadastrar o posto")
		return
	}

	postosParaArquivo = append(postosParaArquivo, &posto)

	postoNoArquivo("postos.json")

	fmt.Println("Posto cadastrado")
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
