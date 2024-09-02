package actor

import (
	"braid/lib/grpc"
	"braid/router"
	context "context"
	fmt "fmt"
	"runtime"

	realgrpc "google.golang.org/grpc"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
)

type acceptor struct {
	server *grpc.Server
}

var acceptorptr *acceptor

type listen struct {
	router.AcceptorServer
}

// Stack returns a formatted stack trace of the goroutine that calls it.
// It calls runtime.Stack with a large enough buffer to capture the entire trace.
// If all is true, Stack formats stack traces of all other goroutines
// into buf after the trace for the current goroutine.
func stack(all bool) []byte {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, all)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}

func recoverHandler(r interface{}) error {
	err, ok := r.(error)
	if !ok {
		err = fmt.Errorf("%v", r)
	}
	buf := stack(false)
	fmt.Println(fmt.Errorf("PANIC: %v\n%s", err, buf).Error())
	return fmt.Errorf("[GRPC-SERVER RECOVER] err: %v stack: %s", err, buf)
}

func NewAcceptor(port string) {

	acceptorptr = &acceptor{
		server: grpc.BuildServerWithOption(
			grpc.WithServerListen(":"+port),
			grpc.WithServerGracefulStop(),
			grpc.ServerRegisterHandler(func(s *realgrpc.Server) {
				router.RegisterAcceptorServer(s, &listen{})
			}),
			grpc.ServerAppendUnaryInterceptors(grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(recoverHandler))),
		),
	}

	acceptorptr.server.Init()
}

func Update() {
	acceptorptr.server.Run()
}

func Exit() {
	acceptorptr.server.Close()
}

// acceptor routing
func (s *listen) Routing(ctx context.Context, msg *router.RouteReq) (*router.RouteRes, error) {
	res := &router.RouteRes{}

	warpper := &router.MsgWrapper{
		Req: msg.Msg,
		Res: &router.Message{
			Header: &router.Header{
				Custom: make(map[string]string),
			},
		},
	}

	err := Call(ctx, router.Target{
		ID: msg.Msg.Header.TargetActorID,
		Ty: msg.Msg.Header.TargetActorType,
		Ev: msg.Msg.Header.Event,
	}, warpper)

	if err != nil {
		fmt.Println("listen routing", msg.Msg.Header.Event, "err", err.Error())
	}

	res.Msg = warpper.Res
	return res, nil
}
