package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"

	"github.com/hj1980/bazel-proxy/protobuf/types/known/wrapper"
	"github.com/hj1980/bazel-proxy/writer"
	build "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	listen     = flag.String("listen", "127.0.0.1:50051", "Local address to listen on eg 0.0.0.0:12345")
	upstream   = flag.String("upstream", "127.0.0.1:1985", "Upstream address eg 1.2.3.4:9000")
	grpcServer *grpc.Server
)

type BuildServer struct {
	build.UnimplementedPublishBuildEventServer
	Client build.PublishBuildEventClient
	writer writer.Writer
}

func (s *BuildServer) PublishLifecycleEvent(ctx context.Context, in *build.PublishLifecycleEventRequest) (out *emptypb.Empty, err error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		log.Printf("PublishLifecycleEvent called but could not get peer information\n")
	} else {
		log.Printf("PublishLifecycleEvent called from: %s\n", p.Addr)
		// fmt.Printf("\t%s\n", in.BuildEvent)

		s.writer.AppendWrappedData(&wrapper.PublishBuildEventWrapper{
			Event: &wrapper.PublishBuildEventWrapper_PublishLifecycleEventRequest{in},
		})

		// fmt.Printf("%s\n", e)
	}

	clientCtx := context.Background()
	out, err = s.Client.PublishLifecycleEvent(clientCtx, in)

	return
}

func (s *BuildServer) PublishBuildToolEventStream(downstream build.PublishBuildEvent_PublishBuildToolEventStreamServer) (err error) {
	ctx := downstream.Context()
	p, ok := peer.FromContext(ctx)
	if !ok {
		log.Printf("PublishBuildToolEventStream called but could not get peer information\n")
	} else {
		log.Printf("PublishBuildToolEventStream called from: %s\n", p.Addr)
	}

	// Create client for upstream connection
	clientCtx := context.Background()
	upstream, err := s.Client.PublishBuildToolEventStream(clientCtx)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}

	// Channels for stream closures
	downstreamErr := make(chan error, 1)
	upstreamErr := make(chan error, 1)

	// Proxy downstream -> upstream
	go func() {
		for {
			req, rerr := downstream.Recv()
			//fmt.Printf("downstream.Recv: %s\n", rerr)

			if rerr == io.EOF {
				cserr := upstream.CloseSend()
				//fmt.Printf("upstream.CloseSend: %s\n", cserr)
				//fmt.Printf("downstreamErr <- %s\n", cserr)
				downstreamErr <- cserr
				return
			}

			if rerr != nil {
				//fmt.Printf("downstreamErr <- %s\n", rerr)
				downstreamErr <- rerr
				return
			}

			//fmt.Printf("\t%s\n", req.OrderedBuildEvent.Event)
			s.writer.AppendWrappedData(&wrapper.PublishBuildEventWrapper{
				Event: &wrapper.PublishBuildEventWrapper_PublishBuildToolEventStreamRequest{req},
			})

			// log.Printf("Proxying upstream: %s", req)
			serr := upstream.Send(req)
			//fmt.Printf("upstream.Send: %s\n", serr)
			if serr != nil {
				downstreamErr <- serr
				return
			}
		}
	}()

	// Proxy upstream -> downstream
	go func() {
		for {
			res, rerr := upstream.Recv()
			//fmt.Printf("upstream.Recv: %s\n", rerr)

			if rerr == io.EOF {
				//fmt.Printf("from upstream: %s\n", rerr)
				upstreamErr <- nil
				return
			}

			if rerr != nil {
				//fmt.Printf("upsteamErr <- %s\n", rerr)
				upstreamErr <- rerr
				return
			}

			// log.Printf("Proxying downstream: %s", req)
			serr := downstream.Send(res)
			//fmt.Printf("downstream.Send: %s\n", rerr)
			if serr != nil {
				upstreamErr <- serr
				return
			}
		}
	}()

	// Wait for both streams to end
	for i := 0; i < 2; i++ {
		select {
		case err := <-downstreamErr:
			if err != nil {
				log.Printf("PublishBuildToolEventStream downstreamErr<-: %s\n", err)
			}
		case err := <-upstreamErr:
			if err != nil {
				if err != nil {
					log.Printf("PublishBuildToolEventStream upstreamErr<-: %s\n", err)
				}
			}
		}
	}

	log.Printf("PublishBuildToolEventStream finished\n")
	return
}

func main() {

	flag.Parse()

	log.Printf("dialing upstream: %v\n", *upstream)
	conn, err := grpc.Dial(*upstream, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to dial upstream: %v", err)
	}
	log.Printf("connected to upstream: %v\n", *upstream)

	log.Printf("attempting to listen: %v\n", *listen)
	lis, err := net.Listen("tcp4", *listen)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("listening: %s\n", lis.Addr())

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dataPath := wd + "/builds"
	err = os.Mkdir(dataPath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	bs := &BuildServer{
		writer: writer.NewWindowedDataWriter(dataPath),
	}
	bs.Client = build.NewPublishBuildEventClient(conn)

	grpcServer = grpc.NewServer()
	build.RegisterPublishBuildEventServer(grpcServer, bs)

	log.Printf("serving\n")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
