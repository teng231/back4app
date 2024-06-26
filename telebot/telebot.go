package telebot

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gopkg.in/telebot.v3"
	tele "gopkg.in/telebot.v3"
)

var (
	selector      = &tele.ReplyMarkup{}
	maxConcurrent = 4
)

type TBot struct {
	bot          *tele.Bot
	userCommands map[int64]*UserCommand
	mt           *sync.RWMutex
}

type UserCommand struct {
	command        string
	userId         int64
	state          int // 1: waitting | 2: working
	created        int64
	description    string
	args           []string
	numRunningTask int
}

type ITBot interface {
	Info() any
	SendMessage(*tele.Message)
}

func Start(token string) *TBot {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	if err != nil {
		log.Fatal(err)
	}
	return &TBot{bot: b,
		mt:           &sync.RWMutex{},
		userCommands: make(map[int64]*UserCommand)}
}

func (b *TBot) isCanProccess(userId int64) (bool, string) {
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

/**
command list:
ping - Kiểm tra bot đang running hay không?
txt2img - Nhập vào 1 text và bot biến thành 1 ảnh.
img2vid_g - Nhập vào 1 ảnh và bot biến nó thành video. Dùng server genM
img2vid_m - Nhập vào 1 ảnh và bot biến nó thành video. Dùng server modelsL
vid2vids - Nhập vào 1 số videos, bot nối chúng thành 1 video dài hơn.
vidaud - Viết 1 đoạn caption của 1 video và bot sẽ đọc nó tạo 1 video với giọng đọc tương ứng.
*/

func (b *TBot) Run() {
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
	b.bot.Handle("/ping", func(c tele.Context) error {
		return c.Send("I'm alive!")
	})
	b.bot.Handle("/cancel", func(c tele.Context) error {
		b.mt.Lock()
		defer b.mt.Unlock()

		cmd := b.userCommands[c.Sender().ID]
		if cmd != nil && cmd.state == 2 {
			return c.Send("Lệnh đang thực thi không thể huỷ. Bạn chỉ huỷ được các lệnh trong trạng thái chờ.")
		}
		delete(b.userCommands, c.Sender().ID)
		return c.Send("Đã huỷ yêu cầu.")
	})
	b.bot.Handle("/img2vid_g", func(c tele.Context) error {
		next, msg := b.isCanProccess(c.Sender().ID)
		if !next {
			return c.Send(msg)
		}
		b.mt.Lock()
		defer b.mt.Unlock()
		description := "Mời nhập vào 1 đoạn lệnh(prompt) và dấu nhắc tiêu cực. Ngăn cách nhau bằng | nhé."
		b.userCommands[c.Sender().ID] = &UserCommand{
			userId:      c.Sender().ID,
			command:     "/txt2img_g",
			created:     time.Now().Unix(),
			state:       1,
			description: description,
			args:        c.Args(),
		}
		return c.Send("Nhận lệnh: " + description)
	})
	b.bot.Handle("/img2vid_m", func(c tele.Context) error {
		next, msg := b.isCanProccess(c.Sender().ID)
		if !next {
			return c.Send(msg)
		}
		b.mt.Lock()
		defer b.mt.Unlock()
		description := "Mời nhập vào 1 ảnh"
		b.userCommands[c.Sender().ID] = &UserCommand{
			userId:  c.Sender().ID,
			command: "/img2vid_m",
			created: time.Now().Unix(),
			// state:       1,
			description: description,
			args:        c.Args(),
		}
		return c.Send("Nhận lệnh: " + description)
	})
	// Xử lý tin nhắn văn bản từ người dùng
	b.bot.Handle(telebot.OnText, func(c tele.Context) error {
		cmd := b.userCommands[c.Sender().ID]
		if cmd != nil && cmd.command == "/txt2img" {
			if cmd.state == 1 {
				cmd.state = 2
				// handleTxt2Img()
				// run
			}
			if cmd.state == 2 {
				return c.Send("Từ từ anh !!!")
			}
		}
		return c.Send(fmt.Sprintf("Bạn đã nói: %s", c.Text()))
	})

	// Xử lý tin nhắn gửi ảnh từ người dùng
	b.bot.Handle(telebot.OnPhoto, func(c tele.Context) error {
		cmd := b.userCommands[c.Sender().ID]
		return c.Send(":))")
	})

	b.bot.Start()
}
