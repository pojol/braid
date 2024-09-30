package span

type SpanParm struct {
	nodeID string
}

type SpanOption func(*SpanParm)

func WithNodeID(id string) SpanOption {
	return func(sp *SpanParm) {
		sp.nodeID = id
	}
}
