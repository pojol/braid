package addressbook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	trdredis "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/dismutex"
	"github.com/pojol/braid/lib/log"
	"github.com/redis/go-redis/v9"
)

var (
	ErrUnknownActor = errors.New("unknown actor")
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

func makeNodeKey(nodid string) string {
	return fmt.Sprintf("{node:%s}", nodid)
}

func (ab *AddressBook) Register(ctx context.Context, ty, id string, weight int) error {

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
		Node:    ab.NodeID,
		ActorId: id,
		ActorTy: ty,
		Ip:      ab.Ip,
		Port:    ab.Port},
	)

	// serialize node info to json (without ActorId and ActorTy)
	nodeInfoJSON, _ := json.Marshal(core.AddressInfo{
		Node: ab.NodeID,
		Ip:   ab.Ip,
		Port: ab.Port},
	)

	// execute multiple redis operations using pipeline
	pipe := trdredis.Pipeline()
	pipe.HSet(ctx, def.RedisAddressbookIDField, id, addrJSON)
	pipe.SAdd(ctx, fmt.Sprintf(def.RedisAddressbookTyField+"%s", ty), addrJSON)

	// Add node info to a separate set
	pipe.HSet(ctx, def.RedisAddressbookNodesField, ab.NodeID, nodeInfoJSON)

	// 更新节点记录
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), fmt.Sprintf("actor:%s", ty), int64(weight))
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), "total_weight", int64(weight))

	_, err = pipe.Exec(ctx)

	if err != nil {
		return fmt.Errorf("redis pipeline exec err %v", err.Error())
	}

	ab.Lock()
	ab.IDMap[id] = true
	ab.Unlock()

	return nil
}

func (ab *AddressBook) Unregister(ctx context.Context, id string, weight int) error {
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
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), fmt.Sprintf("actor:%s", info.ActorTy), int64(-weight))
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), "total_weight", int64(-weight))

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
		if err == redis.Nil {
			return core.AddressInfo{}, ErrUnknownActor
		}
		return core.AddressInfo{}, fmt.Errorf("[braid.addressbook] get by id %s hget err: %s", id, err.Error())
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

const (
	PickLimit          = 10
	LowWeightNodeLimit = 10 // Number of low weight nodes to consider
)

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
			log.WarnF("addressbook unmarshal actor type %v json err %v", actorType, err.Error())
			continue // continue to next address
		}

		// get the weight of the node where the actor is located
		nodeWeight, err := trdredis.HGet(ctx, makeNodeKey(addr.Node), "total_weight").Int()
		if err != nil {
			fmt.Println("skip this actor if unable to get node weight")
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

// GetLowWeightNodeForActor retrieves a low-weight node address with fewer actors of the specified type
func (ab *AddressBook) GetLowWeightNodeForActor(ctx context.Context, actorType string) (core.AddressInfo, error) {
	// 获取所有节点信息
	nodeInfoMap, err := trdredis.HGetAll(ctx, def.RedisAddressbookNodesField).Result()
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("failed to get node infos: %v", err)
	}

	if len(nodeInfoMap) == 0 {
		return core.AddressInfo{}, fmt.Errorf("no nodes found")
	}

	var selectedNode core.AddressInfo
	lowestWeight := int(^uint(0) >> 1) // Max int value

	pipe := trdredis.Pipeline()

	// 使用 pipeline 批量获取节点权重
	for nodeID := range nodeInfoMap {
		pipe.HGet(ctx, makeNodeKey(nodeID), "total_weight")
	}

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("pipeline execution failed: %v", err)
	}

	i := 0
	for nodeID, nodeInfoJSON := range nodeInfoMap {
		if i >= len(cmders) {
			break
		}

		weightStr, err := cmders[i].(*redis.StringCmd).Result()
		if err != nil {
			log.WarnF("unable to get weight for node %s: %v", nodeID, err)
			i++
			continue
		}

		weight, _ := strconv.Atoi(weightStr)

		if weight < lowestWeight {
			var nodeInfo core.AddressInfo
			if err := json.Unmarshal([]byte(nodeInfoJSON), &nodeInfo); err != nil {
				log.WarnF("unable to unmarshal node info: %v", err)
				i++
				continue
			}
			lowestWeight = weight
			selectedNode = nodeInfo
		}

		i++
	}

	if selectedNode.Node == "" {
		return core.AddressInfo{}, fmt.Errorf("no suitable node found")
	}

	return selectedNode, nil
}

// GetActorTypeCount retrieves the count of registered actors of the specified type
func (ab *AddressBook) GetActorTypeCount(ctx context.Context, actorType string) (int64, error) {
	key := fmt.Sprintf(def.RedisAddressbookTyField+"%s", actorType)

	count, err := trdredis.SCard(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist, return 0
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get count for actor type %s: %v", actorType, err)
	}

	// count will be 0 for empty sets
	return count, nil
}

func (ab *AddressBook) Clear(ctx context.Context) error {
	mu := &dismutex.Mutex{Token: ab.NodeID}
	err := mu.Lock(ctx, "[addressbook.register]")
	if err != nil {
		return fmt.Errorf("addressbook.register get distributed mutex err %v", err.Error())
	}
	defer mu.Unlock(ctx)

	// 获取该节点的所有 actor 信息
	nodeKey := makeNodeKey(ab.NodeID)
	actorInfos, err := trdredis.HGetAll(ctx, nodeKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get actor infos: %w", err)
	}

	pipe := trdredis.Pipeline()

	// 删除该节点的所有 actor 信息
	for actorType := range actorInfos {
		if actorType == "total_weight" {
			continue
		}
		actorTypeKey := fmt.Sprintf(def.RedisAddressbookTyField+"%s", actorType[6:]) // 去掉 "actor:" 前缀

		// 获取该类型的所有 actor
		actors, err := trdredis.SMembers(ctx, actorTypeKey).Result()
		if err != nil {
			return fmt.Errorf("failed to get actors of type %s: %w", actorType, err)
		}

		for _, actorJSON := range actors {
			var actor core.AddressInfo
			if err := json.Unmarshal([]byte(actorJSON), &actor); err != nil {
				continue
			}
			if actor.Node == ab.NodeID {
				pipe.SRem(ctx, actorTypeKey, actorJSON)
				pipe.HDel(ctx, def.RedisAddressbookIDField, actor.ActorId)
			}
		}
	}

	pipe.Del(ctx, nodeKey)
	pipe.HDel(ctx, def.RedisAddressbookNodesField, ab.NodeID)

	// 执行 pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear node data: %w", err)
	}

	return nil
}
