package addressbook

type IAddressBook interface {
	Regist(string, string) // 将 actor 注册到 address book；
	GetAddrInfo(actorid string) (*AddressInfo, error)
}

type AddressInfo struct {
	Node    string // node id
	Service string // service name
	Ip      string // ip
	Port    int    // port
}

type AddressBook struct {
	ServiceName string
	NodeID      string
}

func (ab *AddressBook) Regist(ty, id string) {
	// 检查 id 是否已经存在

}

func (ab *AddressBook) GetAddrInfo(actorid string) (*AddressInfo, error) {

	info := &AddressInfo{
		Node:    ab.NodeID,
		Service: ab.ServiceName,
		Ip:      "127.0.0.1",
		Port:    14222,
	}

	// 先在本地检查
	// 再去 redis 检查

	return info, nil //def.ErrSystemUnknowRemoteAddr(actorid)

}
