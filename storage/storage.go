package storage

import (
	"database/sql"
	"fmt"
	"log"
	"tgbot/config"
	"tgbot/models"

	"github.com/google/uuid"
)

func GetUserFromDB(userID int64) *models.User {
	var (
		user     models.User
		nameStr  sql.NullString
		phoneStr sql.NullString
	)

	fmt.Println("User: ", userID) // Debug: userIDni chiqarish

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return nil
	}

	row := db.QueryRow("SELECT user_id, name, phone FROM users WHERE user_id = $1", userID)
	err := row.Scan(&user.ID, &nameStr, &phoneStr)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying user from database: %v", err)
		} else {
			log.Printf("User with ID %d not found in database", userID)
		}
		fmt.Println("NIL")
		return nil // User topilmasa yoki boshqa xatolik yuz bersa, nil qaytarish
	}

	fmt.Println("id:", user.Name) // Debug: user IDni chiqarish

	if nameStr.Valid {
		user.Name = nameStr.String
	}

	fmt.Println("Name: ", user.Name) // Debug: user nomini chiqarish

	if phoneStr.Valid {
		user.Phone = phoneStr.String
	}

	return &user // Populated User struct pointerini qaytarish
}

func SaveUserToDB(user *models.User) {
	fmt.Println("Saving user: ", user.Name)
	fmt.Println("Phone: ", user.Phone)

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return
	}

	_, err := db.Exec("UPDATE users SET name = $2, phone = $3 WHERE user_id = $1", user.ID, user.Name, user.Phone)
	if err != nil {
		log.Printf("Error updating user in database: %v", err)
	}
}

func GetOrders(order models.GetOrders) ([]string, error) {
	db := config.GetDB()
	var orders []string

	rows, err := db.Query("SELECT order_time FROM orders WHERE order_date = $1 AND barber_name = $2 AND status = 'in_process'", order.Date, order.BarberID)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var orderTime string
		if err := rows.Scan(&orderTime); err != nil {
			log.Printf("Error scanning rows: %v", err)
			return nil, err
		}
		orders = append(orders, orderTime)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return nil, err
	}

	return orders, nil
}

func HasInProcessOrder(userID int64) (bool, error) {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return false, fmt.Errorf("database connection is nil")
	}

	// Agar userID 0 bo'lsa, true qaytarish
	if userID == 0 {
		return true, nil
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = 'in_process'", userID).Scan(&count)
	if err != nil {
		log.Printf("Error querying in-process orders: %v", err)
		return false, err
	}

	return count > 0, nil
}



func SaveOrder(order models.Order) error {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	_, err := db.Exec("INSERT INTO orders (id, barber_name, user_id, order_time, order_date, status) VALUES ($1, $2, $3, $4, $5, $6)",
		uuid.New().String(), order.BarberName, order.UserID, order.OrderTime, order.OrderDate, order.Status)
	if err != nil {
		log.Printf("Error inserting order into database: %v", err)
		return err
	}
	return nil
}

func GetBarber(ID int) string {

	var name string

	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return ""
	}
	
	row := db.QueryRow("SELECT name FROM barbers WHERE id = $1", ID)
	err := row.Scan(&name)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error querying user from database: %v", err)
		} else {
			log.Printf("User with ID %d not found in database", ID)
		}
		fmt.Println("NIL")
		return "" // User topilmasa yoki boshqa xatolik yuz bersa, nil qaytarish
	}

	return name
}

func GetUserIDByOrderDetails(barberName, orderDate, orderTime string) int64 {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return 0
	}

	var userID int64
	err := db.QueryRow("SELECT user_id FROM orders WHERE barber_name = $1 AND order_date = $2 AND order_time = $3", barberName, orderDate, orderTime).Scan(&userID)
	if err != nil {
		log.Printf("Error querying user_id by order details: %v", err)
		return 0
	}

	return userID
}

func AddBarber(barber models.Barber) error {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	_, err := db.Exec("INSERT INTO barbers (id, name, user_name, phone) VALUES ($1, $2, $3, $4)",
		barber.ID, barber.Name, barber.UserName, barber.Phone)
	if err != nil {
		log.Printf("Error inserting barber into database: %v", err)
		return err
	}
	return nil
}

func GetBarbers() ([]models.Barber, error) {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return nil, fmt.Errorf("database connection is nil")
	}

	rows, err := db.Query("SELECT id, name, user_name, phone, admin FROM barbers")
	if err != nil {
		log.Printf("Database query error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var barbers []models.Barber
	for rows.Next() {
		var barber models.Barber
		if err := rows.Scan(&barber.ID, &barber.Name, &barber.UserName, &barber.Phone, &barber.Admin); err != nil {
			log.Printf("Error scanning rows: %v", err)
			return nil, err
		}
		barbers = append(barbers, barber)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return nil, err
	}

	return barbers, nil
}

func DeleteOrder(barberName, orderDate, orderTime string) error {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	_, err := db.Exec("DELETE FROM orders WHERE barber_name = $1 AND order_date = $2 AND order_time = $3", barberName, orderDate, orderTime)
	if err != nil {
		log.Printf("Error deleting order: %v", err)
		return err
	}

	return nil
}

func CompleteOrder(barberName, orderDate, orderTime string) error {
	db := config.GetDB()
	if db == nil {
		log.Println("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	_, err := db.Exec("UPDATE orders SET status = 'done' WHERE barber_name = $1 AND order_date = $2 AND order_time = $3", barberName, orderDate, orderTime)
	if err != nil {
		log.Printf("Error completing order: %v", err)
		return err
	}

	return nil
}


