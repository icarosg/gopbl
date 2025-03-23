package modelo

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Posto struct {
	ID           string
	Latitude     float64
	Longitude    float64
	mu           sync.Mutex
	Fila         []*Veiculo
	QtdFila      int
	BombaOcupada bool
}

func NovoPosto(id string, lat float64, long float64) Posto {
	fmt.Printf("Posto %s criado na localização (%.6f, %.6f)",
		id, lat, long)

	return Posto{
		ID:           id,
		Latitude:     lat,
		Longitude:    long,
		Fila:         make([]*Veiculo, 0),
		QtdFila:      0,
		BombaOcupada: false,
	}
}

func ReservarVaga(p *Posto, v *Veiculo) bool { // retorna true se caso conseguir abastecer diretamente
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.BombaOcupada && p.Latitude == v.Latitude && p.Longitude == v.Longitude {
		p.BombaOcupada = true
		fmt.Printf("Posto %s: Vaga reservada para veículo %s.", p.ID, v.ID)
		return true
	}

	p.Fila = append(p.Fila, v)
	fmt.Printf("Posto %s: Veículo %s adicionado à fila de espera. Posição: %d", p.ID, v.ID, p.QtdFila)
	return false
}

func LiberarVaga(p *Posto) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.BombaOcupada = false
	fmt.Printf("Posto %s com a bomba liberada.\n", p.ID)

	if len(p.Fila) > 0 {
		for i := range p.Fila {
			if p.Fila[i].Latitude == p.Latitude && p.Fila[i].Longitude == p.Longitude {
				carregarBateriaVeiculo := p.Fila[i]

				p.Fila = append(p.Fila[:i], p.Fila[i+1:]...) // remove o veículo da fila; índice do primeiro elemento a ser removido e índice após o último elemento a ser removido

				CarregarBateria(carregarBateriaVeiculo)
				fmt.Printf("Posto %s: Veículo %s removido da fila e iniciando carregamento\n", p.ID, carregarBateriaVeiculo.ID)

				p.BombaOcupada = true
				break
			}
		}
	}
}

func GetBombaDisponivel(p *Posto) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.BombaOcupada
}

func GetLocalizacaoPosto(p *Posto) (float64, float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Latitude, p.Longitude
}

func PararCarregamentoBateria(v *Veiculo) {

	v.IsCarregando = false
	v.Bateria = 100.0
	fmt.Printf("[%s] Carregamento concluído em: %s | Nível de bateria: %.2f%%\n", v.ID, time.Now().Format("02/01/2006 15:04:05"), v.Bateria)
}

func CarregarBateria(v *Veiculo) {
	v.IsCarregando = true
	tempoInicio := time.Now()
	fmt.Printf("[%s] Carregamento iniciado em: %s | Nível de bateria inicial: %.2f%%\n", v.ID, tempoInicio.Format("02/01/2006 15:04:05"), v.Bateria)

	// Goroutine para parar o carregamento após 1 minuto por cada 1% de bateria que falta
	go func() {
		time.Sleep(time.Duration(100-v.Bateria) * time.Minute) // Espera 1 minuto por cada 1% de bateria que falta

		//v.IsCarregando = false
		PararCarregamentoBateria(v)
	}()
}

// teste
func TempoEstimado(p *Posto, tempoDistanciaVeiculo time.Duration) (time.Duration, int) {
	tempo_total := tempoDistanciaVeiculo
	posicao_na_fila := -1

	// Se não houver veículos na fila, retorna apenas o tempo de chegada
	if len(p.Fila) == 0 {
		return tempo_total, 0
	}

	// Calcula o tempo total considerando todos os veículos na fila
	for i := range p.Fila {
		veiculo := p.Fila[i]
		tempo_carregamento := time.Duration(100-veiculo.Bateria) * time.Minute
		tempo_ate_posto_veiculo_fila := time.Duration(math.Abs(veiculo.Latitude-p.Latitude)+math.Abs(veiculo.Longitude-p.Longitude)) * 15 * time.Second

		// Adiciona o tempo de carregamento do veículo atual na fila
		tempo_total += tempo_carregamento

		// Se o veículo atual chegar antes que este veículo da fila, podemos inserir na posição i
		if tempoDistanciaVeiculo < tempo_ate_posto_veiculo_fila {
			posicao_na_fila = i
			break
		}
	}

	// Se não encontrou uma posição na fila, será o último
	if posicao_na_fila == -1 {
		posicao_na_fila = len(p.Fila)
	}

	return tempo_total, posicao_na_fila
}
