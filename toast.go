package main

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type toast struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Id    string `json:"id"`
}

func (ctf *ctf) getToasts(c *fiber.Ctx) ([]toast, error) {
	sess, err := ctf.Sessions.Get(c)
	if err != nil {
		return []toast{}, err
	}
	toastInterface := sess.Get("toasts")
	if toastInterface == nil {
		return []toast{}, nil
	}
	var toasts []toast
	if json.Unmarshal(toastInterface.([]byte), &toasts) != nil {
		return toasts, err
	}

	return toasts, nil
}

func (ctf *ctf) saveToasts(c *fiber.Ctx, toasts []toast) error {
	b, err := json.Marshal(toasts)
	if err != nil {
		return err
	}
	err = ctf.setSessionKey(c, "toasts", b)
	return err
}

func (ctf *ctf) addToast(c *fiber.Ctx, title, text string) {
	t := toast{
		Title: title,
		Text:  text,
		Id:    uuid.New().String(),
	}

	userToasts, err := ctf.getToasts(c)
	if err != nil {
		userToasts = []toast{t}
	}
	userToasts = append(userToasts, t)

	_ = ctf.saveToasts(c, userToasts)
}

func (ctf *ctf) deleteToast(c *fiber.Ctx, id string) {
	userToasts, err := ctf.getToasts(c)
	var newToasts []toast
	if err != nil {
		userToasts = []toast{}
	}

	for _, toast := range userToasts {
		if toast.Id != id {
			newToasts = append(newToasts, toast)
		}
	}
	_ = ctf.saveToasts(c, newToasts)
}
