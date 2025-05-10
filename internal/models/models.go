package models

type Deal struct {
	Id       int64
	Title    string
	Expenses float64
	Profit   float64
	Status   bool
}
