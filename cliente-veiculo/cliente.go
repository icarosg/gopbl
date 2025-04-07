package main

import (
	"encoding/json"
	"fmt"
	"gopbl/modelo"
	"net"
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
	Comando string          `json:"comando"`
	Dados   json.RawMessage `json:"dados"`
}

var opcao int
var (
	id        string
	latitude  float64
	longitude float64 //bateria   float64
)
var veiculo modelo.Veiculo
var posto_selecionado *modelo.Posto
var ticker *time.Ticker
var goroutineCriada bool
var conexao net.Conn

func main() {
	var erro error
	conexao, erro = net.Dial("tcp", "localhost:8080")
	
	// em ambiente Docker, usar o nome do serviço em vez de localhost
	//conexao, erro = net.Dial("tcp", "servidor:9090")
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
	case "listar-veiculos":
		return response.Dados
	case "new-vaga":
		return response.Dados
	case "atualizar-posicao-veiculo":
		return response.Dados
	case "tipo-cliente":
		return response.Dados
	}

	return nil
}

func retornarConexaoCliente() {
	for {
		resp := receberResposta()
		if resp == nil {
			fmt.Println("Erro ao retornar conexão do posto")
			continue
		}

		var tipo string
		erro := json.Unmarshal(resp, &tipo)
		if erro != nil {
			fmt.Println("Erro ao converter JSON:", erro)
			continue
		}
		if tipo == "tipo-cliente" {
			req := Requisicao{
				Comando: "adicionar-conexao",
				Dados:   json.RawMessage(`"cliente"`),
			}
			enviarRequisicao(req)
			break
		}
	}
}

func selecionarObjetivo() {
	retornarConexaoCliente()

	for {
		if veiculo.ID != "" {
			if !goroutineCriada {
				ticker = time.NewTicker(10 * time.Second) // temporizador faz com que chame a função a cada 5 segundos
				go func() {
					for range ticker.C {
						if veiculo.IsDeslocandoAoPosto {
							modelo.DeslocarParaPosto(&veiculo, posto_selecionado) // altera a posição para ir para o posto e faz a requisição para salvar nova posição na fila do posto
							atualizarPosicaoVeiculoNaFila()
							// modelo.ArrumarPosicaoFila(posto_selecionado)
						} else {
							modelo.AtualizarLocalizacao(&veiculo)
						}
					}
				}()
				goroutineCriada = true
			}
		}

		if veiculo.ID != "" {
			fmt.Println("****************************")
			fmt.Printf("O id do seu veiculo é: %s. \nLongitude%.2f \nLatitude: %.2f \nBateria em: %.2f\n", veiculo.ID, veiculo.Longitude, veiculo.Latitude, veiculo.Bateria)
			fmt.Println("****************************")		
		} else {
			fmt.Println("****************************")
			fmt.Println("Você ainda não possui um veículo.")
			fmt.Println("****************************")	
		}

		opcao = -1
		fmt.Println("-------------------------------------------")
		fmt.Printf("Digite 0 para cadastrar seu veículo\n")
		fmt.Printf("Digite 1 para listar os veículos e importar algum\n")
		fmt.Print("Digite 2 para encontrar posto recomendado\n")
		fmt.Printf("Digite 3 para reservar vaga em um posto\n")
		fmt.Printf("Digite 4 para listar todos os postos\n")
		fmt.Printf("Digite 5 para exibir os pagamentos realizados\n")
		fmt.Println("-------------------------------------------")
		fmt.Scanln(&opcao)
		switch {
		case opcao == 0:
			fmt.Println("Cadastrar veículo")
			cadastrarVeiculo()

		case opcao == 1:
			fmt.Println("Listar e importar veículo")
			listarEImportarVeiculo()

		case opcao == 2:
			fmt.Println("Encontrar posto recomendado")
			encontrarPostoRecomendado()

		case opcao == 3:
			fmt.Println("Reservar vaga em um posto")
			reservarVaga()

		case opcao == 4:
			fmt.Println("Listar todos os postos")
			listarPostos()
		case opcao == 5:
			fmt.Println("Listar pagamentos realizados")
			exibirPagamentosRealizados()

		default:
			fmt.Println("Opção inválida")
		}
	}
}

func exibirPagamentosRealizados() {
	for i := range veiculo.Pagamentos {
		pagamento := veiculo.Pagamentos[i]
		fmt.Println()
		fmt.Println("--------------------------------------------")
		fmt.Println("Pagamentos realizados: ")
		fmt.Printf("ID do veículo: %s\n", pagamento.Veiculo)
		fmt.Printf("ID do posto: %s\n", pagamento.ID_posto)
		fmt.Printf("Valor pago: %.2f\n", pagamento.Valor)
		fmt.Println("----------------------------------------")
		fmt.Println()
	}
}

