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
	photoFilePath := "./photo.jpg"
	caption := "Please select a barber:"

	photoFileBytes, err := ioutil.ReadFile(photoFilePath)
	if err != nil {
		log.Printf("Error reading photo file: %v", err)
		return
	}

	msg := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{
		Name:  "photo.jpg",
		Bytes: photoFileBytes,
	})
	msg.Caption = caption

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Baxtiyor", "select_date_Baxtiyor"),
			tgbotapi.NewInlineKeyboardButtonData("Ali", "select_date_Ali"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Aziz", "select_date_Aziz"),
			tgbotapi.NewInlineKeyboardButtonData("Obid", "select_date_Obid"),
		),
	)
	msg.ReplyMarkup = &keyboard

	sentMessage, err := botInstance.Send(msg)
	if err != nil {
		log.Printf("Error sending photo: %v", err)
	}

	log.Println("Sent barber selection buttons")

	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func SelectDate(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, prevMessageID int) {
	if barberName == "" {
		log.Println("No barber selected. Skipping date selection.")
		return
	}

	deleteMessage := tgbotapi.NewDeleteMessage(chatID, prevMessageID)
	botInstance.Send(deleteMessage)

	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)
	dayAfterTomorrow := today.AddDate(0, 0, 2)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(today.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, today.Format("2006-01-02"))),
			tgbotapi.NewInlineKeyboardButtonData(tomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, tomorrow.Format("2006-01-02"))),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dayAfterTomorrow.Format("02.01.2006"), fmt.Sprintf("datte_%s_%s", barberName, dayAfterTomorrow.Format("2006-01-02"))),
		),
	)

	dateSelectionMsg := tgbotapi.NewMessage(chatID, "Please select a date (day.month.year):")
	dateSelectionMsg.ReplyMarkup = &keyboard
	sentMessage, err := botInstance.Send(dateSelectionMsg)
	if err != nil {
		log.Printf("Error sending date selection keyboard: %v", err)
		return
	}

	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func SelectOrder(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, orderDate string, update tgbotapi.Update, prevMessageID int) {
	deleteMessage := tgbotapi.NewDeleteMessage(chatID, prevMessageID)
	botInstance.Send(deleteMessage)

	order := models.GetOrders{
		BarberID: barberName,
		Date:     orderDate,
	}

	existingOrderTimes, err := storage.GetOrders(order)
	if err != nil {
		log.Printf("Error fetching orders: %v", err)
		return
	}

	existingTimesSet := make(map[string]struct{})
	for _, time := range existingOrderTimes {
		existingTimesSet[time] = struct{}{}
	}

	timeSlots := []string{"9:00", "10:00", "11:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00", "21:00"}

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

	timeSelectionMsg := tgbotapi.NewMessage(chatID, "Please select a time:")
	timeSelectionMsg.ReplyMarkup = &keyboard
	sentMessage, err := botInstance.Send(timeSelectionMsg)
	if err != nil {
		log.Printf("Error sending time selection keyboard: %v", err)
		return
	}

	userStates.Lock()
	userStates.m[chatID] = sentMessage.MessageID
	userStates.Unlock()
}

func HandleConfirmation(chatID int64, botInstance *tgbotapi.BotAPI, barberName string, orderDate string, orderTime string, update tgbotapi.Update, prevMessageID int) {
	callbackData := update.CallbackQuery.Data

	deleteMessage := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
	botInstance.Send(deleteMessage)

	if strings.HasPrefix(callbackData, "confirm_") {
		data := strings.Split(callbackData, "_")
		if len(data) < 4 {
			log.Println("Invalid callback data for confirmation")
			return
		}
		barberName := data[1]
		orderDate := data[2]
		orderTime := data[3]

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Confirm", fmt.Sprintf("book_%s_%s_%s", barberName, orderDate, orderTime)),
				tgbotapi.NewInlineKeyboardButtonData("Back", "back"),
			),
		)

		confirmationMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Do you want to book on %s at %s?", orderDate, orderTime))
		confirmationMsg.ReplyMarkup = &keyboard
		sentMessage, err := botInstance.Send(confirmationMsg)
		if err != nil {
			log.Printf("Error sending confirmation message: %v", err)
			return
		}

		userStates.Lock()
		userStates.m[chatID] = sentMessage.MessageID
		userStates.Unlock()
	} else if strings.HasPrefix(callbackData, "book_") {
		data := strings.Split(callbackData, "_")
		if len(data) < 4 {
			log.Println("Invalid callback data for booking")
			return
		}
		barberName := data[1]
		orderDate := data[2]
		orderTime := data[3]

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

		userStates.RLock()
		userStates.RUnlock()
		deleteMessage := tgbotapi.NewDeleteMessage(chatID, update.CallbackQuery.Message.MessageID)
		botInstance.Send(deleteMessage)

		confirmationMsg := tgbotapi.NewMessage(chatID, "Order successfully saved!")
		sentMessage, err := botInstance.Send(confirmationMsg)
		if err != nil {
			log.Printf("Error sending confirmation message: %v", err)
		}

		userStates.Lock()
		userStates.m[chatID] = sentMessage.MessageID
		userStates.Unlock()

		sendOrderDetailsToChannel(order, botInstance)
	} else if callbackData == "back" {
		userStates.RLock()
		lastMessageID := userStates.m[chatID]
		userStates.RUnlock()
		SelectOrder(chatID, botInstance, barberName, orderDate, update, lastMessageID)
	}
}

func sendOrderDetailsToChannel(order models.Order, botInstance *tgbotapi.BotAPI) {
	message := fmt.Sprintf("New Order:\nBarber: %s\nDate: %s\nTime: %s\nUser ID: %d", order.BarberName, order.OrderDate, order.OrderTime, order.UserID)
	channelMessage := tgbotapi.NewMessageToChannel("@BMC_Director", message)
	_, err := botInstance.Send(channelMessage)
	if err != nil {
		log.Printf("Error sending order details to channel: %v", err)
	}
}
