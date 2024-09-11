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
	NodeID string
	Ip     string
	Port   int

	IDMap map[string]bool

	sync.RWMutex
}

func New(info core.AddressInfo) *AddressBook {
	return &AddressBook{
		IDMap:  make(map[string]bool),
		NodeID: info.Node,
		Ip:     info.Ip,
		Port:   info.Port,
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

	// serialize address info to json
	addrJSON, _ := json.Marshal(core.AddressInfo{
		ActorId: id,
		ActorTy: ty,
		Ip:      ab.Ip,
		Port:    ab.Port},
	)
	// execute multiple redis operations using pipeline
	pipe := trdredis.Pipeline()
	pipe.HSet(ctx, def.RedisAddressbookIDField, id, addrJSON)
	pipe.SAdd(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", ty), addrJSON)

	// 更新节点记录
	nodeKey := fmt.Sprintf("node:%s", ab.NodeID)
	pipe.HIncrBy(ctx, nodeKey, fmt.Sprintf("actor:%s", ty), 1)
	pipe.HIncrBy(ctx, nodeKey, "total_weight", 1)

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

	// get address info first
	addrJSON, err := trdredis.HGet(ctx, def.RedisAddressbookIDField, id).Result()
	if err != nil {
		return fmt.Errorf("address not found for id: %s", id)
	}

	info := &core.AddressInfo{}
	err = json.Unmarshal([]byte(addrJSON), info)
	if err != nil {
		return fmt.Errorf("addressbook.unregister json unmarshal err %v", err.Error())
	}

	// execute multiple redis operations using pipeline
	pipe := trdredis.Pipeline()
	pipe.HDel(ctx, def.RedisAddressbookIDField, id)
	pipe.SRem(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", info.ActorTy), addrJSON)

	// 更新节点记录
	nodeKey := fmt.Sprintf("node:%s", ab.NodeID)
	pipe.HIncrBy(ctx, nodeKey, fmt.Sprintf("actor:%s", info.ActorTy), -1)
	pipe.HIncrBy(ctx, nodeKey, "total_weight", -1)

	_, err = pipe.Exec(ctx)
	if err == nil {
		ab.Lock()
		delete(ab.IDMap, id) // try delete
		ab.Unlock()
	}

	return err
}

// GetByID get actor address by id
func (ab *AddressBook) GetByID(ctx context.Context, id string) (core.AddressInfo, error) {

	if id == "" {
		return core.AddressInfo{}, fmt.Errorf("actor id or type is empty")
	}

	ab.RLock()
	if _, ok := ab.IDMap[id]; ok {
		ab.RUnlock()
		return core.AddressInfo{Node: ab.NodeID, Ip: ab.Ip, Port: ab.Port}, nil // return local node info directly
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

// GetByType get actor address by type
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

const PickLimit = 10

// GetWildcardActor retrieves a random actor address of the specified actorType
func (ab *AddressBook) GetWildcardActor(ctx context.Context, actorType string) (core.AddressInfo, error) {
	key := fmt.Sprintf(def.RedisAddressbookTyField+"%s", actorType)

	// get a random one
	addrJSONs, err := trdredis.SRandMemberN(ctx, key, PickLimit).Result()
	if err != nil {
		if err == redis.Nil {
			return core.AddressInfo{}, fmt.Errorf("no actors found for type %s", actorType)
		}
		return core.AddressInfo{}, fmt.Errorf("GetWildcardActor SRandMember err %v", err)
	}

	// unmarshal
	if len(addrJSONs) == 0 {
		return core.AddressInfo{}, fmt.Errorf("no actors found for type %s", actorType)
	}

	var lowestWeightAddr core.AddressInfo
	lowestWeight := int(^uint(0) >> 1) // // Maximum int value, used as a sentinel to check if a valid weighted node address has been found

	for _, addrJSON := range addrJSONs {
		var addr core.AddressInfo
		if err := json.Unmarshal([]byte(addrJSON), &addr); err != nil {
			continue // continue to next address
		}

		// get the weight of the node where the actor is located
		nodeKey := fmt.Sprintf("node:%s", addr.Ip)
		nodeWeight, err := trdredis.HGet(ctx, nodeKey, "total_weight").Int()
		if err != nil {
			continue // skip this actor if unable to get node weight
		}

		if nodeWeight < lowestWeight {
			lowestWeight = nodeWeight
			lowestWeightAddr = addr
		}
	}

	if lowestWeight == int(^uint(0)>>1) {
		return core.AddressInfo{}, fmt.Errorf("no valid actors found for type %s", actorType)
	}

	return lowestWeightAddr, nil
}
