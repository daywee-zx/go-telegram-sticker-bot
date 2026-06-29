package api

import (
	"fmt"
	"strconv"
)

type userState int

const (
	selectSet userState = iota
	selectDeletion
	selectTitle
	selectName
	end
)

func (s *Service) switchState(userID int64, newState userState) error {
	key := fmt.Sprintf("user:%d:state", userID)
	state := strconv.Itoa(int(newState))
	err := s.rdb.Set(s.ctx, key, state, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to switch user state: %w", err)
	}
	return nil
}

func (s *Service) getState(userID int64) (userState, error) {
	key := fmt.Sprintf("user:%d:state", userID)
	state, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		return -1, fmt.Errorf("failed to get user state: %w", err)
	}
	st, err := strconv.Atoi(state)
	if err != nil {
		return -1, fmt.Errorf("failed to get user state: %w", err)
	}
	return userState(st), nil
}

func (s *Service) addToDeletionList(userID int64, stickerFileID string) error {
	key := fmt.Sprintf("user:%d:deletion_list", userID)
	return s.rdb.SAdd(s.ctx, key, stickerFileID).Err()
}

func (s *Service) getDeletionList(userID int64) ([]string, error) {
	key := fmt.Sprintf("user:%d:deletion_list", userID)
	deletionList, err := s.rdb.SMembers(s.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get deletion list: %w", err)
	}
	return deletionList, nil
}

func (s *Service) addStickerSet(userID int64, setName string) error {
	key := fmt.Sprintf("user:%d:sticker_set", userID)
	return s.rdb.Set(s.ctx, key, setName, 0).Err()
}

func (s *Service) getStickerSet(userID int64) (string, error) {
	key := fmt.Sprintf("user:%d:sticker_set", userID)
	setName, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get sticker set: %w", err)
	}
	return setName, nil
}

func (s *Service) addTitle(userID int64, t string) error {
	key := fmt.Sprintf("user:%d:title", userID)
	return s.rdb.Set(s.ctx, key, t, 0).Err()
}

func (s *Service) getTitle(userID int64) (string, error) {
	key := fmt.Sprintf("user:%d:title", userID)
	title, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get new sticker set title: %w", err)
	}
	return title, nil
}

func (s *Service) addName(userID int64, n string) error {
	key := fmt.Sprintf("user:%d:name", userID)
	return s.rdb.Set(s.ctx, key, n, 0).Err()
}

func (s *Service) getName(userID int64) (string, error) {
	key := fmt.Sprintf("user:%d:name", userID)
	name, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get sticker set name: %w", err)
	}
	return name, nil
}

func (s *Service) flushSelection(userID int64) error {
	key := fmt.Sprintf("user:%d:deletion_list", userID)
	if err := s.rdb.Del(s.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to flush deletion list: %w", err)
	}

	key = fmt.Sprintf("user:%d:sticker_set", userID)
	if err := s.rdb.Del(s.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to flush sticker set: %w", err)
	}

	key = fmt.Sprintf("user:%d:title", userID)
	if err := s.rdb.Del(s.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to flush title: %w", err)
	}

	key = fmt.Sprintf("user:%d:name", userID)
	if err := s.rdb.Del(s.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to flush name: %w", err)
	}

	if err := s.switchState(userID, selectSet); err != nil {
		return fmt.Errorf("failed to revert to start: %w", err)
	}

	return nil
}

func (s *Service) getAll(userID int64) (title, newName, oldName string, deletionList []string, err error) {
	title, err = s.getTitle(userID)
	if err != nil {
		return "", "", "", []string{}, fmt.Errorf("Error occurred while fetching new sticker title: %w", err)
	}
	newName, err = s.getName(userID)
	if err != nil {
		return "", "", "", []string{}, fmt.Errorf("Error occurred while fetching new sticker name: %w", err)
	}
	deletionList, err = s.getDeletionList(userID)
	if err != nil {
		return "", "", "", []string{}, fmt.Errorf("Error occurred while fetching deletion list: %w", err)
	}
	oldName, err = s.getStickerSet(userID)
	if err != nil {
		return "", "", "", []string{}, fmt.Errorf("Error occurred while fetching new sticker title: %w", err)
	}
	return
}
