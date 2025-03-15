package main

import (
	"fmt"
	"gopbl/modelo"
	"time"
)

func main() {
	fmt.Println("Iniciando simulação de veiculos elétricos e postos de recarga")

	// criar alguns veiculos
	veiculo1 := modelo.NovoVeiculo("V001", -23.5505, -46.6333) 
	veiculo2 := modelo.NovoVeiculo("V002", -22.9068, -43.1729) 

	// criar postos
	posto1 := modelo.NovoPosto("P001", -23.5505, -46.6333) 
	posto2 := modelo.NovoPosto("P002", -22.9068, -43.1729) 

	fmt.Println("\n--- Informações Iniciais ---")

	// verificar as bomba
	fmt.Printf("Posto %s - Bomba ocupada: %t\n", posto1.ID, modelo.GetBombaDisponivel(&posto1))
	fmt.Printf("Posto %s - Bomba ocupada: %t\n", posto2.ID, modelo.GetBombaDisponivel(&posto2))

	//testar 5 vezes
	for i := 0; i < 5; i++ {
		fmt.Printf("\n--- Ciclo %d ---\n", i+1)

		// atualizar veiculo 1
		fmt.Printf("veiculo %s:\n", veiculo1.ID)
		modelo.AtualizarLocalizacao(&veiculo1)
		modelo.DiminuirNivelBateria(&veiculo1)
		lat1, long1 := modelo.GetLocalizacaoVeiculo(&veiculo1)
		bateria1 := modelo.GetNivelBateria(&veiculo1)
		fmt.Printf("  Localização: (%.6f, %.6f)\n", lat1, long1)
		fmt.Printf("  Bateria: %.2f%%\n", bateria1)

		// atualizar veiculo 2
		fmt.Printf("veiculo %s:\n", veiculo2.ID)
		modelo.AtualizarLocalizacao(&veiculo2)
		modelo.DiminuirNivelBateria(&veiculo2)
		lat2, long2 := modelo.GetLocalizacaoVeiculo(&veiculo2)
		bateria2 := modelo.GetNivelBateria(&veiculo2)
		fmt.Printf("  Localização: (%.6f, %.6f)\n", lat2, long2)
		fmt.Printf("  Bateria: %.2f%%\n", bateria2)

		// testar a recarga
		if bateria1 < 50 {
			fmt.Printf("\n  Tentando reservar vaga no posto %s para o veiculo %s\n", posto1.ID, veiculo1.ID)
			
			//veiculo chegou no posto
			veiculo1.Latitude, veiculo1.Longitude = modelo.GetLocalizacaoPosto(&posto1)
			reservaOk := modelo.ReservarVaga(&posto1, &veiculo1)

			if reservaOk {
				fmt.Printf("  veiculo %s conseguiu vaga direta no posto %s\n", veiculo1.ID, posto1.ID)
				modelo.CarregarBateria(&veiculo1)
			} else {
				fmt.Printf("  veiculo %s está na fila do posto %s\n", veiculo1.ID, posto1.ID)
			}
		}

		if bateria2 < 50 {
			fmt.Printf("\n  Tentando reservar vaga no posto %s para o veiculo %s\n", posto2.ID, veiculo2.ID)
			//veiculo chegou no posto
			veiculo2.Latitude, veiculo2.Longitude = modelo.GetLocalizacaoPosto(&posto2)
			reservaOk := modelo.ReservarVaga(&posto2, &veiculo2)

			if reservaOk {
				fmt.Printf("  veiculo %s conseguiu vaga direta no posto %s\n", veiculo2.ID, posto2.ID)
				modelo.CarregarBateria(&veiculo2)
			} else {
				fmt.Printf("  veiculo %s está na fila do posto %s\n", veiculo2.ID, posto2.ID)
			}
		}

		// simular liberacao de vaga
		if i%2 == 1 {
			if modelo.GetBombaDisponivel(&posto1) {
				fmt.Printf("\n  Liberando vaga no posto %s\n", posto1.ID)
				modelo.LiberarVaga(&posto1)
			}

			if modelo.GetBombaDisponivel(&posto2) {
				fmt.Printf("\n  Liberando vaga no posto %s\n", posto2.ID)
				modelo.LiberarVaga(&posto2)
			}
		}

		// acelerar o carregamento
		if veiculo1.IsCarregando {
			fmt.Printf("\n  Simulando fim de carregamento para veiculo %s (acelerado para teste)\n", veiculo1.ID)
			modelo.PararCarregamentoBateria(&veiculo1)
			fmt.Printf("  Bateria após carregamento: %.2f%%\n", modelo.GetNivelBateria(&veiculo1))
		}

		if veiculo2.IsCarregando {
			fmt.Printf("\n  Simulando fim de carregamento para veiculo %s (acelerado para teste)\n", veiculo2.ID)
			modelo.PararCarregamentoBateria(&veiculo2)
			fmt.Printf("  Bateria após carregamento: %.2f%%\n", modelo.GetNivelBateria(&veiculo2))
		}

		// esperar um pouco
		time.Sleep(2 * time.Second)
	}

	fmt.Println("\nteste encerrado.")
}
