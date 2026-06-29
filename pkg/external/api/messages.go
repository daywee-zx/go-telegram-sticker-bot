package api

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TODO family-friendly messages
const (
	GreetingMsg     = "Hello! I am a bot created to reappropriate any stickerpack you want without stickers you don't need and a title of your choice!\n\n To get started, send me a sticker from the stickerpack you want to claim for yourself.\n\nAt any point you can send me /cancel to get back to the start"
	SetAcquiredMsg  = "Sticker set '%s' selected. To continue, send me the stickers you want to delete from the set."
	AddedStickerMsg = "Sticker added to deletion list! Send more stickers to add them to the list, or send /done when you're finished."
	SelectTitleMsg  = "Great! Now you will be required to send me the new sticker set's title and a short name, used in links and such.\nFor example,\nTitle: My sticker set\nName: mystickerset\n\nLet's start with the title you want:"
	SelectNameMsg   = "Now for the last - send me a short name for your sticker set. Only latin, digits and underscores, please.\n Due to Telegram rules, I will have to add my name at the end as well :)"
	ConfirmMsg      = "That's it! Take a look at the info in previous messages. If you want to change something, send me /cancel and we'll do it all over again.\nIf not, send me /confirm and I'll get your sticker set spizdinated!"
	CancelMsg       = "You canceled the operation. Let's start from the beginning!\n\nTo get started, send me a sticker from the set you want to rightfully (or not) take!"

	ErrNameTooLong              = "Name you sent is too long. Do not forget I have to add \"_by_%s\". All of it has to be no longer that 64 symbols"
	ErrNameInvalid              = "Please use only english letters, digits or underscores."
	ErrTitleTooLong             = "Title you sent is too long. It has to be no longer than 64 symbols"
	ErrStickerNotExpected       = "I was not expecting a sticker at this moment. Check the latest instruction"
	ErrDoneOrStickerExpectedMsg = "I was expecting a sticker, but you sent something else. Please send me a sticker from the set you want to spizdit, or send /done when you're finished."
	ErrNoSetSelectedMsg         = "You haven't selected a sticker set yet. Please send me a sticker from the set you want to spizdit to get started."
	ErrEmptyDeletionListMsg     = "Your deletion list is empty. Please send me the stickers you want to delete from the set, or send /done when you're finished."
)

// TODO use more to communicate status
func (s *Service) sendMessage(chatID int64, text string, options ...any) {
	if len(options) > 0 {
		text = fmt.Sprintf(text, options...)
	}
	reply := tgbotapi.NewMessage(chatID, text)
	s.bot.Send(reply)
}
