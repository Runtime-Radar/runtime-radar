package config

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"slices"

	"github.com/runtime-radar/runtime-radar/lib/config"
)

// AutoscalerTarget holds the scaling settings of one Deployment.
type AutoscalerTarget struct {
	Name             string
	envPrefix        string // for error messages
	Enabled          bool
	Deployment       string
	Queue            string
	MinReplicas      int
	MaxReplicas      int
	ScaleUpBacklog   int // scale up when the backlog stays above this for ScaleUpWindow samples
	ScaleUpWindow    int
	ScaleDownBacklog int // scale down when the backlog stays below this for ScaleDownWindow samples
	ScaleDownWindow  int
}

const (
	eventProcessorEnvPrefix  = "AUTOSCALER_EVENT_PROCESSOR_"
	eventProcessorFlagPrefix = "autoscalerEventProcessor"
	historyAPIEnvPrefix      = "AUTOSCALER_HISTORY_API_"
	historyAPIFlagPrefix     = "autoscalerHistoryAPI"
)

func eventProcessorDefaults() AutoscalerTarget {
	return AutoscalerTarget{
		Name:             "event-processor",
		Enabled:          true,
		Deployment:       "event-processor",
		Queue:            "runtime_events",
		MinReplicas:      1,
		MaxReplicas:      20,
		ScaleUpBacklog:   1000,
		ScaleUpWindow:    4,
		ScaleDownBacklog: 10,
		ScaleDownWindow:  10,
	}
}

func historyAPIDefaults() AutoscalerTarget {
	return AutoscalerTarget{
		Name:             "history-api",
		Enabled:          true,
		Deployment:       "history-api",
		Queue:            "history_events",
		MinReplicas:      1,
		MaxReplicas:      20,
		ScaleUpBacklog:   20000,
		ScaleUpWindow:    4,
		ScaleDownBacklog: 200,
		ScaleDownWindow:  10,
	}
}

func registerAutoscalerTarget(fs *flag.FlagSet, envPrefix, flagPrefix string, defaults AutoscalerTarget) *AutoscalerTarget {
	t := &AutoscalerTarget{Name: defaults.Name, envPrefix: envPrefix}

	fs.BoolVar(&t.Enabled, flagPrefix+"Enabled", config.LookupEnvBool(envPrefix+"ENABLED", defaults.Enabled),
		fmt.Sprintf("Set to enable autoscaling of the %s deployment.", defaults.Name))
	fs.StringVar(&t.Deployment, flagPrefix+"Deployment", config.LookupEnvString(envPrefix+"DEPLOYMENT", defaults.Deployment),
		fmt.Sprintf("Name of the %s deployment to scale.", defaults.Name))
	fs.StringVar(&t.Queue, flagPrefix+"Queue", config.LookupEnvString(envPrefix+"QUEUE", defaults.Queue),
		fmt.Sprintf("Name of the RabbitMQ queue whose backlog drives scaling of %s.", defaults.Name))
	fs.IntVar(&t.MinReplicas, flagPrefix+"MinReplicas", config.LookupEnvInt(envPrefix+"MIN_REPLICAS", defaults.MinReplicas),
		fmt.Sprintf("Minimum number of %s replicas.", defaults.Name))
	fs.IntVar(&t.MaxReplicas, flagPrefix+"MaxReplicas", config.LookupEnvInt(envPrefix+"MAX_REPLICAS", defaults.MaxReplicas),
		fmt.Sprintf("Maximum number of %s replicas.", defaults.Name))
	fs.IntVar(&t.ScaleUpBacklog, flagPrefix+"ScaleUpBacklog", config.LookupEnvInt(envPrefix+"SCALE_UP_BACKLOG", defaults.ScaleUpBacklog),
		fmt.Sprintf("Number of ready messages above which %s scales up.", defaults.Name))
	fs.IntVar(&t.ScaleUpWindow, flagPrefix+"ScaleUpWindow", config.LookupEnvInt(envPrefix+"SCALE_UP_WINDOW", defaults.ScaleUpWindow),
		fmt.Sprintf("Number of consecutive samples above the scale up backlog required to scale %s up.", defaults.Name))
	fs.IntVar(&t.ScaleDownBacklog, flagPrefix+"ScaleDownBacklog", config.LookupEnvInt(envPrefix+"SCALE_DOWN_BACKLOG", defaults.ScaleDownBacklog),
		fmt.Sprintf("Number of ready messages below which %s scales down.", defaults.Name))
	fs.IntVar(&t.ScaleDownWindow, flagPrefix+"ScaleDownWindow", config.LookupEnvInt(envPrefix+"SCALE_DOWN_WINDOW", defaults.ScaleDownWindow),
		fmt.Sprintf("Number of consecutive samples below the scale down backlog required to scale %s down.", defaults.Name))

	return t
}

