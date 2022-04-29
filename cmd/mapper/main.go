package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hj1980/bazel-proxy/mapper"
	"github.com/hj1980/bazel-proxy/protobuf/types/known/build_event_stream"
	"github.com/hj1980/bazel-proxy/protobuf/types/known/wrapper"
	build "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const maximumMessageSize = 10 * 1024 * 1024
const varintBufferSize = 8

var (
	filename = flag.String("file", "", "Binary wrapped events file to parse")
)

func main() {

	mappers := []mapper.BazelBuildEventMapper{
		mapper.NewTargetMapper(),
	}

	flag.Parse()

	if len(*filename) == 0 {
		flag.Usage()
		return
	}

	f, err := os.Open(*filename)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	for {

		viBuf := make([]byte, varintBufferSize)
		n, err := io.ReadFull(f, viBuf)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("Error when trying to populate viBuf: %s\n", err)
			return
		}

		v, n := protowire.ConsumeVarint(viBuf)
		if n < 0 {
			err = protowire.ParseError(n)
			log.Printf("Error when trying to consume varint: %s\n", err)
			return
		}
		if v > maximumMessageSize {
			err = fmt.Errorf("Maximum message size exceeded: %d", v)
			return
		}

		remainingBuf := make([]byte, v-varintBufferSize+uint64(n))
		_, err = io.ReadFull(f, remainingBuf)
		if err != nil {
			log.Printf("Error when trying to populate remainingBuf: %s\n", err)
			return
		}

		buf := append(viBuf[n:], remainingBuf...)
		e := &wrapper.PublishBuildEventWrapper{}
		err = proto.Unmarshal(buf, e)
		if err != nil {
			log.Printf("Error when trying to unmarshal: %s\n", err)
			return
		}

		switch vbe := e.Event.(type) {
		case *wrapper.PublishBuildEventWrapper_PublishBuildToolEventStreamRequest:
			switch vbbe := vbe.PublishBuildToolEventStreamRequest.OrderedBuildEvent.Event.Event.(type) {
			case *build.BuildEvent_BazelEvent:
				bbe := &build_event_stream.BuildEvent{}
				err = proto.Unmarshal(vbbe.BazelEvent.Value, bbe)
				if err != nil {
					log.Printf("Error when trying to unmarshal: %s\n", err)
					return
				}
				for _, mapper := range mappers {
					mapper.HandleBazelBuildEvent(vbe.PublishBuildToolEventStreamRequest.OrderedBuildEvent.Event, bbe)
				}
			}
		}
	}
}
