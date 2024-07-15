package telebot

import (
	"log"
	"sync"
	"time"

	"gopkg.in/telebot.v3"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/react"
)

type IBot interface {
	Info() any
	SendMessage(*tele.Message)
}

type Bot struct {
	*tele.Bot
	userCommands map[int64]*UserCommand
	mt           *sync.RWMutex
}

func (b Bot) GetUserCommand(userid int64) *UserCommand {
	return b.userCommands[userid]
}

type UserCommand struct {
	command        string
	userId         int64
	state          int // 1: waitting | 2: working
	created        int64
	args           []string
	numRunningTask int
}

func Start(token string) *Bot {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	newBot := &Bot{
		Bot:          b,
		mt:           &sync.RWMutex{},
		userCommands: make(map[int64]*UserCommand)}

	newBot.autoClearCommand()

	return newBot
}

func (b *Bot) PrivateHandlers() *Bot {
	b.Handle(telebot.OnText, func(c tele.Context) error {
		fire := react.Sunglasses.Emoji
		return c.Send(fire + fire + fire)
	})

	// // Xử lý tin nhắn gửi ảnh từ người dùng
	// b.Handle(telebot.OnPhoto, func(c tele.Context) error {
	// 	// cmd := b.userCommands[c.Sender().ID]
	// 	return c.Send(":))")
	// })
	return b
}

func (b *Bot) isCanProccess(userId int64) (bool, string) {
	b.mt.Lock()
	defer b.mt.Unlock()

	cmd := b.userCommands[userId]
	if cmd == nil {
		return true, ""
	}
	if cmd.state == 1 {
		return false, "Công việc đang đợi mời bạn tiếp tục. Hoặc nhấn cancel để huỷ hành động!"
	}
	if cmd.state == 2 {
		return false, "Công việc đang thực thi. Chờ chút nhé!"
	}
	return true, ""
}

func (b *Bot) autoClearCommand() {
	t := time.NewTicker(time.Minute)
	go func() {
		for {
			<-t.C
			b.mt.Lock()
			for userId, cmd := range b.userCommands {
				if cmd.created+3*60 < time.Now().Unix() {
					// xoa task qua han
					delete(b.userCommands, userId)
				}
			}
			b.mt.Unlock()
		}
	}()
}
