package models

type User struct {
	ID    int64
	Name  string
	Phone string
}

type ForClients struct {
	ClientType   string
	TimeDuration int
	Price        int
}
