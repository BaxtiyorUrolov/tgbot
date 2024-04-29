package bot

import (
	"io/ioutil"
	"log"
	"tgbot/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// SelectBarber sends a photo with an inline keyboard to the specified chat.
func SelectBarber(chatID int64) {
	// Path to your photo
	photoFilePath := "./photo.jpg"
	caption := "Iltimoss, sartaroshni tanlang:"

	// Bot instance
	botInstance := config.GetBot()
	if botInstance == nil {
		log.Println("Bot instance is nil")
		return
	}

	// Open and read the photo file
	photoFileBytes, err := readFileBytes(photoFilePath)
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

	// Create an inline keyboard with two rows of buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		// First row of buttons
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Baxtiyor", "Baxtiyor"),
			tgbotapi.NewInlineKeyboardButtonData("Ali", "Ali"),
		),
		// Second row of buttons
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Aziz", "Aziz"),
			tgbotapi.NewInlineKeyboardButtonData("Obid", "Obid"),
		),
	)

	// Attach the inline keyboard to the message
	msg.ReplyMarkup = &keyboard

	// Send the photo message
	_, err = botInstance.Send(msg)
	if err != nil {
		log.Printf("Failed to send photo: %v", err)
	}
}

// readFileBytes reads and returns the bytes of a file.
func readFileBytes(filePath string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
