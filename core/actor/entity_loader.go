package actor

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pojol/braid/3rd/mgo"
	trhreids "github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/core"
	"github.com/pojol/braid/def"
	"github.com/pojol/braid/lib/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

type BlockLoader struct {
	BlockName string
	BlockType reflect.Type

	Ins      interface{}
	oldBytes []byte
}

// EntityLoader entity装载器
type EntityLoader struct {
	DBName string
	DBCol  string

	WrapperEntity core.IEntity
	Loaders       []BlockLoader
}

func BuildEntityLoader(dbName, dbCol string, wrapper core.IEntity) *EntityLoader {
	wrapperType := reflect.TypeOf(wrapper).Elem()
	loaders := make([]BlockLoader, 0)

	for i := 0; i < wrapperType.NumField(); i++ {
		field := wrapperType.Field(i)
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			bsonTag := field.Tag.Get("bson")
			blockName := strings.Split(bsonTag, ",")[0] // 获取 bson 标签的第一部分作为名称
			if blockName == "" {
				blockName = strings.ToLower(field.Name) // 如果没有 bson 标签，使用字段名的小写形式
			}
			loaders = append(loaders, BlockLoader{
				BlockName: blockName,
				BlockType: field.Type,
			})

		}
	}
	return &EntityLoader{DBName: dbName, DBCol: dbCol, WrapperEntity: wrapper, Loaders: loaders}
}

func (loader *EntityLoader) tryLoad2DB(ctx context.Context) error {
	collection := mgo.Collection(loader.DBName, loader.DBCol)
	if collection == nil {
		return def.ErrEntityLoadDBColNotFound(loader.WrapperEntity.GetID(), loader.DBName, loader.DBCol)
	}

	var entityDoc bson.M
	err := collection.FindOne(ctx, bson.M{"_id": loader.WrapperEntity.GetID()}).Decode(&entityDoc)
	if err != nil {
		return err
	}

	for idx, load := range loader.Loaders {
		if moduleData, ok := entityDoc[load.BlockName]; ok {
			bsonData, err := bson.Marshal(moduleData)
			if err != nil {
				return err
			}
			protoMsg := reflect.New(load.BlockType.Elem()).Interface().(proto.Message)
			if err := bson.Unmarshal(bsonData, protoMsg); err != nil {
				return err
			}

			loader.Loaders[idx].Ins = protoMsg
			loader.WrapperEntity.SetModule(load.BlockType, protoMsg)
		}
	}

	return nil
}

func (loader *EntityLoader) Load(ctx context.Context) error {
	if len(loader.Loaders) == 0 {
		return def.ErrEntityLoadEntityLoadersEmpty(loader.WrapperEntity.GetID())
	}

	var cmds []redis.Cmder

	cmds, err := trhreids.TxPipelined(ctx, "[EntityLoader.Load]", func(pipe redis.Pipeliner) error {
		for _, load := range loader.Loaders {
			key := fmt.Sprintf("{%s}_%s", loader.WrapperEntity.GetID(), load.BlockName)
			pipe.Get(ctx, key)
		}
		return nil
	})
	if err != nil {
		if err == redis.Nil {
			err = loader.tryLoad2DB(ctx)
			if err != nil {
				return err
			}

			// sync to redis
			return loader.Sync(ctx)
		} else {
			return err
		}
	}

	var bytSlice [][]byte
	bytSlice, err = trhreids.GetCmdsByteSlice(cmds)
	if err != nil {
		return err
	}

	for idx, load := range loader.Loaders {
		protoMsg := reflect.New(load.BlockType.Elem()).Interface().(proto.Message)

		if len(bytSlice[idx]) == 0 {
			return fmt.Errorf("load block %s is not empty", load.BlockName)
		}

		if err := proto.Unmarshal(bytSlice[idx], protoMsg); err != nil {
			return def.ErrEntityLoadUnpack(loader.WrapperEntity.GetID(), load.BlockName)
		}

		loader.Loaders[idx].oldBytes = bytSlice[idx]
		loader.Loaders[idx].Ins = protoMsg
		loader.WrapperEntity.SetModule(load.BlockType, protoMsg)
	}

	return nil
}

func (loader *EntityLoader) Sync(ctx context.Context) error {
	if len(loader.Loaders) == 0 {
		return def.ErrEntityLoadEntityLoadersEmpty(loader.WrapperEntity.GetID())
	}

	_, err := trhreids.TxPipelined(ctx, "[EntityLoader.Sync]", func(pipe redis.Pipeliner) error {
		for idx, load := range loader.Loaders {
			if loader.Loaders[idx].Ins == nil {
				log.Warn("sync %s Ins is nil", load.BlockName)
				continue
			}

			byt, err := proto.Marshal(loader.Loaders[idx].Ins.(proto.Message))
			if err != nil {
				return err
			}

			if !bytes.Equal(loader.Loaders[idx].oldBytes, byt) {
				loader.Loaders[idx].oldBytes = byt // update
				key := fmt.Sprintf("{%s}_%s", loader.WrapperEntity.GetID(), load.BlockName)
				pipe.Set(ctx, key, byt, 0)
			}
		}
		return nil
	})

	return err
}

func (loader *EntityLoader) Store(ctx context.Context) error {
	_, err := trhreids.TxPipelined(ctx, "[EntityLoader.Store]", func(pipe redis.Pipeliner) error {
		for _, load := range loader.Loaders {
			key := fmt.Sprintf("{%s}_%s", loader.WrapperEntity.GetID(), load.BlockName)
			pipe.Del(ctx, key)
		}
		return nil
	})

	return err
}

func (loader *EntityLoader) IsDirty() bool {
	for _, load := range loader.Loaders {

		byt, err := proto.Marshal(load.Ins.(proto.Message))
		if err != nil {
			return false
		}

		if !bytes.Equal(load.oldBytes, byt) {
			return true
		}
	}

	return false
}
func (loader *EntityLoader) GetModule(typ reflect.Type) interface{} {
	for _, load := range loader.Loaders {
		if load.BlockType == typ {
			return load.Ins
		}
	}
	return nil
}

func (loader *EntityLoader) SetModule(typ reflect.Type, module interface{}) {

	flag := false

	for idx, load := range loader.Loaders {
		if load.BlockType == typ {
			loader.Loaders[idx].Ins = module
			flag = true
		}
	}

	if !flag {
		log.Warn("set module not found", typ)
	}
}
