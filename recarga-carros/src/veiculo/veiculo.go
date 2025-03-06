package veiculo

// definir uma localização fixa para os postos, com longitude e latitude.
// definir um valor aleatório inicialmente para os veículos e ir sempre alterando de 1 em 1 este valor. Caso a bateria atinga ponto crítico,
// fazer uma função para que a longitude e latitude do veículo seja alterada de 1 em 1 para a localização do posto recomendado.

type VehicleLocation struct {
	VeiculoID string  `json:"veiculo_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

