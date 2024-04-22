package bot

import (
	"log"
	"tgbot/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SendInitialOptions(chatID int64) {
	aloqaButton := tgbotapi.NewKeyboardButton("Aloqa")
	navbatButton := tgbotapi.NewKeyboardButton("Navbat olish")

	replyMarkup := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(aloqaButton, navbatButton),
	)

	botInstance := config.GetBot()

	msgSend := tgbotapi.NewMessage(chatID, "Iltimos, tanlang:")
	msgSend.ReplyMarkup = &replyMarkup

	if _, err := botInstance.Send(msgSend); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
