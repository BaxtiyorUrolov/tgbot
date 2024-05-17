package storage

import (
	"database/sql"
	"fmt"
	"log"
	"tgbot/config"
	"tgbot/models"
)

func GetUserFromDB(userID int64) *models.User {
	var (
		user     models.User
		nameStr  sql.NullString
		phoneStr sql.NullString
	)

	fmt.Println("User: ", userID) // For debugging: print the userID being fetched

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
		return nil // Return nil if user not found or other error occurred
	}

	fmt.Println("id:", user.ID) // For debugging: print the user ID after scan

	if nameStr.Valid {
		user.Name = nameStr.String
	}

	if phoneStr.Valid {
		user.Phone = phoneStr.String
	}

	return &user // Return pointer to populated User struct
}

func SaveUserToDB(user *models.User) {

	fmt.Println("Saving user: ", user.Name)

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

func GetOrders(order models.GetOrders) (string, error) {
	db := config.GetDB()

	rows, err := db.Query("SELECT order_time FROM orders WHERE order_date = $1 AND barber_id = $2 AND status = 'in_process'", order.Date, order.BarberID)
	if err != nil {
		log.Printf("Database query error: %v", err)
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.Scan(&order.Time)
		if err != nil {
			log.Printf("Error scanning rows: %v", err)
			return "", err
		}
	} else {
		return "", nil
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return "", err
	}

	return order.Time, nil
}
