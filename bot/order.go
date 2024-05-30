package bot

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"

	"tgbot/models"
	"tgbot/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var userStates = struct {
	sync.RWMutex
	m map[int64]int
}{m: make(map[int64]int)}

func SelectBarber(chatID int64, botInstance *tgbotapi.BotAPI) {
	// Rasmning yo'lini belgilash
	photoFilePath := "./photo.jpg"
	caption := "Iltimos, sartaroshni tanlang:"

	// Rasm faylini ochish va o'qish
	photoFileBytes, err := ioutil.ReadFile(photoFilePath)
	if err != nil {
		log.Printf("Rasm faylini o'qishda xatolik: %v", err)
		return
	}

	// Jo'natiladigan rasm uchun yangi rasm konfiguratsiyasi yaratish
	msg := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{
		Name:  "photo.jpg",
		Bytes: photoFileBytes,
	})

	// Rasm sarlavhasini sozlash
	msg.Caption = caption

	// Ikki qatorlik sartaroshni tanlash tugmalarini o'z ichiga olgan to'plam yaratish
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// Birinchi qator (sartarosh tanlash)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Baxtiyor", "select_date_Baxtiyor"),
			tgbotapi.NewInlineKeyboardButtonData("Ali", "select_date_Ali"),
		),
		// Ikkinchi qator (sartarosh tanlash)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Aziz", "select_date_Aziz"),
			tgbotapi.NewInlineKeyboardButtonData("Obid", "select_date_Obid"),
		),
	)

	// Sartarosh tanlash tugmalari bilan sartarosh tanlash to'plamini bog'lash
	msg.ReplyMarkup = &keyboard

	// Rasm xabarni jo'natish
	sentMessage, err := botInstance.Send(msg)
	if err != nil {
		log.Printf("Rasmni jo'natishda xatolik: %v", err)
	}

	// Logda sartarosh tanlanganda chiqariladigan xabar
	log.Println("Sartaroshni tanlash tugmalarini jo'natish uchun botdan so'roq jo'natildi")

	// Store the message ID in the user states
	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

