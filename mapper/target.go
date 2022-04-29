package mapper

import (
	"fmt"

	"github.com/hj1980/bazel-proxy/protobuf/types/known/build_event_stream"
	build "google.golang.org/genproto/googleapis/devtools/build/v1"
)

type Timing struct {
	startSeconds int64
	stopSeconds  int64
}

func Start(s int64) (newT *Timing) {
	newT = &Timing{
		startSeconds: s,
	}
	return
}

func (t *Timing) Stop(s int64) {
	t.stopSeconds = s
}

func (t *Timing) Duration() (d int64) {
	d = t.stopSeconds - t.startSeconds
	return
}

type TargetMapper struct {
	ids map[string]*Timing
}

func NewTargetMapper() (tm *TargetMapper) {
	tm = &TargetMapper{}
	tm.ids = make(map[string]*Timing)
	return
}

// Print to console for now
func (m *TargetMapper) Report() {
	for k, v := range m.ids {
		fmt.Printf("%s\t%d\n", k, v.Duration())
	}
}

func (m *TargetMapper) HandleBazelBuildEvent(be *build.BuildEvent, bbe *build_event_stream.BuildEvent) {

	if bbe.LastMessage {
		m.Report()
		return
	}

	switch id := bbe.Id.Id.(type) {

	case *build_event_stream.BuildEventId_TargetConfigured:
		sid := id.TargetConfigured.Label
		m.ids[sid] = Start(be.EventTime.Seconds)

	case *build_event_stream.BuildEventId_TargetCompleted:
		sid := id.TargetCompleted.Label
		if timing, exists := m.ids[sid]; exists {
			timing.Stop(be.EventTime.Seconds)
		}

	}

}
