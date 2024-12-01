package mock

import "time"

type tccState struct {
	originValue  int
	currentValue int
	status       string // "try", "confirmed", "cancelled"
	createdAt    time.Time
}

type TCC struct {
	stateMap map[string]*tccState
}
