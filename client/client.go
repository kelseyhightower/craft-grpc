// The client command issues RPCs to a Google server and prints the
// results.
//
// In "search" mode, client calls Search on the server and prints the
// results.
//
// In "watch" mode, client starts a Watch on the server and prints the
// result stream.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pb "github.com/kelseyhightower/craft-grpc/search"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	server = flag.String("server", "localhost:36060", "server address")
	mode   = flag.String("mode", "search", `one of "search" or "watch"`)
	query  = flag.String("query", "test", "query string")
)

func main() {
	flag.Parse()

	// Connect to the server.
	conn, err := grpc.Dial(*server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGoogleClient(conn)

	// Run the RPC.
	switch *mode {
	case "search":
		search(client, *query)
	case "watch":
		watch(client, *query)
	default:
		log.Fatalf("unknown mode: %q", *mode)
	}
}

// search issues a search for query and prints the result.
func search(client pb.GoogleClient, query string) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req := &pb.Request{Query: query}
	res, err := client.Search(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}

// watch runs a Watch RPC and prints the result stream.
func watch(client pb.GoogleClient, query string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req := &pb.Request{Query: query}
	stream, err := client.Watch(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("and now your watch is ended")
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
	}
}
