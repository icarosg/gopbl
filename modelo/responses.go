package modelo

type RecomendadoResponse struct{
	ID_posto string `json:"id_posto"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Posicao_na_fila int `json:"posicao_na_fila"`
}