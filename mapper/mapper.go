package mapper

import (
	"github.com/hj1980/bazel-proxy/protobuf/types/known/build_event_stream"
	build "google.golang.org/genproto/googleapis/devtools/build/v1"
)

type BazelBuildEventMapper interface {
	HandleBazelBuildEvent(be *build.BuildEvent, bbe *build_event_stream.BuildEvent)
}
