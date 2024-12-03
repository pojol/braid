package def

import "github.com/pojol/braid/router/msg"

const (
	KeyActorID       = "ActorID"
	KeyActorTy       = "ActorTy"
	KeyTranscationID = "TransactionID"
)

func ActorID(id string) msg.Attr       { return msg.Attr{Key: KeyActorID, Value: id} }
func ActorTy(ty string) msg.Attr       { return msg.Attr{Key: KeyActorTy, Value: ty} }
func TransactionID(id string) msg.Attr { return msg.Attr{Key: KeyTranscationID, Value: id} }