func cadastrarVeiculo() {
	fmt.Println("--------------------------------------------")
	fmt.Println("Digite o ID do veículo a ser cadastrado:")
	fmt.Scanln(&id)
	fmt.Println("Digite a latitude do veículo:")
	fmt.Scanln(&latitude)
	fmt.Println("Digite a longitude do veículo:")
	fmt.Scanln(&longitude)
	fmt.Println("--------------------------------------------")
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

func listarEImportarVeiculo() []modelo.Veiculo {
	req := Requisicao{
		Comando: "listar-veiculos",
	}

	enviarRequisicao(req)

	resp := receberResposta()
	if resp == nil {
		fmt.Println("Erro ao listar veículos")
		return nil
	}

	//to convertendo o JSON para um slice de veiculos
	var veiculos []modelo.Veiculo
	erro := json.Unmarshal(resp, &veiculos)
	if erro != nil {
		fmt.Println("Erro ao converter JSON:", erro)
		return nil
	}

	for i := range veiculos {
		veiculo := &veiculos[i]
		fmt.Println()
		fmt.Println("--------------------------------------------")
		fmt.Println("Abaixo veiculo disponívei para importação")
		fmt.Printf("ID: %s\n", veiculo.ID)
		fmt.Printf("Latitude: %.2f\n", veiculo.Latitude)
		fmt.Printf("Longitude: %.2f\n", veiculo.Longitude)
		fmt.Printf("Nível da bateria %.2f\n", veiculo.Bateria)
		fmt.Println("----------------------------------------")
	}

	fmt.Println("Digite o ID do veículo que deseja importar: ")
	var idVeiculo string
	fmt.Scanln(&idVeiculo)

	var veiculoEncontrado bool = false
	var veiculo_selecionado *modelo.Veiculo

	for i := range veiculos {
		v := &veiculos[i]
		if v.ID == idVeiculo {
			veiculoEncontrado = true
			veiculo_selecionado = v
			break
		}
	}
	if !veiculoEncontrado {
		fmt.Println("--------------------------------------------")
		fmt.Println("Veículo não encontrado")
		fmt.Println("--------------------------------------------")
		return nil
	} else {
		veiculoImportado := *veiculo_selecionado
		veiculo = modelo.NovoVeiculo(veiculoImportado.ID, veiculoImportado.Latitude, veiculoImportado.Latitude)
		veiculo.Bateria = veiculoImportado.Bateria
		veiculo.Pagamentos = veiculoImportado.Pagamentos

		veiculoJSON, erro := json.Marshal(veiculo)
		if erro != nil {
			fmt.Printf("Erro ao converter veículo para JSON: %v\n", erro)
			return nil
		}

		req := Requisicao{
			Comando: "cadastrar-veiculo",
			Dados:   veiculoJSON,
		}

		erro = enviarRequisicao(req)

		if erro == nil {
			fmt.Println("--------------------------------------------")
			fmt.Println("Veículo importado com sucesso!")
			fmt.Println("--------------------------------------------")
		}
	}

	return veiculos
}

func listarPostos() []modelo.Posto {
	req := Requisicao{
		Comando: "listar-postos",
	}

	enviarRequisicao(req)

	time.Sleep(2 * time.Second) // aguarda por 2 segundos
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

	for i := range postos {
		posto := &postos[i]
		fmt.Println("----------------------------------------")
		fmt.Printf("ID: %s\n", posto.ID)
		fmt.Printf("Latitude: %.2f\n", posto.Latitude)
		fmt.Printf("Longitude: %.2f\n", posto.Longitude)
		fmt.Printf("Quantidade de carros na fila: %d\n", len(posto.Fila))
		fmt.Printf("Bomba disponivel : %t\n", posto.BombaOcupada)
		fmt.Println("----------------------------------------")
	}

	return postos
}

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

	time.Sleep(2 * time.Second) // aguarda por 2 segundos

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
	//fmt.Printf("Posição na fila: %d\n", recomendado.Posicao_na_fila)
	fmt.Println("*******************************************************")
}

