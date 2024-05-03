package bot

import (
	"fmt"
	"io/ioutil"
	"log"
	"tgbot/config"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// SelectBarber sends a photo with an inline keyboard to the specified chat.
func SelectBarber(chatID int64) {
	// Path to your photo
	photoFilePath := "./photo.jpg"
	caption := "Iltimos, sartaroshni tanlang:"

	// Bot instance
	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	// Open and read the photo file
	photoFileBytes, err := ioutil.ReadFile(photoFilePath)
	if err != nil {
		log.Printf("Failed to read photo file: %v", err)
		return
	}

	// Create a new photo configuration for sending a photo
	msg := tgbotapi.NewPhotoUpload(chatID, tgbotapi.FileBytes{
		Name:  "photo.jpg",
		Bytes: photoFileBytes,
	})

	// Set the photo caption
	msg.Caption = caption

	// Create an inline keyboard with barber selection buttons in two rows
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// First row of buttons (barber selection)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Baxtiyor", "select_barber Baxtiyor"),
			tgbotapi.NewInlineKeyboardButtonData("Ali", "select_barber Ali"),
		),
		// Second row of buttons (barber selection)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Aziz", "select_barber Aziz"),
			tgbotapi.NewInlineKeyboardButtonData("Obid", "select_barber Obid"),
		),
	)

	// Attach the inline keyboard to the message
	msg.ReplyMarkup = &keyboard

	// Send the photo message
	if _, err := botInstance.Send(msg); err != nil {
		log.Printf("Failed to send photo: %v", err)
	}

	// Get user selection after sending barber options
	getSelectedBarber(chatID)
}

// getSelectedBarber gets the selected barber from the user.
func getSelectedBarber(chatID int64) {
	// Bot instance
	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	// Listen for user selection
	updates, err := botInstance.GetUpdatesChan(tgbotapi.NewUpdate(0))
	if err != nil {
		log.Printf("Failed to get updates channel: %v", err)
		return
	}

	// Wait for user selection
	for update := range updates {
		if update.CallbackQuery != nil {
			// Get selected barber name
			selectedBarberName := update.CallbackQuery.Data
			log.Printf("Selected barber: %s", selectedBarberName)

			// Call SelectDate function with selected barber name
			SelectDate(chatID, selectedBarberName)
			break
		}
	}
}

// SelectDate prompts the user to select a date after selecting a barber.
func SelectDate(chatID int64, barberName string) {
	// If barberName is empty, do not prompt for date selection
	if barberName == "" {
		log.Println("Barber is not selected. Skipping date selection.")
		return
	}

	// Get today's date
	today := time.Now()
	// Get tomorrow's date
	tomorrow := today.AddDate(0, 0, 1)
	// Get the day after tomorrow's date
	dayAfterTomorrow := today.AddDate(0, 0, 2)

	// Create an inline keyboard with date selection buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// First row of buttons (date selection)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(today.Format("02.01.2006"), fmt.Sprintf("select_date %s %s", today.Format("2006-01-02"), barberName)),
			tgbotapi.NewInlineKeyboardButtonData(tomorrow.Format("02.01.2006"), fmt.Sprintf("select_date %s %s", tomorrow.Format("2006-01-02"), barberName)),
		),
		// Second row of buttons (date selection)
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(dayAfterTomorrow.Format("02.01.2006"), fmt.Sprintf("select_date %s %s", dayAfterTomorrow.Format("2006-01-02"), barberName)),
		),
	)

	// Send the inline keyboard for date selection
	dateSelectionMsg := tgbotapi.NewMessage(chatID, "Sana tanlang:")
	dateSelectionMsg.ReplyMarkup = &keyboard
	if _, err := config.GetBot().Send(dateSelectionMsg); err != nil {
		log.Printf("Failed to send date selection keyboard: %v", err)
	}
}