// AutoscalerTargets returns every target, disabled ones included.
func (c *Config) AutoscalerTargets() []AutoscalerTarget {
	return []AutoscalerTarget{*c.AutoscalerEventProcessor, *c.AutoscalerHistoryAPI}
}

// ValidateAutoscaler reports all invalid settings at once. Disabled autoscaler or targets are not checked.
func (c *Config) ValidateAutoscaler() error {
	if !c.AutoscalerEnabled {
		return nil
	}

	errs := []error{}
	if c.CSNamespace == "" {
		errs = append(errs, errors.New("CS_NAMESPACE must not be empty when AUTOSCALER_ENABLED is set"))
	}
	if c.AutoscalerInterval <= 0 {
		errs = append(errs, fmt.Errorf("AUTOSCALER_INTERVAL must be > 0, got %s", c.AutoscalerInterval))
	}
	if c.AutoscalerCooldown < 0 {
		errs = append(errs, fmt.Errorf("AUTOSCALER_COOLDOWN must be >= 0, got %s", c.AutoscalerCooldown))
	}

	targets := c.AutoscalerTargets()
	if !slices.ContainsFunc(targets, func(t AutoscalerTarget) bool { return t.Enabled }) {
		errs = append(errs, errors.New("at least one autoscaler target must be enabled when AUTOSCALER_ENABLED is set"))
	}

	for _, t := range targets {
		if !t.Enabled {
			continue
		}
		errs = append(errs, t.validate()...)
	}

	return errors.Join(errs...)
}

func (t *AutoscalerTarget) validate() []error {
	errs := []error{}

	if t.Deployment == "" {
		errs = append(errs, fmt.Errorf("%sDEPLOYMENT must not be empty", t.envPrefix))
	}
	if t.Queue == "" {
		errs = append(errs, fmt.Errorf("%sQUEUE must not be empty", t.envPrefix))
	}
	if t.MinReplicas < 1 {
		errs = append(errs, fmt.Errorf("%sMIN_REPLICAS must be >= 1, got %d", t.envPrefix, t.MinReplicas))
	}
	if t.MaxReplicas < t.MinReplicas {
		errs = append(errs, fmt.Errorf("%sMAX_REPLICAS must be >= MIN_REPLICAS, got %d < %d", t.envPrefix, t.MaxReplicas, t.MinReplicas))
	}
	// Replica bounds become int32 for Kubernetes.
	if t.MaxReplicas > math.MaxInt32 {
		errs = append(errs, fmt.Errorf("%sMAX_REPLICAS must be <= %d, got %d", t.envPrefix, math.MaxInt32, t.MaxReplicas))
	}
	if t.ScaleUpWindow < 1 {
		errs = append(errs, fmt.Errorf("%sSCALE_UP_WINDOW must be >= 1, got %d", t.envPrefix, t.ScaleUpWindow))
	}
	if t.ScaleDownWindow < 1 {
		errs = append(errs, fmt.Errorf("%sSCALE_DOWN_WINDOW must be >= 1, got %d", t.envPrefix, t.ScaleDownWindow))
	}
	if t.ScaleUpBacklog < 0 {
		errs = append(errs, fmt.Errorf("%sSCALE_UP_BACKLOG must be >= 0, got %d", t.envPrefix, t.ScaleUpBacklog))
	}
	// The rule tests backlog < SCALE_DOWN_BACKLOG, so a floor of 0 is unreachable and would leave
	// the target unable to ever scale down.
	if t.ScaleDownBacklog < 1 {
		errs = append(errs, fmt.Errorf("%sSCALE_DOWN_BACKLOG must be >= 1, got %d", t.envPrefix, t.ScaleDownBacklog))
	}
	if t.ScaleDownBacklog >= t.ScaleUpBacklog {
		errs = append(errs, fmt.Errorf("%sSCALE_DOWN_BACKLOG must be < SCALE_UP_BACKLOG, got %d >= %d", t.envPrefix, t.ScaleDownBacklog, t.ScaleUpBacklog))
	}

	return errs
}
