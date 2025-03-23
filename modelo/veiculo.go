package modelo

// definir rotas para: consulta de pontos de recarga disponíveis;
// reserva de pontos de recarga;
// registro de recargas realizadas.

import (
	"fmt"
	"math/rand"
)

type Veiculo struct {
	ID           string
	Latitude     float64
	Longitude    float64
	Bateria      float64
	IsCarregando bool
}

func NovoVeiculo(id string, inicialLat float64, inicialLong float64) Veiculo {
	return Veiculo{
		ID:           id,
		Latitude:     inicialLat,
		Longitude:    inicialLong,
		Bateria:      100.0, // começa com bateria cheia
		IsCarregando: false,
	}
}

func AtualizarLocalizacao(v *Veiculo) {

	//o defer garante que a liberação do bloqueio ocorra de maneira segura e sempre que a função terminar sua execução

	v.Latitude += (rand.Float64() - 0.5) * 0.001
	v.Longitude += (rand.Float64() - 0.5) * 0.001
	fmt.Println("_________________________________________________________________________________________________")
	fmt.Printf("localalizacao atual do veiculo: latitude %.4f e longitude %.4f\n", v.Latitude, v.Longitude)
	fmt.Println("_________________________________________________________________________________________________")
}

func DiminuirNivelBateria(v *Veiculo) {

	if !v.IsCarregando {
		// diminui a bateria entre 30.4 e 15.1 por atualização
		v.Bateria -= rand.Float64()*30.4 + 15.1
		if v.Bateria < 0 {
			v.Bateria = 0
		}
	}
}

func GetNivelBateria(v *Veiculo) float64 {

	return v.Bateria
}

func GetLocalizacaoVeiculo(v *Veiculo) (float64, float64) {

	return v.Latitude, v.Longitude
}

// func CarregarBateria(v *Veiculo) {
//
//

// 	v.IsCarregando = true
// }
