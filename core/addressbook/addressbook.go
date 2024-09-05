package addressbook

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/dismutex"
	"golang.org/x/exp/rand"
)

type IAddressBook interface {
	Register(context.Context, string, string) error // 将 actor 注册到 address book；
	Unregister(context.Context, string) error

	GetByID(context.Context, string) (AddressInfo, error)
	GetByType(context.Context, string) ([]AddressInfo, error)
}

type AddressInfo struct {
	ActorId string `json:"actor_id"`
	ActorTy string `json:"actor_ty"`
	Node    string `json:"node"`
	Service string `json:"service"`
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
}

type AddressBook struct {
	NodInfo AddressInfo

	IDMap map[string]bool

	sync.RWMutex
}

func New(info AddressInfo) *AddressBook {
	return &AddressBook{
		IDMap:   make(map[string]bool),
		NodInfo: info,
	}
}

func (ab *AddressBook) Register(ctx context.Context, ty, id string) error {

	// check id
	if id == "" || ty == "" {
		return fmt.Errorf("actor id or type is empty")
	}

	ab.RLock()
	if _, ok := ab.IDMap[id]; ok {
		ab.RUnlock()
		return fmt.Errorf("actor id %v already registered", id)
	}
	ab.RUnlock()

	mu := &dismutex.Mutex{Token: id}
	err := mu.Lock(ctx, "[addressbook.register]")
	if err != nil {
		return fmt.Errorf("addressbook.register get distributed mutex err %v", err.Error())
	}
	defer mu.Unlock(ctx)

	// 将地址信息序列化为 JSON
	addrJSON, _ := json.Marshal(AddressInfo{
		ActorId: id,
		ActorTy: ty,
		Ip:      ab.NodInfo.Ip,
		Port:    ab.NodInfo.Port},
	)
	// 使用管道来执行多个 Redis 操作
	pipe := redis.Pipeline()
	pipe.HSet(ctx, def.RedisAddressbookIDField, id, addrJSON)
	pipe.SAdd(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", ty), addrJSON)
	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("redis pipeline exec err %v", err.Error())
	}

	ab.Lock()
	ab.IDMap[id] = true
	ab.Unlock()

	return nil
}

func (ab *AddressBook) Unregister(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("actor id or type is empty")
	}

	mu := &dismutex.Mutex{Token: id}
	err := mu.Lock(ctx, "[addressbook.unregister]")
	if err != nil {
		return fmt.Errorf("addressbook.unregister get distributed mutex err %v", err.Error())
	}
	defer mu.Unlock(ctx)

	// 首先获取地址信息
	addrJSON, err := redis.HGet(ctx, def.RedisAddressbookIDField, id).Result()
	if err != nil {
		return fmt.Errorf("address not found for id: %s", id)
	}

	info := &AddressInfo{}
	err = json.Unmarshal([]byte(addrJSON), info)
	if err != nil {
		return fmt.Errorf("addressbook.unregister json unmarshal err %v", err.Error())
	}

	// 使用管道来执行多个 Redis 操作
	pipe := redis.Pipeline()
	pipe.HDel(ctx, def.RedisAddressbookIDField, id)
	pipe.SRem(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", info.ActorTy), addrJSON)
	_, err = pipe.Exec(ctx)
	if err == nil {
		ab.Lock()
		delete(ab.IDMap, id) // try delete
		ab.Unlock()
	}

	return err
}

// GetByID 通过 ID 获取 actor 地址
func (ab *AddressBook) GetByID(ctx context.Context, id string) (AddressInfo, error) {

	if id == "" {
		return AddressInfo{}, fmt.Errorf("actor id or type is empty")
	}

	ab.RLock()
	if _, ok := ab.IDMap[id]; ok {
		ab.RUnlock()
		return ab.NodInfo, nil // 直接返回本节点信息
	}
	ab.RUnlock()

	addrJSON, err := redis.HGet(ctx, def.RedisAddressbookIDField, id).Result()
	if err != nil {
		return AddressInfo{}, fmt.Errorf("address not found for id: %s", id)
	}

	var addr AddressInfo
	err = json.Unmarshal([]byte(addrJSON), &addr)
	if err != nil {
		return AddressInfo{}, fmt.Errorf("failed to unmarshal address: %v", err)
	}

	return addr, nil
}

// GetByType 通过类型获取所有 actor 地址
func (ab *AddressBook) GetByType(ctx context.Context, actorType string) ([]AddressInfo, error) {
	addrJSONs, err := redis.SMembers(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", actorType)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses for type: %s", actorType)
	}

	addresses := make([]AddressInfo, 0, len(addrJSONs))
	for _, addrJSON := range addrJSONs {
		var addr AddressInfo
		err = json.Unmarshal([]byte(addrJSON), &addr)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal address: %v", err)
		}
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

func (ab *AddressBook) GetWildcardActor(lst []AddressInfo) (AddressInfo, error) {
	if len(lst) == 0 {
		return AddressInfo{}, fmt.Errorf("GetWildcardActor lst is empty")
	}

	localActors := make([]AddressInfo, 0)
	for _, actor := range lst {
		if actor.Ip == ab.NodInfo.Ip && actor.Port == ab.NodInfo.Port {
			localActors = append(localActors, actor)
		}
	}

	if len(localActors) > 0 {
		return localActors[rand.Intn(len(localActors))], nil
	}

	return lst[rand.Intn(len(lst))], nil
}
