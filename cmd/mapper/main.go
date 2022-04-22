package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hj1980/bazel-proxy/protobuf/types/known/wrapper"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

const maximumMessageSize = 10 * 1024 * 1024
const varintBufferSize = 8

var (
	filename = flag.String("file", "", "Binary wrapped events file to parse")
)

func main() {
	//log.Printf("Mapper started\n")

	flag.Parse()

	if len(*filename) == 0 {
		flag.Usage()
		return
	}

	f, err := os.Open(*filename)
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
		//log.Printf("Read %d bytes\n", n)

		v, n := protowire.ConsumeVarint(viBuf)
		if n < 0 {
			err = protowire.ParseError(n)
			log.Printf("Error when trying to consume varint: %s\n", err)
			return
		}
		//log.Printf("Message is %d bytes, Varint was %d bytes\n", v, n)

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
		//log.Printf("remainingBuf is %d bytes, buf is %d bytes\n", len(remainingBuf), len(buf))

		// fmt.Printf("%+2x\n", buf)
		// fmt.Printf("%s\n", buf)

		e := &wrapper.PublishBuildEventWrapper{}
		err = proto.Unmarshal(buf, e)
		if err != nil {
			log.Printf("Error when trying to unmarshal: %s\n", err)
			return
		}

		fmt.Printf("%+v\n", e)

		// log.Printf("Message is %d bytes, Varint was %d bytes\n", v, n)
	}

	f.Close()
}
