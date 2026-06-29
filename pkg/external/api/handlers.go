package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (s *Service) handleMessage(msg *tgbotapi.Message) {
	if msg.Text == "/start" {
		if err := s.switchState(msg.From.ID, selectSet); err != nil {
			s.logs.Printf("Error occurred while switching user state: %v\n", err)
			return
		}
		s.sendMessage(msg.Chat.ID, GreetingMsg)
		return
	}

	state, err := s.getState(msg.From.ID)
	if err != nil {
		s.logs.Printf("Error occurred while fetching user state: %v\n", err)
		return
	}

	switch state {
	case selectSet:
		if err := s.handleStateSet(msg); err != nil {
			s.logs.Printf("Error at the start:%v\n", err)
		}
	case selectDeletion:
		if err := s.handleStateDeletion(msg); err != nil {
			s.logs.Printf("Error at deletionList stage: %v\n", err)
		}
	case selectTitle:
		if err := s.handleStateTitle(msg); err != nil {
			s.logs.Printf("Error at selecting title stage: %v\n", err)
		}
	case selectName:
		if err := s.handleStateName(msg); err != nil {
			s.logs.Printf("Error at selecting name stage: %v\n", err)
		}
	case end:
		if err := s.handleStateEnd(msg); err != nil {
			s.logs.Printf("Error at the final stage: %v\n", err)
		}
	}
}

func (s *Service) handleStateSet(msg *tgbotapi.Message) error {
	if msg.Sticker == nil {
		s.sendMessage(msg.Chat.ID, ErrNoSetSelectedMsg)
		return fmt.Errorf("No sticker at the start error")
	}

	s.logs.Printf("User %s sent a sticker, starting spizding process\n", msg.From.UserName)

	if err := s.switchState(msg.From.ID, selectDeletion); err != nil {
		return fmt.Errorf("Error occurred while switching state: %w", err)
	}

	if err := s.addStickerSet(msg.From.ID, msg.Sticker.SetName); err != nil {
		return fmt.Errorf("Error occurred while adding sticker set: %w", err)
	}

	s.sendMessage(msg.Chat.ID, SetAcquiredMsg, msg.Sticker.SetName)
	return nil
}

func (s *Service) handleStateDeletion(msg *tgbotapi.Message) error {
	if msg.Text == "/cancel" {
		s.flushSelection(msg.From.ID)
	}

	if msg.Text == "/done" {
		deletionlist, err := s.getDeletionList(msg.From.ID)
		if err != nil {
			return fmt.Errorf("Error occured while fetching deletion list: %w", err)
		}
		if len(deletionlist) == 0 {
			s.sendMessage(msg.Chat.ID, ErrEmptyDeletionListMsg)
			return fmt.Errorf("Empty deletion list error")
		}

		if err := s.switchState(msg.From.ID, selectTitle); err != nil {
			return fmt.Errorf("Error occurred while switching state: %w", err)
		}
		s.sendMessage(msg.Chat.ID, SelectTitleMsg)
		return nil
	}

	if msg.Sticker == nil {
		s.sendMessage(msg.Chat.ID, ErrDoneOrStickerExpectedMsg)
		return fmt.Errorf("Sticker not expected error")
	}

	s.logs.Printf("User %s sent a sticker, adding to deletion list\n", msg.From.UserName)

	if err := s.addToDeletionList(msg.From.ID, msg.Sticker.FileID); err != nil {
		return fmt.Errorf("Error occurred while adding sticker to deletion list: %w", err)
	}

	s.sendMessage(msg.Chat.ID, AddedStickerMsg)
	return nil
}

func (s *Service) handleStateTitle(msg *tgbotapi.Message) error {
	if msg.Text == "/cancel" {
		s.flushSelection(msg.From.ID)
	}

	if msg.Sticker != nil {
		s.sendMessage(msg.Chat.ID, ErrStickerNotExpected)
		return fmt.Errorf("Sticker not expected error")
	}

	if len(msg.Text) > 64 {
		s.sendMessage(msg.Chat.ID, ErrTitleTooLong)
		return fmt.Errorf("Title too long error")
	}

	if err := s.addTitle(msg.From.ID, msg.Text); err != nil {
		return fmt.Errorf("Error occured while adding title: %w", err)
	}

	if err := s.switchState(msg.From.ID, selectName); err != nil {
		return fmt.Errorf("Error occured while switching state: %w", err)
	}

	s.sendMessage(msg.Chat.ID, SelectNameMsg)
	return nil
}

func (s *Service) handleStateName(msg *tgbotapi.Message) error {
	if msg.Text == "/cancel" {
		s.flushSelection(msg.From.ID)
	}

	if msg.Sticker != nil {
		s.sendMessage(msg.Chat.ID, ErrStickerNotExpected)
		return fmt.Errorf("Sticker not expected error")
	}

	if len(msg.Text+"_by_"+s.bot.Self.UserName) > 64 {
		s.sendMessage(msg.Chat.ID, ErrNameTooLong, s.bot.Self.UserName)
		return fmt.Errorf("Title too long error")
	}

	if !isValidName(msg.Text) {
		s.sendMessage(msg.Chat.ID, ErrNameInvalid)
		return fmt.Errorf("User sent invalid name")
	}

	if err := s.addName(msg.From.ID, msg.Text); err != nil {
		return fmt.Errorf("Error occured while adding name: %w", err)
	}

	if err := s.switchState(msg.From.ID, end); err != nil {
		return fmt.Errorf("Error occured while switching state: %w\n", err)
	}

	s.sendMessage(msg.Chat.ID, ConfirmMsg)
	return nil
}

