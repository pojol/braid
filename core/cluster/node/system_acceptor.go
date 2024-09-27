package node

import (
	context "context"
	fmt "fmt"
	"runtime"
	"strconv"

	"github.com/pojol/braid/core"
	"github.com/pojol/braid/lib/grpc"
	"github.com/pojol/braid/router"

	realgrpc "google.golang.org/grpc"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
)

type acceptor struct {
	server *grpc.Server
}

var acceptorptr *acceptor

type listen struct {
	router.AcceptorServer
	sys core.ISystem
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

func acceptorInit(sys core.ISystem, port int) {

	acceptorptr = &acceptor{
		server: grpc.BuildServerWithOption(
			grpc.WithServerListen(":"+strconv.Itoa(port)),
			grpc.WithServerGracefulStop(),
			grpc.ServerRegisterHandler(func(s *realgrpc.Server) {
				router.RegisterAcceptorServer(s, &listen{sys: sys})
			}),
			grpc.ServerAppendUnaryInterceptors(grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(recoverHandler))),
		),
	}

	err := acceptorptr.server.Init()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize acceptor server: %v", err))
	}
}

func acceptorUpdate() {
	acceptorptr.server.Run()
}

func acceptorExit() {
	acceptorptr.server.Close()
}

// acceptor routing
func (s *listen) Routing(ctx context.Context, msg *router.RouteReq) (*router.RouteRes, error) {
	res := &router.RouteRes{}
	warpper := &router.MsgWrapper{
		Ctx: ctx,
		Req: msg.Msg,
		Res: &router.Message{
			Header: &router.Header{
				Custom: make(map[string]string),
			},
		},
	}

	warpper.Req.Header.PrevActorType = "GrpcAcceptor"

	err := s.sys.Call(router.Target{
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
