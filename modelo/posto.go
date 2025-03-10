package modelo

import (
	"fmt"
	"sync"
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
	fmt.Printf("Posto %s criado na localização em (%.6f, %.6f)",
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

	if !p.BombaOcupada {
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