package models

type Deal struct {
	Id       int64   `json:"id"`
	Title    string  `json:"title"`
	Expenses float64 `json:"expenses"`
	Profit   float64 `json:"profit"`
	Status   string  //"processed" or "not processed"
}
