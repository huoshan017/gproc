package gproc

import (
	//"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// 商店处理器
type ShopHandler struct {
	RequestHandler

	itemList []*ShopItem
}

// 创建商店服务
func NewShopHandler() *ShopHandler {
	h := &ShopHandler{}
	return h
}

// 初始化
func (s *ShopHandler) Init() {
	s.InitDefault()
	s.itemList = make([]*ShopItem, 0)
	s.RegisterHandle("getItemList", s.getItemList)
	s.RegisterHandle("buyItem", s.buyItem)
}

// 添加物品
func (s *ShopHandler) AddItem(item *ShopItem) {
	s.itemList = append(s.itemList, item)
}

// 删除物品
func (s *ShopHandler) RemoveItem(instId int32, count int32) bool {
	if count <= 0 {
		return false
	}

	for i := 0; i < len(s.itemList); i++ {
		item := s.itemList[i]
		if item.instId == instId && item.count >= count {
			item.count -= count
			return true
		}
	}

	return false
}

// 处理获取物品列表
func (s *ShopHandler) getItemList(sender ISender, args interface{}) {
	resp := &GetItemListResp{}
	for i := 0; i < len(s.itemList); i++ {
		item := s.itemList[i]
		resp.itemList = append(resp.itemList, &ShopItem{
			instId: item.instId,
			id:     item.id,
			count:  item.count,
			price:  item.price,
		})
	}
	sender.Send("getItemList", resp)
}

// 处理购买物品
func (s *ShopHandler) buyItem(sender ISender, args interface{}) {
	resp := &BuyItemResp{}
	req := args.(*BuyItemReq)
	if req.instId <= 0 || req.count <= 0 {
		resp.Err = -99
		resp.instId = req.instId
		resp.count = req.count
		sender.Send("buyItem", resp)
		return
	}

	var foundItem *ShopItem
	for i := 0; i < len(s.itemList); i++ {
		item := s.itemList[i]
		if item.instId == req.instId {
			foundItem = item
			break
		}
	}
	if foundItem == nil {
		resp.Err = -1 // 没有该物品
		//log.Printf("not found item %v", req.instId)
	} else {
		if foundItem.price*req.count > req.totalMoney {
			resp.Err = -2
			//log.Printf("money %v not enough to buy item %v count %v, need money %v", req.totalMoney, foundItem.instId, req.count, foundItem.price*req.count)
		} else if foundItem.count < req.count {
			resp.Err = -3
			//log.Printf("item %v count %v not enough to buy, need %v", foundItem.instId, foundItem.count, req.count)
		} else {
			foundItem.count -= req.count
			resp.instId = foundItem.instId
			resp.id = foundItem.id
			resp.count = foundItem.count
			resp.costMoney = foundItem.price * req.count
			//log.Printf("bought item %v count %v, cost money %v", foundItem.instId, req.count, resp.costMoney)
		}
	}
	sender.Send("buyItem", resp)
}

/*
// 定时器处理
func (s *ShopHandler) tick(tick int32) {
	// 价格随时间变化
	for i := 0; i < len(s.itemList); i++ {
		item := s.itemList[i]
		if item != nil {
			item.price += (rand.Int31n(int32(item.price)) - item.price/2)
		}
	}
}
*/

// 玩家
type PlayerRequester struct {
	ResponseHandler // 一般是通过继承来使用ResponseHandler
	shopRequester   IRequester
	money           int32
	itemList        []*Item
}

// 创建玩家
func NewPlayerRequester(money int32) *PlayerRequester {
	p := &PlayerRequester{money: money, itemList: make([]*Item, 0)}
	p.InitDefault()
	return p
}

// 创建商店请求者
func (p *PlayerRequester) CreateShopRequester(shop *ShopHandler) {
	p.shopRequester = NewRequester(p, shop)
}

// 注册回调
func (p *PlayerRequester) RegisterResponseHandlers() {
	p.shopRequester.RegisterCallback("getItemList", func(param interface{}) {
		//resp := param.(*GetItemListResp)
		//log.Printf("get item list: %v", resp.itemList)
	})
	p.shopRequester.RegisterCallback("buyItem", func(param interface{}) {
		resp := param.(*BuyItemResp)
		if resp.Err < 0 {
			//log.Printf("buy item %v failed, err %v, count %v", resp.instId, resp.Err, resp.count)
			return
		}
		p.money -= resp.costMoney
		var item *Item
		for i := 0; i < len(p.itemList); i++ {
			tmpItem := p.itemList[i]
			if tmpItem.instId == resp.instId {
				item = tmpItem
				break
			}
		}
		if item != nil {
			item.count += resp.count
		} else {
			p.itemList = append(p.itemList, &Item{instId: resp.instId, id: resp.id, count: resp.count})
		}
		//log.Printf("buy item %v count %v success, cost %v money", resp.instId, resp.count, resp.costMoney)
	})
}

// 获得物品列表
func (p *PlayerRequester) GetItemList() {
	p.shopRequester.Request("getItemList", &GetItemListReq{})
}

// 购买物品
func (p *PlayerRequester) BuyItem(instId int32, count int32) {
	p.shopRequester.Request("buyItem", &BuyItemReq{instId: instId, count: count, totalMoney: p.money})
}

// 循环处理
func (p *PlayerRequester) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 1000; i++ {
		r := rand.Int31n(2)
		if r == 0 {
			p.GetItemList()
		} else {
			itemIdList := []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}
			idx := rand.Int31n(int32(len(itemIdList)))
			p.BuyItem(itemIdList[idx], rand.Int31n(10)+1)
		}
		p.Update()
		time.Sleep(time.Microsecond * 10)
	}
}

// 商店服务测试
func TestShopHandler(t *testing.T) {
	itemList := []*ShopItem{
		{1, 1, 100, 10},
		{2, 2, 200, 20},
		{3, 3, 300, 30},
		{4, 4, 400, 40},
		{5, 5, 500, 50},
		{6, 6, 600, 60},
		{7, 7, 700, 70},
		{8, 8, 800, 80},
		{9, 9, 900, 90},
		{10, 10, 1000, 100},
		{11, 11, 1100000, 11},
		{12, 12, 1200000, 12},
		{13, 13, 1300000, 13},
		{14, 14, 1400000, 14},
		{15, 15, 1500000, 15},
		{16, 16, 1600000, 16},
		{17, 17, 1700000, 17},
		{18, 18, 1800000, 18},
		{19, 19, 1900000, 19},
		{20, 20, 2000000, 20},
	}
	shop := NewShopHandler()
	shop.Init()
	for i := 0; i < len(itemList); i++ {
		shop.AddItem(itemList[i])
	}

	playerCount := 100000
	var wg sync.WaitGroup
	wg.Add(playerCount)

	go shop.Run()
	defer shop.Close()

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < playerCount; i++ {
		p := NewPlayerRequester(100000)
		p.CreateShopRequester(shop)
		p.RegisterResponseHandlers()
		go p.Run(&wg)
		defer p.Close()
	}

	wg.Wait()
}
