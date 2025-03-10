package modelo

// definir rotas para: consulta de pontos de recarga disponíveis;
// reserva de pontos de recarga;
// registro de recargas realizadas.

import (
	"math/rand"
	"sync"
)

type Veiculo struct {
	ID         string
	Latitude   float64
	Longitude  float64
	Bateria    float64
	mu         sync.Mutex
	IsCarregando bool
}

func NovoVeiculo(id string, inicialLat float64, inicialLong float64) Veiculo {
	return Veiculo{
		ID:        id,
		Latitude:  inicialLat,
		Longitude: inicialLong,
		Bateria:   100.0, // começa com bateria cheia
	}
}

func AtualizarLocalizacao(v *Veiculo) {
	v.mu.Lock()
	defer v.mu.Unlock() //o defer garante que a liberação do bloqueio ocorra de maneira segura e sempre que a função terminar sua execução

	v.Latitude += (rand.Float64() - 0.5) * 0.001
	v.Longitude += (rand.Float64() - 0.5) * 0.001
}

func UpdateBAtattery(v *Veiculo) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.IsCarregando {
		// diminui a bateria entre 30.4 e 15.1 por atualização
		v.Bateria -= rand.Float64()*30.4 + 15.1
		if v.Bateria < 0 {
			v.Bateria = 0
		}
	}
}

func GetNivelBateria(v *Veiculo) float64 {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.Bateria
}

func GetLocalizacaoVeiculo(v *Veiculo) (float64, float64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.Latitude, v.Longitude
}

func CarregarBateria(v *Veiculo) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.IsCarregando = true
}

func PararCarregamentoBateria(v *Veiculo) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.IsCarregando = false
	v.Bateria = 100.0
}