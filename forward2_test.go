package gproc

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

/********************** 消息 *********************/
// 更新好友信息
type msgUpdateFriendInfo struct {
	selfId int32
	level  int32
}

// 请求推荐列表
type msgRecommendationFriendListReq struct {
	selfId int32
	level  int32
}

// 返回推荐列表
type msgRecommendationFriendListResp struct {
	friendList []int32
}

// 请求添加好友
type msgFriendAddReq struct {
}

// 添加好友回复
type msgFriendAddAck struct {
}

// 请求删除好友
type msgFriendRemoveReq struct {
}

// 删除好友回复
type msgFriendRemoveAck struct {
}

/************************************************/

const (
	PlayerFriendStateGetRecommendationList = 1
	PlayerFriendStateAddFriendReq          = 2
)

const (
	MsgIdFriendRecommendationList = 1
	MsgIdFriendAdd                = 2
	MsgIdFriendRemove             = 3
	MsgIdFriendAddAck             = 4
	MsgIdFriendRemoveAck          = 5
	MsgIdUpdateFriendInfo         = 6
)

// 创建商店请求者
func (p *Player) CreateFriendRequester(fs *FriendService) {
	p.friendRequester = NewRequester(p, fs, p.id)
}

// 注册回调
func (p *Player) RegisterFriendHandlers() {
	p.friendRequester.RegisterCallback(MsgIdFriendRecommendationList, p.handleFriendRecommendationList)
	p.friendRequester.RegisterForward(MsgIdFriendAdd, p.handleFriendAdd)
	p.friendRequester.RegisterForward(MsgIdFriendRemove, p.handleFriendRemove)
	p.friendRequester.RegisterForward(MsgIdFriendAddAck, p.handleFriendAddAck)
	p.friendRequester.RegisterForward(MsgIdFriendRemoveAck, p.handleFriendRemoveAck)
}

// 添加好友
func (p *Player) addFriend(pid int32) bool {
	found := false
	for i := 0; i < len(p.friendList); i++ {
		if pid == p.friendList[i] {
			found = true
			break
		}
	}
	if !found {
		p.friendList = append(p.friendList, pid)
	}
	return !found
}

// 删除好友
func (p *Player) removeFriend(pid int32) bool {
	idx := 0
	for ; idx < len(p.friendList); idx++ {
		if pid == p.friendList[idx] {
			break
		}
	}
	if idx == len(p.friendList) {
		return false
	}
	p.friendList = append(p.friendList[:idx], p.friendList[idx+1:]...)
	return true
}

// 获取推荐列表
func (p *Player) handleFriendRecommendationList(args interface{}) {
	msg := args.(*msgRecommendationFriendListResp)
	p.step += 1
	fmt.Println("玩家 ", p.id, " 推荐列表 ", msg.friendList)
}

// 处理好友添加
func (p *Player) handleFriendAdd(fromKey interface{}, args interface{}) {
	pid := fromKey.(int32)
	if p.id == pid {
		fmt.Println("Player ", p.id, " cant add self to friend")
		return
	}
	if !p.addFriend(pid) {
		fmt.Println("Player ", p.id, " already added friend ", pid)
		return
	}
	fmt.Println("Player ", p.id, " added new friend ", pid)

	p.friendRequester.RequestForward(fromKey, MsgIdFriendAddAck, &msgFriendAddAck{})

	fmt.Println("Player ", p.id, " ack add friend ", pid)
}

// 处理好友添加回复
func (p *Player) handleFriendAddAck(fromKey interface{}, args interface{}) {
	pid := fromKey.(int32)
	if !p.addFriend(pid) {
		fmt.Println("Player ", p.id, " already added friend ", pid, " , add ack failed")
		return
	}
	p.step += 1
	fmt.Println("Player ", p.id, " ack added new friend ", pid)
}

// 处理好友删除
func (p *Player) handleFriendRemove(fromKey interface{}, args interface{}) {
	pid := fromKey.(int32)
	if p.id == pid {
		fmt.Println("Player ", p.id, " cant remove self as friend")
		return
	}
	if !p.removeFriend(pid) {
		fmt.Println("Player ", p.id, " not found friend ", pid)
		return
	}
	fmt.Println("Player ", p.id, " removed friend ", pid)

	p.friendRequester.RequestForward(fromKey, MsgIdFriendRemoveAck, &msgFriendRemoveAck{})

	fmt.Println("Player ", p.id, " ack remove friend ", pid)
}

