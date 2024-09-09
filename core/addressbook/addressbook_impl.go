package addressbook

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	trdredis "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/dismutex"
	"github.com/redis/go-redis/v9"
)

type AddressBook struct {
	NodInfo core.AddressInfo

	IDMap map[string]bool

	sync.RWMutex
}

func New(info core.AddressInfo) *AddressBook {
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
	addrJSON, _ := json.Marshal(core.AddressInfo{
		ActorId: id,
		ActorTy: ty,
		Ip:      ab.NodInfo.Ip,
		Port:    ab.NodInfo.Port},
	)
	// 使用管道来执行多个 Redis 操作
	pipe := trdredis.Pipeline()
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
	addrJSON, err := trdredis.HGet(ctx, def.RedisAddressbookIDField, id).Result()
	if err != nil {
		return fmt.Errorf("address not found for id: %s", id)
	}

	info := &core.AddressInfo{}
	err = json.Unmarshal([]byte(addrJSON), info)
	if err != nil {
		return fmt.Errorf("addressbook.unregister json unmarshal err %v", err.Error())
	}

	// 使用管道来执行多个 Redis 操作
	pipe := trdredis.Pipeline()
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
func (ab *AddressBook) GetByID(ctx context.Context, id string) (core.AddressInfo, error) {

	if id == "" {
		return core.AddressInfo{}, fmt.Errorf("actor id or type is empty")
	}

	ab.RLock()
	if _, ok := ab.IDMap[id]; ok {
		ab.RUnlock()
		return ab.NodInfo, nil // 直接返回本节点信息
	}
	ab.RUnlock()

	addrJSON, err := trdredis.HGet(ctx, def.RedisAddressbookIDField, id).Result()
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("address not found for id: %s", id)
	}

	var addr core.AddressInfo
	err = json.Unmarshal([]byte(addrJSON), &addr)
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("failed to unmarshal address: %v", err)
	}

	return addr, nil
}

// GetByType 通过类型获取所有 actor 地址
func (ab *AddressBook) GetByType(ctx context.Context, actorType string) ([]core.AddressInfo, error) {
	addrJSONs, err := trdredis.SMembers(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", actorType)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses for type: %s", actorType)
	}

	addresses := make([]core.AddressInfo, 0, len(addrJSONs))
	for _, addrJSON := range addrJSONs {
		var addr core.AddressInfo
		err = json.Unmarshal([]byte(addrJSON), &addr)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal address: %v", err)
		}
		addresses = append(addresses, addr)
	}

	return addresses, nil
}

// GetWildcardActor 获取一个指定 actorType 的随机 actor 地址
func (ab *AddressBook) GetWildcardActor(ctx context.Context, actorType string) (core.AddressInfo, error) {
	key := fmt.Sprintf(def.RedisAddressbookTyField+"%s", actorType)

	// 如果没有本地 actor，则随机获取一个
	addrJSON, err := trdredis.SRandMember(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return core.AddressInfo{}, fmt.Errorf("no actors found for type %s", actorType)
		}
		return core.AddressInfo{}, fmt.Errorf("GetWildcardActor SRandMember err %v", err)
	}

	var addr core.AddressInfo
	err = json.Unmarshal([]byte(addrJSON), &addr)
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("failed to unmarshal address: %v", err)
	}

	return addr, nil
}
