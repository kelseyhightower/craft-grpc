// The backend command runs a Google server that returns fake results.
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	pb "github.com/kelseyhightower/craft-grpc/search"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
)

var hostname string

type server struct{}

// randomDuration returns a random duration up to max, at intervals of max/10.
func randomDuration(max time.Duration) time.Duration {
	return time.Duration(1+int64(rand.Intn(10))) * (max / 10)
}

// Search sleeps for a random interval then returns a string
// identifying the query and this backend.
func (s *server) Search(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	d := randomDuration(100 * time.Millisecond)
	logSleep(ctx, d)
	select {
	case <-time.After(d):
		return &pb.Result{
			Title: fmt.Sprintf("result for [%s] from backend %s", req.Query, hostname),
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func logSleep(ctx context.Context, d time.Duration) {
	if tr, ok := trace.FromContext(ctx); ok {
		tr.LazyPrintf("sleeping for %s", d)
	}
}

// Watch returns a stream of results identifying the query and this
// backend, sleeping a random interval between each send.
func (s *server) Watch(req *pb.Request, stream pb.Google_WatchServer) error {
	ctx := stream.Context()
	for i := 0; ; i++ {
		d := randomDuration(1 * time.Second)
		logSleep(ctx, d)
		select {
		case <-time.After(d):
			err := stream.Send(&pb.Result{
				Title: fmt.Sprintf("result %d for [%s] from backend %s", i, req.Query, hostname),
			})
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func main() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}
	log.Printf("Starting backend %s...", hostname)

	rand.Seed(time.Now().UnixNano())
	go http.ListenAndServe(":36661", nil)
	lis, err := net.Listen("tcp", ":36061")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	g := grpc.NewServer()
	pb.RegisterGoogleServer(g, new(server))
	g.Serve(lis)
}
