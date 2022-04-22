package writer

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hj1980/bazel-proxy/protobuf/types/known/wrapper"
	build "google.golang.org/genproto/googleapis/devtools/build/v1"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type Writer interface {
	AppendWrappedData(*wrapper.PublishBuildEventWrapper) error
}

// Lockable to avoid closing while writing data to file
type destination struct {
	sync.Mutex
	file  *os.File
	timer *time.Timer
}

// Lockable for concurrent access to destinations map
type WindowedDataWriter struct {
	sync.Mutex
	DataPath     string
	destinations map[string]*destination
}

func NewWindowedDataWriter(dataPath string) (w *WindowedDataWriter) {
	w = &WindowedDataWriter{
		DataPath:     dataPath,
		destinations: make(map[string]*destination),
	}
	return
}

func generateLongFilenameFromStreamId(id *build.StreamId) (name string, err error) {
	if len(id.BuildId) == 0 {
		err = fmt.Errorf("BuildId not specified in StreamId\n")
		return
	}
	name = id.BuildId + "/" + id.Component.String()
	if len(id.InvocationId) > 0 {
		name += "/" + id.InvocationId
	}
	return
}

func generateShortFilenameFromStreamId(id *build.StreamId) (name string, err error) {
	if len(id.BuildId) == 0 {
		err = fmt.Errorf("BuildId not specified in StreamId\n")
		return
	}
	name = id.BuildId
	return
}

func (w *WindowedDataWriter) appendToFile(filename string, buf []byte) (err error) {
	w.Lock()
	defer w.Unlock()

	d, ok := w.destinations[filename]
	if !ok {
		d = &destination{}
		fmt.Printf("Opening: %s\n", filename)
		d.file, err = os.OpenFile(w.DataPath+"/"+filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return
		}
		d.timer = time.AfterFunc(time.Second*5, func() {
			d.Lock()
			defer d.Unlock()

			fmt.Printf("Closing: %s\n", filename)
			err = d.file.Close()
			if err != nil {
				fmt.Printf("Error when closing file: %s\n", err)
			}
			delete(w.destinations, filename)
		})
		w.destinations[filename] = d
	}

	fmt.Printf("Appending: %d bytes to %s\n", len(buf), filename)
	d.Lock()
	defer d.Unlock()
	rawBuf := protowire.AppendBytes(nil, buf)
	fmt.Printf("Writing: %d bytes to %s\n", len(rawBuf), filename)
	_, err = d.file.Write(rawBuf)
	return
}

func (w *WindowedDataWriter) AppendWrappedData(d *wrapper.PublishBuildEventWrapper) (err error) {
	// fmt.Printf("Appending data: %T\n", d.Event)
	var filename string

	switch e := d.Event.(type) {
	case *wrapper.PublishBuildEventWrapper_PublishLifecycleEventRequest:
		// fmt.Printf("%+v\n", e.PublishLifecycleEventRequest.BuildEvent.StreamId)
		filename, err = generateShortFilenameFromStreamId(e.PublishLifecycleEventRequest.BuildEvent.StreamId)
		if err != nil {
			return
		}
		var buf []byte
		buf, err = proto.Marshal(d)
		if err != nil {
			return
		}
		err = w.appendToFile(filename, buf)
		if err != nil {
			return
		}

	case *wrapper.PublishBuildEventWrapper_PublishBuildToolEventStreamRequest:
		// fmt.Printf("%+v\n", e.PublishBuildToolEventStreamRequest.OrderedBuildEvent.StreamId)
		filename, err = generateShortFilenameFromStreamId(e.PublishBuildToolEventStreamRequest.OrderedBuildEvent.StreamId)
		if err != nil {
			return
		}
		var buf []byte
		buf, err = proto.Marshal(d)
		if err != nil {
			return
		}
		err = w.appendToFile(filename, buf)
		if err != nil {
			return
		}

	default:
		log.Printf("Ignoring unknown wrapped type %T\n", e)
		return
	}

	// fmt.Printf("Would write to %s\n", w.DataPath+"/"+filename)
	return
}
