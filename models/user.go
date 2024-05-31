package models

type Order struct {
	ID         string
	BarberName string
	UserID     int
	OrderTime  string
	OrderDate  string
	Status     string
}

type User struct {
	ID    int64
	Name  string
	Phone string
}

type Barber struct {
	ID    int64
	Name  string
	UserName string
	Phone string
	Admin string
}
