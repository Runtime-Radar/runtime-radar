package autoscaler

// TargetConfig holds the scaling settings of one Deployment.
type TargetConfig struct {
	Name        string // used in logs, metrics and events
	Deployment  string
	Queue       string
	Min         int32
	Max         int32
	UpBacklog   int // scale up when the backlog stays above this for UpWindow samples
	UpWindow    int
	DownBacklog int // scale down when the backlog stays below this for DownWindow samples
	DownWindow  int
	Prober      Prober
}

// targetState counts consecutive samples above UpBacklog (high) and below DownBacklog (low).
type targetState struct {
	high int
	low  int
}

func (s *targetState) resetHigh() {
	s.high = 0
}

func (s *targetState) resetLow() {
	s.low = 0
}

func (s *targetState) reset() {
	s.high = 0
	s.low = 0
}
