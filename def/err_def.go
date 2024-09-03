package def

import "github.com/pojol/braid/lib/errcode"

var (
	ErrActorRepeatRegisterEvent = func(args ...interface{}) errcode.Code {
		return errcode.Add(-10000, " [actor] repeat register event", args...)
	}

	ErrMsgNonMessageType = func(args ...interface{}) errcode.Code {
		return errcode.Add(-11001, " [message] non message type", args...)
	}
	ErrMsgNoHandler = func(args ...interface{}) errcode.Code {
		return errcode.Add(-11002, " [message] non message handler", args...)
	}

	ErrSystemCantFindCreateActorStrategy = func(args ...interface{}) errcode.Code {
		return errcode.Add(-12000, " [system] can't find create actor strategy", args...)
	}
	ErrSystemCantFindLocalActor = func(args ...interface{}) errcode.Code {
		return errcode.Add(-12001, " [system] can't find local actor", args...)
	}
	ErrSystemRepeatRegistActor = func(args ...interface{}) errcode.Code {
		return errcode.Add(-12002, " [system] repeat regist actor", args...)
	}
	ErrSystemParm = func(args ...interface{}) errcode.Code {
		return errcode.Add(-12003, " [system] regist actor parm err", args...)
	}
	ErrSystemUnknowRemoteAddr = func(args ...interface{}) errcode.Code {
		return errcode.Add(-12004, " [system] address book can't find actor ", args...)
	}

	ErrAddressbookCheck = func(args ...interface{}) errcode.Code {
		return errcode.Add(-13000, " [addressbook] check", args...)
	}
)
