package gproc

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

var test_string = `e32qr213343wg43wg4w3g4wdbsrbsrw3e3$#@#%$%^$)_8998t349t43gw34btrsbtrsh4a122176i87v是然而色不同认识你是32它35654`

var playerIds = []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30}

func getAChatMessage() string {
	data := []byte(test_string)
	s := rand.Intn(len(data))
	e := rand.Intn(len(data))
	if e == s {
		e = s + rand.Intn(10)
	} else if e < s {
		e, s = s, e
	}
	return string(data[s:e])
}

type msgChat struct {
	message string
}

type msgChatAck struct {
	message string
}

// 聊天玩家
type ChatPlayer struct {
	ResponseHandler
	id            int32
	chatRequester IRequester
}

func NewChatPlayer(id int32) *ChatPlayer {
	p := &ChatPlayer{
		id: id,
	}
	p.InitDefault()
	return p
}

func (p *ChatPlayer) InitRequester(chatService *ChatService) {
	p.chatRequester = p.CreateRequester(chatService, p.id)
}

func (p *ChatPlayer) RegisterHandles() {
	p.chatRequester.RegisterForward("msgChat", p.handleChat)
	p.chatRequester.RegisterForward("msgChatAck", p.handleChatAck)
}

func (p *ChatPlayer) handleChat(fromKey interface{}, args interface{}) {
	chatMsg := args.(*msgChat)
	fmt.Println("player ", p.id, " recv message ", chatMsg.message, " from player ", fromKey)
	ack := &msgChatAck{message: chatMsg.message}
	err := p.chatRequester.RequestForward(fromKey, "msgChatAck", ack)
	if err != nil {
		fmt.Println("player ", p.id, " request forward player ", fromKey, " err: ", err)
		return
	}
	fmt.Println("player ", p.id, " ack message ", ack.message, " to sender ", fromKey)
}

func (p *ChatPlayer) handleChatAck(fromKey interface{}, args interface{}) {
	chatAckMsg := args.(*msgChatAck)
	fmt.Println("player ", p.id, " recv ack message ", chatAckMsg.message, " from player ", fromKey)
}

func (p *ChatPlayer) ChatTo(pid int32, message string) {
	err := p.chatRequester.RequestForward(pid, "msgChat", &msgChat{message: message})
	if err != nil {
		fmt.Println("player ", p.id, " chat to ", pid, " err: ", err)
	} else {
		fmt.Println("player ", p.id, " chat to ", pid, " message: ", message)
	}
}

func (p *ChatPlayer) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 1; i++ {
		pid := p.randomPlayerId()
		p.ChatTo(pid, getAChatMessage())
		time.Sleep(time.Millisecond)
	}

	for {
		p.Update()
		time.Sleep(time.Millisecond)
	}
}

func (p *ChatPlayer) randomPlayerId() int32 {
	idx := rand.Int31n(int32(len(playerIds)))
	if playerIds[idx] == p.id {
		idx += 1
		if int(idx) >= len(playerIds) {
			idx = 0
		}
	}
	return playerIds[idx]
}

// 聊天服务
type ChatService struct {
	RequestHandler
}

func NewChatService() *ChatService {
	return &ChatService{}
}

func (s *ChatService) Init() {
	s.InitDefault()
	s.SetTickHandle(s.tick, time.Millisecond)
}

func (s *ChatService) tick(tick time.Duration) {

}

func TestChatForward(t *testing.T) {
	chatService := NewChatService()
	chatService.Init()
	go chatService.Run()
	defer chatService.Close()

	playerCount := len(playerIds)
	var wg sync.WaitGroup
	wg.Add(playerCount)

	rand.Seed(time.Now().Unix())
	for i := 0; i < playerCount; i++ {
		p := NewChatPlayer(playerIds[i])
		defer p.Close()

		p.InitRequester(chatService)
		p.RegisterHandles()
		go p.Run(&wg)
	}

	wg.Wait()

	time.Sleep(time.Second * 5)
}
