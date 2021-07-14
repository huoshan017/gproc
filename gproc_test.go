package gproc

import (
	//"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// 商店物品
type ShopItem struct {
	instId int32
	id     int32
	count  int32
	price  int32
}

// 物品列表请求
type GetItemListReq struct {
}

// 物品列表返回
type GetItemListResp struct {
	itemList []*ShopItem
}

// 购买物品请求
type BuyItemReq struct {
	instId     int32
	count      int32
	totalMoney int32
}

// 购买物品返回
type BuyItemResp struct {
	Err       int32
	instId    int32
	id        int32
	count     int32
	costMoney int32
}

// 商店服务
type ShopService struct {
	LocalService

	itemList []*ShopItem
}

// 创建商店服务
func NewShopService() *ShopService {
	service := &ShopService{}
	return service
}

// 初始化
func (s *ShopService) Init() {
	s.LocalService.Init()
	s.itemList = make([]*ShopItem, 0)
	s.RegisterHandle("getItemListReq", s.getItemList)
	s.RegisterHandle("buyItemReq", s.buyItem)
	s.SetTickHandle(s.tick)
}

// 添加物品
func (s *ShopService) AddItem(item *ShopItem) {
	s.itemList = append(s.itemList, item)
}

// 删除物品
func (s *ShopService) RemoveItem(instId int32, count int32) bool {
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
func (s *ShopService) getItemList(sender ISender, args interface{}) {
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
	sender.Send("getItemListResp", resp)
}

// 处理购买物品
func (s *ShopService) buyItem(sender ISender, args interface{}) {
	resp := &BuyItemResp{}
	req := args.(*BuyItemReq)
	if req.instId <= 0 || req.count <= 0 {
		resp.Err = -99
		resp.instId = req.instId
		resp.count = req.count
		sender.Send("buyItemResp", resp)
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
	sender.Send("buyItemResp", resp)
}

// 定时器处理
func (s *ShopService) tick(tick int32) {
	// 价格随时间变化
	for i := 0; i < len(s.itemList); i++ {
		item := s.itemList[i]
		if item != nil {
			item.price += (rand.Int31n(int32(item.price)) - item.price/2)
		}
	}
}

type Item struct {
	instId int32
	id     int32
	count  int32
}

// 玩家
type Player struct {
	ResponseHandler // 一般是通过继承来使用ResponseHandler
	shopRequester   IRequester
	money           int32
	itemList        []*Item
}

// 创建玩家
func NewPlayer(money int32) *Player {
	p := &Player{money: money, itemList: make([]*Item, 0)}
	p.Init(true)
	return p
}

// 创建商店请求者
func (p *Player) CreateShopRequester(shop *ShopService) {
	p.shopRequester = NewRequester(&p.ResponseHandler, &shop.RequestHandler)
}

// 注册回调
func (p *Player) RegisterResponseHandlers() {
	p.shopRequester.RegisterCallback("getItemListReq", "getItemListResp", func(param interface{}) {
		//resp := param.(*GetItemListResp)
		//log.Printf("get item list: %v", resp.itemList)
	})
	p.shopRequester.RegisterCallback("buyItemReq", "buyItemResp", func(param interface{}) {
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
func (p *Player) GetItemList() {
	p.shopRequester.Request("getItemListReq", &GetItemListReq{})
}

// 购买物品
func (p *Player) BuyItem(instId int32, count int32) {
	p.shopRequester.Request("buyItemReq", &BuyItemReq{instId: instId, count: count, totalMoney: p.money})
}

// 循环处理
func (p *Player) Run(wg *sync.WaitGroup) {
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
func TestShopService(t *testing.T) {
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
	shop := NewShopService()
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
		p := NewPlayer(100000)
		p.CreateShopRequester(shop)
		p.RegisterResponseHandlers()
		go p.Run(&wg)
		defer p.Close()
	}

	wg.Wait()
}
