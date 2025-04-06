package modelo

type RecomendadoResponse struct {
	ID_posto        string  `json:"id_posto"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

type PagamentoJson struct {
	Veiculo  Veiculo `json:"id_veiculo"`
	Valor    float64 `json:"valor"`
	ID_posto string  `json:"id_posto"`
}

type ReservarVagaJson struct {
	Veiculo Veiculo `json:"id_veiculo"`
}

type RetornarVagaJson struct {
	Posto Posto `json:"id_posto"`
}

type AtualizarPosicaoNaFila struct {
	Veiculo  Veiculo `json:"veiculo"`
	ID_posto string  `json:"id_posto"`
}

type RetornarAtualizarPosicaoFila struct {
	Veiculo Veiculo `json:"veiculo"`
	Posto   Posto   `json:"posto"`
}
