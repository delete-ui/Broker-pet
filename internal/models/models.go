package models

type Deal struct {
	Id       int64   `json:"id"`
	Title    string  `json:"title"`
	Expenses float64 `json:"expenses"`
	Profit   float64 `json:"profit"`
	Status   string  //"processed" or "not processed"
}

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"` //in a real project, store a hash
}

type NewUserResponse struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
}

type LogUserResponse struct {
	Username string `json:"username"`
	Status   string `json:"status"`
}

type ProfitSQLDeal struct {
	Id        int64
	DealId    int64
	AllProfit float64
}