func reservarVaga() {
	fmt.Println("--------------------------------------------")
	fmt.Println("Posto recomendado atualmente: ")
	encontrarPostoRecomendado()
	fmt.Println("A seguir a lista com todos os postos disponíveis: ")
	listaDosPosto := listarPostos()
	fmt.Println("Digite o ID do posto que deseja reservar: ")
	var idPosto string
	fmt.Scanln(&idPosto)
	fmt.Println("--------------------------------------------")

	var postoEncontrado bool = false
	//var pagamentoRealizado bool = false

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
		fmt.Println("--------------------------------------------")
		fmt.Println("É necessario realizar o pagamento para reservar a vaga")
		fmt.Printf("O valor a ser pago é de %.2f\n", valorPraPagar)
		fmt.Println("Deseja concluir o pagamento? (0 - sim, 1 - nao): ")
		fmt.Println("--------------------------------------------")
		var opcao int
		fmt.Scanf("%d", &opcao)
		if opcao == 0 {
			fmt.Printf("Estamos tentando realizar o pagamento de %.2f\n", valorPraPagar)
			fmt.Println("--------------------------------------------")
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
		Veiculo:  veiculo,
		Valor:    valorPraPagar,
		ID_posto: posto_selecionado.ID,
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

	time.Sleep(2 * time.Second) // aguarda por 2 segundos

	resp := receberResposta()
	if resp == nil {
		fmt.Println("*******************")
		fmt.Println("ID do posto não encontrado! Não foi possível reservar a vaga")
		fmt.Printf("Pagamento de %.2f foi estornado para sua conta\n", valorPraPagar)
		fmt.Println("*******************")
		return
	} else {		
		//to convertendo o JSON para um slice de postos
		var vagaFeita RecomendadoResponse
		erro = json.Unmarshal(resp, &vagaFeita)
		if erro != nil {
			fmt.Println("Erro ao converter JSON da resposta:", erro)
			return
		}
		fmt.Printf("Pagamento realizado com sucesso!, o valor de %.2f foi cobrado da sua conta bancaria\n", valorPraPagar)
		pag := modelo.Pagamento{
			Veiculo:  veiculo.ID,
			Valor:    valorPraPagar,
			ID_posto: posto_selecionado.ID,
		}
		//veiculo.Pagamentos = append(veiculo.Pagamentos, pagamentoFeito)
		veiculo.Pagamentos = append(veiculo.Pagamentos, pag)

		veiculo.IsDeslocandoAoPosto = true //ao reservar, automaticamente o veículo começa se deslocar para o posto
		fmt.Println("*******************")
		fmt.Println("vaga reservada no posto: ", vagaFeita.ID_posto)
		fmt.Println("latitude: ", vagaFeita.Latitude)
		fmt.Println("longitude: ", vagaFeita.Longitude)
		fmt.Println("*******************")
		// fmt.Println("posicao na fila: ", vagaFeita.Posicao_na_fila)
	}
}

func atualizarPosicaoVeiculoNaFila() {
	attPosicao := modelo.AtualizarPosicaoNaFila{
		Veiculo:  veiculo,
		ID_posto: posto_selecionado.ID,
	}
	req, err := json.Marshal(attPosicao)
	if err != nil {
		fmt.Printf("Erro ao converter atualização de localização para JSON: %v\n", err)
		return
	}

	requisicao := Requisicao{
		Comando: "atualizar-posicao-veiculo",
		Dados:   req,
	}

	erro := enviarRequisicao(requisicao)

	if erro != nil {
		fmt.Println("erro ao enviar requisiçao")
	}

	time.Sleep(2 * time.Second) // aguarda por 2 segundos

	resp := receberResposta()
	if resp == nil {
		fmt.Println("Erro ao receber resposta da atualização da localização do veículo")
		return
	}

	var dados modelo.RetornarAtualizarPosicaoFila
	erro = json.Unmarshal(resp, &dados)
	if erro != nil {
		fmt.Println("Erro ao decodificar JSON", erro)
		return
	}

	if dados.Posto.ID != "" {
		fmt.Printf("\n\nveiculo recebido: %f %f\n\n", dados.Veiculo.Longitude, dados.Veiculo.Latitude)
		//fmt.Println("postorecebido:", &dados.Posto)

		veiculo = dados.Veiculo
		posto_selecionado = &dados.Posto

		if !veiculo.IsDeslocandoAoPosto {
			fmt.Println("*************************")
			fmt.Printf("Posto %s: Veículo %s removido da fila.\n", posto_selecionado.ID, veiculo.ID)
			modelo.CarregarBateria(&veiculo, posto_selecionado)
			fmt.Println("*************************")
		}
	} else {
		fmt.Printf("\n\nO posto foi desconectado! O veículo não está mais se deslocando para lá!\n\n")

		veiculo = dados.Veiculo
		veiculo.IsDeslocandoAoPosto = false
		veiculo.IsCarregando = false
	}

}
