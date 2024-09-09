package entitytest

import fmt "fmt"

func (b *EntityBagModule) produce(item *Item) {
	_, ok := b.Bag[item.DictID]

	// 检查是否为自动恢复类道具

	if !ok {
		b.Bag[item.DictID] = &ItemList{
			Items: []*Item{item},
		}
	} else {
		// 查看是否可堆叠
		// 查看是否数量溢出

	}
}

func (b *EntityBagModule) consume(dict bool, item *Item, num int32) (Item, error) {
	items, ok := b.Bag[item.DictID]
	refreshItem := Item{}

	if !ok {
		return refreshItem, fmt.Errorf("not found")
	}

	var rmv bool // 移除标签
	var idx int
	var flag bool

	for k, v := range items.Items {
		equal := false

		if dict {
			equal = v.DictID == item.DictID
		} else {
			equal = v.ID == item.ID
		}

		if equal {
			if v.Num < num {
				return refreshItem, fmt.Errorf("not enough")
			}

			items.Items[k].Num -= item.Num
			if items.Items[k].Num == 0 /* not resource */ {
				rmv = true
				idx = k
			}

			flag = true
			break
		}
	}

	if !flag {
		return refreshItem, fmt.Errorf("not found")
	}

	if rmv {
		items.Items = append(items.Items[:idx], items.Items[idx+1:]...)
		if len(items.Items) == 0 {
			delete(b.Bag, item.DictID)
		}
	}

	return refreshItem, nil
}

func (b *EntityBagModule) enough(dict bool, item *Item) bool {
	items, ok := b.Bag[item.DictID]
	if !ok {
		return false
	}

	for _, v := range items.Items {
		equal := false

		if dict {
			equal = v.DictID == item.DictID
		} else {
			equal = v.ID == item.ID
		}

		if equal {

			b._checkRecoverIimeItem()
			b._checkTimeoutItem()

			if v.Num >= item.Num {
				return true
			}
			break
		}
	}

	return false
}

// _checkRecoverIimeItem - 对自动恢复类的道具进行检查
func (b *EntityBagModule) _checkRecoverIimeItem() {

}

// _checkTimeoutItem - 对可能超时的道具进行检查
func (b *EntityBagModule) _checkTimeoutItem() {

}

// EnoughItemWithInsID - 通过 实例ID(唯一) 判断背包中道具是否足够，注：通过实例ID进行判定同时需要传 字典ID
func (b *EntityBagModule) EnoughItemWithInsID(id string, dictid, num int32) bool {
	return b.enough(false, &Item{ID: id, DictID: dictid, Num: num})
}

// EnoughItem - 通过 字典ID 判断背包中道具是否足够
func (b *EntityBagModule) EnoughItem(id, num int32) bool {
	return b.enough(true, &Item{DictID: id, Num: num})
}

// EnoughItems - 通过 字典ID 判断背包中道具是否足够
func (b *EntityBagModule) EnoughItems(items []*Item) bool {
	for _, v := range items {
		if !b.enough(true, v) {
			return false
		}
	}

	return true
}

func (b *EntityBagModule) ProduceItem(item *Item, num uint32, reason, detail string) {

}

func (b *EntityBagModule) ProduceItems(items []*Item, reason, detail string) {

}

// ExistItem - 按字典id获取背包中的道具（注：如果
func (b *EntityBagModule) GetItemNum(id int32) int64 {
	var num int64

	if _, ok := b.Bag[id]; ok {

		for _, v := range b.Bag[id].Items {
			num += int64(v.Num)
		}

	}

	return num
}

// ConsumeItem - 消耗道具（注意消耗之前一定要进行 enough 判定）
func (b *EntityBagModule) ConsumeItem(id, num int32, reason, detail string) []*Item {
	refresh := []*Item{}

	ritem, err := b.consume(true, &Item{DictID: id, Num: num}, num)
	if err != nil {
		fmt.Errorf("consume item %v num %v reason %v error: %w", id, num, reason, err)
	}

	refresh = append(refresh, &Item{
		ID:     ritem.ID,
		DictID: ritem.DictID,
		Num:    ritem.Num,
	})

	// log 记录

	return refresh
}

// ConsumeItems - 消耗道具（注意消耗之前一定要进行 enough 判定）
func (b *EntityBagModule) ConsumeItems(items []*Item, reason, detail string) []*Item {

	refresh := []*Item{}

	for _, v := range items {
		ritem, err := b.consume(true, v, v.Num)
		if err != nil {
			fmt.Printf("consume item %v num %v reason %v error: %v", v.DictID, v.Num, reason, err)
		}

		refresh = append(refresh, &Item{
			ID:     ritem.ID,
			DictID: ritem.DictID,
			Num:    ritem.Num,
		})
	}

	// log 记录

	return refresh
}
