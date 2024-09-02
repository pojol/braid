package net

type IAcceptor interface {
	Start() error
	Stop() error
	Accept() (ISession, error)
}

type IConnector interface {
	Connect(address string) (ISession, error)
	Disconnect() error
}

type ISession interface {
	Read() ([]byte, error)
	Write([]byte) error
	Close() error
}