// 处理好友删除回复
func (p *Player) handleFriendRemoveAck(fromKey interface{}, args interface{}) {
	pid := fromKey.(int32)
	if !p.removeFriend(pid) {
		fmt.Println("Player ", p.id, " not found friend ", pid, " remove ack failed")
		return
	}
	p.step += 1
	fmt.Println("Player ", p.id, " ack removed friend ", pid)
}

// 更新等级
func (p *Player) updateFriendInfo() {
	p.friendRequester.Request(MsgIdUpdateFriendInfo, &msgUpdateFriendInfo{selfId: p.id, level: p.level})
}

// 请求推荐列表
func (p *Player) getFriendRecommendationList() {
	p.friendRequester.Request(MsgIdFriendRecommendationList, &msgRecommendationFriendListReq{selfId: p.id, level: p.level})
}

// 添加好友请求
func (p *Player) addFriendReq(pid int32) {
	p.friendRequester.RequestForward(pid, MsgIdFriendAdd, &msgFriendAddReq{})
}

// 删除好友请求
func (p *Player) removeFriendReq(pid int32) {
	p.friendRequester.RequestForward(pid, MsgIdFriendRemove, &msgFriendRemoveReq{})
}

// 创建Player
func CreatePlayer(id int32, level int32, fs *FriendService) *Player {
	p := NewPlayer(id, &playerConfig{level: level})
	p.InitDefault()
	p.CreateFriendRequester(fs)
	p.RegisterFriendHandlers()
	return p
}

/******************************* 好友服务 *****************************/
type FriendService struct {
	*RequestHandler
	PlayerIds []int32
}

func NewFriendService() *FriendService {
	return &FriendService{
		RequestHandler: NewDefaultRequestHandler(),
	}
}

func (s *FriendService) Init() {
	s.RegisterHandles()
}

func (s *FriendService) RegisterHandles() {
	s.RegisterHandle(MsgIdUpdateFriendInfo, func(_ ISender, args interface{}) {
		msg := args.(*msgUpdateFriendInfo)
		s.PlayerIds = append(s.PlayerIds, msg.selfId)
		fmt.Println("Player ", msg.selfId, " update friend info, now player list ", s.PlayerIds)
	})
	s.RegisterHandle(MsgIdFriendRecommendationList, func(sender ISender, args interface{}) {
		req := args.(*msgRecommendationFriendListReq)
		var pidList []int32
		for _, pid := range s.PlayerIds {
			if int32(pid) == req.selfId {
				continue
			}
			pidList = append(pidList, int32(pid))
		}
		sender.Send(MsgIdFriendRecommendationList, &msgRecommendationFriendListResp{friendList: pidList})
		fmt.Println("Player ", req.selfId, " get friend recommendation list ", pidList)
	})
}

func CreateFriendService() *FriendService {
	s := NewFriendService()
	s.Init()
	s.RegisterHandles()
	return s
}

/********************************************************************/

func randPlayerId(selfId int32, playerIds []int32) int32 {
	if len(playerIds) == 0 {
		fmt.Println("Player ", selfId, " no friends")
		return -1
	}
	r := rand.Int31n(int32(len(playerIds)))
	if playerIds[r] == selfId {
		r += 1
		if int(r) == len(playerIds) {
			r = 0
		}
	}
	return playerIds[r]
}

func TestFriendService(t *testing.T) {
	fs := CreateFriendService()
	defer fs.Close()
	go fs.Run()

	playerCount := int32(4)
	players := make([]*Player, playerCount)
	var wg sync.WaitGroup
	wg.Add(int(playerCount))
	for id := int32(1); id <= playerCount; id++ {
		p := CreatePlayer(id, 1, fs)
		defer p.Close()
		p.updateFriendInfo()
		players[id-1] = p
	}

	rand.Seed(time.Now().UnixNano())

	count := int32(1)
	for id := int32(1); id <= playerCount; id++ {
		p := players[id-1]
		go func(p *Player) {
			state := 0
			for state != 4 {
				switch p.step {
				case 0:
					// 获得推荐列表
					if state == 0 {
						p.getFriendRecommendationList()
						state = 1
					}
				case 1:
					if state == 1 {
						pid := randPlayerId(p.id, fs.PlayerIds)
						p.addFriendReq(pid)
						state = 2
					}
				case 2:
					if state == 2 {
						pid := randPlayerId(p.id, p.friendList)
						p.removeFriendReq(pid)
						state = 3
					}
				case 3:
					if state == 3 {
						wg.Done()
						state = 4
						atomic.AddInt32(&count, 1)
						fmt.Println("wgDone ", count)
					}
				}
				p.Update()
				time.Sleep(time.Millisecond)
			}
		}(p)
	}

	wg.Wait()
}