// SelectDate allows the user to select a date after selecting a barber
func SelectDate(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, prevMessageID int) {
	// Ensure a barber name is provided before selecting a date
	if barberName == "" {
		log.Println("No barber selected. Skipping date selection.")
		return
	}

	// Delete the previous message
	deleteMessage := tgbotapi.NewDeleteMessage(chatID, prevMessageID)
	botInstance.Send(deleteMessage)

	fmt.Println("Delete")
	fmt.Println(deleteMessage)

	// Get today's date
	today := time.Now()
	// Get tomorrow's date
	tomorrow := today.AddDate(0, 0, 1)
	// Get the day after tomorrow's date
	dayAfterTomorrow := today.AddDate(0, 0, 2)

	// Create inline buttons for date selection with both barber name and date in callback data
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(today.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, today.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(tomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, tomorrow.Format("2006-01-02"))),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dayAfterTomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, dayAfterTomorrow.Format("2006-01-02"))),
		),
	)

	// Send date selection message
	dateSelectionMsg := tgbotapi.NewMessage(chatID, "Sana tanlang (kun.oy.yil):")
	dateSelectionMsg.ReplyMarkup = &keyboard
	sentMessage, err := botInstance.Send(dateSelectionMsg)
	if err != nil {
		log.Printf("Error sending date selection keyboard: %v", err)
		return
	}

	// Store the message ID in the user states
	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func SelectOrder(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, orderDate string, update tgbotapi.Update, prevMessageID int) {
	// Delete the previous message
	deleteDateMessage := tgbotapi.NewDeleteMessage(chatID, prevMessageID)
	botInstance.Send(deleteDateMessage)

	fmt.Println(deleteDateMessage)

	fmt.Println("Entering order selection")
	fmt.Println("Barber: ", barberName)
	fmt.Println("Date: ", orderDate)

	// Fetch existing orders for the selected barber and date
	order := models.GetOrders{
		BarberID: barberName,
		Date:     orderDate,
	}

	existingOrderTimes, err := storage.GetOrders(order)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		return
	}

	// Create a set of existing order times
	existingTimesSet := make(map[string]struct{})
	for _, time := range existingOrderTimes {
		existingTimesSet[time] = struct{}{}
	}

	// Define the available time slots
	timeSlots := []string{"9:00", "10:00", "11:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00"}

	// Create inline keyboard buttons for the time slots
	var rows [][]tgbotapi.InlineKeyboardButton
	for i := 0; i < len(timeSlots); i += 4 {
		var row []tgbotapi.InlineKeyboardButton
		for j := 0; j < 4 && i+j < len(timeSlots); j++ {
			timeSlot := timeSlots[i+j]
			if _, exists := existingTimesSet[timeSlot]; exists {
				row = append(row, tgbotapi.NewInlineKeyboardButtonData("❌", "X"))
			} else {
				callbackData := fmt.Sprintf("confirm_%s_%s_%s", barberName, orderDate, timeSlot)
				row = append(row, tgbotapi.NewInlineKeyboardButtonData(timeSlot, callbackData))
			}
		}
		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// Send a message to the user to select a time slot
	timeSelectionMsg := tgbotapi.NewMessage(chatID, "Navbat tanlang:")
	timeSelectionMsg.ReplyMarkup = &keyboard
	sentMessage, err := botInstance.Send(timeSelectionMsg)
	if err != nil {
		log.Printf("Error sending time selection keyboard: %v", err)
		return
	}

	// Store the message ID in the user states
	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func HandleConfirmation(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, orderDate string, orderTime string, update tgbotapi.Update, prevMessageID int) {
	callbackData := update.CallbackQuery.Data

	deleteMessage := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
		botInstance.Send(deleteMessage)

	if strings.HasPrefix(callbackData, "confirm_") {
		// Extract barber name, order date, and order time from the callback data
		data := strings.Split(callbackData, "_")
		if len(data) < 4 {
			log.Println("Invalid callback data for confirmation")
			return
		}
		barberName := data[1]
		orderDate := data[2]
		orderTime := data[3]

		// Create inline buttons for confirmation
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Tasdiqlash", fmt.Sprintf("book_%s_%s_%s", barberName, orderDate, orderTime)),
				tgbotapi.NewInlineKeyboardButtonData("Ortga qaytish", "back"),
			),
		)

		// Send a message asking for confirmation
		confirmationMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Siz %s kuni %s vaqtiga navbat olishni xohlaysizmi?", orderDate, orderTime))
		confirmationMsg.ReplyMarkup = &keyboard
		sentMessage, err := botInstance.Send(confirmationMsg)
		if err != nil {
			log.Printf("Confirmation message sending error: %v", err)
			return
		}

		// Store the message ID in the user states
		userStates.Lock()
		userStates.m[chatID] = sentMessage.MessageID
		userStates.Unlock()
	} else if strings.HasPrefix(callbackData, "book_") {
		// Extract barber name, order date, and order time from the callback data
		data := strings.Split(callbackData, "_")
		if len(data) < 4 {
			log.Println("Invalid callback data for booking")
			return
		}
		barberName := data[1]
		orderDate := data[2]
		orderTime := data[3]

		// Insert the booking details into the orders table
		order := models.Order{
			BarberName: barberName,
			UserID:     update.CallbackQuery.From.ID,
			OrderDate:  orderDate,
			OrderTime:  orderTime,
			Status:     "in_process",
		}

		if err := storage.SaveOrder(order); err != nil {
			log.Printf("Error saving order: %v", err)
			return
		}

		// Send a confirmation message
		// Delete the previous message
		userStates.RLock()
		//lastMessageID := userStates.m[chatID]
		userStates.RUnlock()
		deleteMessage := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
		botInstance.Send(deleteMessage)

		confirmationMsg := tgbotapi.NewMessage(chatID, "Navbat muvaffaqiyatli saqlandi!")
		sentMessage, err := botInstance.Send(confirmationMsg)
		if err != nil {
			log.Printf("Error sending confirmation message: %v", err)
		}

		// Store the confirmation message ID in the user states
		userStates.Lock()
		userStates.m[chatID] = sentMessage.MessageID
		userStates.Unlock()
	} else if callbackData == "back" {
		// Handle the "Back" button press by redisplaying the time slots
		userStates.RLock()
		lastMessageID := userStates.m[chatID]
		userStates.RUnlock()
		SelectOrder(chatID, botInstance, barberName, orderDate, update, lastMessageID)
	}
}
