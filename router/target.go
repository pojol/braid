package router

type Target struct {
	ID string // Unique identifier (can also use def.Symbol to represent special routing methods)
	Ty string
	Ev string
}
