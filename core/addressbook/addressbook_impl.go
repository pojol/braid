package addressbook

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"

	trdredis "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/dismutex"
	"github.com/pojol/braid/lib/log"
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

func makeNodeKey(nodid string) string {
	return fmt.Sprintf("{node:%s}", nodid)
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
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), fmt.Sprintf("actor:%s", ty), 1)
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), "total_weight", 1)

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
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), fmt.Sprintf("actor:%s", info.ActorTy), -1)
	pipe.HIncrBy(ctx, makeNodeKey(ab.NodeID), "total_weight", -1)

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
			log.Warn("addressbook unmarshal actor type %v json err %v", actorType, err.Error())
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
	// Use a distributed lock to ensure consistency
	mu := &dismutex.Mutex{Token: fmt.Sprintf("low_weight_node:%s", actorType)}
	err := mu.Lock(ctx, "[addressbook.GetLowWeightNodeForActor]")
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("GetLowWeightNodeForActor get distributed mutex err %v", err.Error())
	}
	defer mu.Unlock(ctx)

	// Get all node infos from the set
	nodeInfoMap, err := trdredis.HGetAll(ctx, def.RedisAddressbookNodesField).Result()
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("GetLowWeightNodeForActor SMembers err %v", err)
	}

	if len(nodeInfoMap) == 0 {
		return core.AddressInfo{}, fmt.Errorf("no nodes found")
	}

	type weightedNode struct {
		addr       core.AddressInfo
		weight     int
		actorCount int
	}

	var weightedNodes []weightedNode
	pipe := trdredis.Pipeline()

	for nodeID, nodeInfoJSON := range nodeInfoMap {
		var nodeInfo core.AddressInfo
		if err := json.Unmarshal([]byte(nodeInfoJSON), &nodeInfo); err != nil {
			log.Warn("unable to unmarshal node info: %v", err)
			continue
		}
		pipe.HMGet(ctx, makeNodeKey(nodeID), "total_weight", fmt.Sprintf("actor:%s", actorType))
	}

	// Execute pipeline
	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return core.AddressInfo{}, fmt.Errorf("pipeline execution failed: %v", err)
	}

	// Process results
	i := 0
	for _, nodeInfoJSON := range nodeInfoMap {
		if i >= len(cmders) {
			break
		}

		result, err := cmders[i].(*redis.SliceCmd).Result()
		if err != nil {
			log.Warn("unable to get node info: %v", err)
			i++
			continue
		}

		if len(result) < 2 {
			log.Warn("unexpected result length: %d", len(result))
			i++
			continue
		}

		weight := 0
		if weightStr, ok := result[0].(string); ok {
			weight, _ = strconv.Atoi(weightStr)
		}

		actorCount := 0
		if countStr, ok := result[1].(string); ok {
			actorCount, _ = strconv.Atoi(countStr)
		}

		var nodeInfo core.AddressInfo
		if err := json.Unmarshal([]byte(nodeInfoJSON), &nodeInfo); err != nil {
			log.Warn("unable to unmarshal node info: %v", err)
			i++
			continue
		}

		weightedNodes = append(weightedNodes, weightedNode{
			addr:       nodeInfo,
			weight:     weight,
			actorCount: actorCount,
		})

		i++
	}

	// Sort nodes by weight
	sort.Slice(weightedNodes, func(i, j int) bool {
		return weightedNodes[i].weight < weightedNodes[j].weight
	})

	// Select the lowest weight nodes
	lowWeightNodes := weightedNodes
	if len(weightedNodes) > LowWeightNodeLimit {
		lowWeightNodes = weightedNodes[:LowWeightNodeLimit]
	}

	// Find the node with the lowest actor count among the low weight nodes
	var selectedAddr core.AddressInfo
	lowestActorCount := int(^uint(0) >> 1) // Max int value

	for _, node := range lowWeightNodes {
		if node.actorCount < lowestActorCount {
			lowestActorCount = node.actorCount
			selectedAddr = node.addr
		}
	}

	// If the weight of the current actor type is greater than the available weight of the selected node, return a failure
	// if actorTypeWeight > selectedAddr.availableWeight {
	//	 return core.AddressInfo{}, fmt.Errorf("no node with sufficient capacity found for actor type %s", actorType)
	// }

	if selectedAddr.Node == "" {
		return core.AddressInfo{}, fmt.Errorf("no suitable node found for actor type %s", actorType)
	}

	return selectedAddr, nil
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