func (s *Service) handleStateEnd(msg *tgbotapi.Message) error {
	if msg.Text == "/cancel" {
		s.flushSelection(msg.From.ID)
	}

	if msg.Sticker != nil {
		s.sendMessage(msg.Chat.ID, ErrStickerNotExpected)
		return fmt.Errorf("Sticker not expected error")
	}

	if msg.Text == "/confirm" {
		if err := s.process(msg); err != nil {
			return fmt.Errorf("Error while proccesing: %w", err)
		}
	}
	return nil
}

type addStickerToSetConfig struct {
	UserID  int64        `json:"user_id"`
	Name    string       `json:"name"`
	Sticker InputSticker `json:"sticker"`
}

type createNewStickerSetConfig struct {
	UserID   int64          `json:"user_id"`
	Name     string         `json:"name"`
	Title    string         `json:"title"`
	Stickers []InputSticker `json:"stickers"`
}

type InputSticker struct {
	Sticker    string   `json:"sticker"`
	Format     string   `json:"format"`
	Emoji_list []string `json:"emoji_list"`
}

func (s *Service) process(msg *tgbotapi.Message) error {
	title, name, oldName, deletionList, err := s.getAll(msg.From.ID)
	if err != nil {
		return err
	}
	set, err := s.bot.GetStickerSet(tgbotapi.GetStickerSetConfig{Name: oldName})
	if err != nil {
		return fmt.Errorf("Error occurred while fetching sticker set: %v\n", err)
	}

	newSet, err := s.newStickers(set.Stickers, deletionList)
	if err != nil {
		return fmt.Errorf("Error occurred while creating new sticker set config: %v\n", err)
	}

	name = name + "_by_" + s.bot.Self.UserName
	if err := s.requestNewStickerSet(msg.From.ID, name, title, newSet); err != nil {
		return fmt.Errorf("Error occurred while requesting new sticker set: %v\n", err)
	}

	s.logs.Printf("Successfully created new sticker set for user %s\n", msg.From.UserName)

	stickerSet, err := s.bot.GetStickerSet(tgbotapi.GetStickerSetConfig{Name: name})
	if err != nil {
		return fmt.Errorf("Error occurred while fetching new sticker set: %v\n", err)
	}

	s.sendMessage(msg.Chat.ID, "Your new sticker set '%s' has been created! You can find it here: https://t.me/addstickers/%s", stickerSet.Name, stickerSet.Name)
	return nil
}

func (s *Service) newStickers(stickers []tgbotapi.Sticker, deletionList []string) ([]InputSticker, error) {
	newSet := make([]InputSticker, 0, len(stickers))

	counter := 0
	for _, sticker := range stickers {
		if !slices.Contains(deletionList, sticker.FileID) {
			newSet = append(newSet, InputSticker{
				Sticker:    sticker.FileID,
				Format:     "static",
				Emoji_list: []string{sticker.Emoji},
			})
			counter++
		}

	}
	return newSet, nil
}

func (s *Service) requestNewStickerSet(userID int64, name, title string, stickers []InputSticker) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/createNewStickerSet", s.bot.Token)

	if len(stickers) > 120 {
		stickers = stickers[:120]
	}

	var leftovers []InputSticker
	if len(stickers) > 50 {
		leftovers = stickers[50:]
		stickers = stickers[:50]
	}
	config := createNewStickerSetConfig{
		UserID:   userID,
		Name:     name,
		Title:    title,
		Stickers: stickers,
	}

	s.logs.Printf("Trying to create stickerset: %s\n", name)

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("Error occurred while marshaling new sticker set data: %v\n", err)
	}

	creResp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("Error occurred during post request: %v\n", err)
	}
	defer creResp.Body.Close()

	if creResp.StatusCode != http.StatusOK {
		errorBody := new(bytes.Buffer)
		errorBody.ReadFrom(creResp.Body)
		return fmt.Errorf("Received non-OK response from Telegram API: %s\nResponse body: %s\n", creResp.Status, errorBody.String())
	}

	if len(leftovers) <= 0 {
		return nil
	}

	url = fmt.Sprintf("https://api.telegram.org/bot%s/addStickerToSet", s.bot.Token)
	for i := range leftovers {
		addConfig := addStickerToSetConfig{
			UserID:  userID,
			Name:    name,
			Sticker: leftovers[i],
		}

		jsonData, err := json.Marshal(addConfig)
		if err != nil {
			return fmt.Errorf("Error occurred while marshaling add sticker data: %v\n", err)
		}

		addResp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("Error occurred during post request: %v\n", err)
		}

		if addResp.StatusCode != http.StatusOK {
			errorBody := new(bytes.Buffer)
			errorBody.ReadFrom(addResp.Body)
			addResp.Body.Close()
			return fmt.Errorf("Received non-OK response from Telegram API: %s\nResponse body: %s\n", addResp.Status, errorBody.String())
		}

		addResp.Body.Close()
	}
	return nil
}
