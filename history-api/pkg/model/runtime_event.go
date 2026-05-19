package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/event-processor/api"
	enf_model "github.com/runtime-radar/runtime-radar/policy-enforcer/pkg/model"
	"google.golang.org/protobuf/encoding/protojson"
	"gorm.io/gorm"
)

const (
	RuntimeEventTypeUndef             = "UNDEF"
	RuntimeEventTypeProcessExec       = "PROCESS_EXEC"
	RuntimeEventTypeProcessExit       = "PROCESS_EXIT"
	RuntimeEventTypeProcessKprobe     = "PROCESS_KPROBE"
	RuntimeEventTypeProcessTracepoint = "PROCESS_TRACEPOINT"
	RuntimeEventTypeProcessLoader     = "PROCESS_LOADER"
	RuntimeEventTypeProcessUprobe     = "PROCESS_UPROBE"
)

type (
	CapabilitiesTypes   []tetragon.CapabilitiesType
	RuntimeEventThreats []*api.Threat
	PodLabels           map[string]string
	// RuntimeEventJSON is type over tetragon.GetEventsResponse implementing sql.Scanner and driver.Valuer
	// so that it can be saved at database as JSON object in proper way.
	RuntimeEventJSON tetragon.GetEventsResponse
	DetectErrors     []*api.DetectError
)

// RuntimeEvent represents runtime event to be stored in database.
// It contains event itself in JSON format (SourceEvent) and set of fields that can be used for filtering.
// JSON fields Threats, ProcessPodPodLabels, ParentPodPodLabels are saved as NULLs if they're nil.
// Otherwise, they're saved as JSON objects, even having length of zero.
type RuntimeEvent struct {
	// Note: we cannot use Base model for ID and CreatedAt because type and indices annotations of uuid and time in PG and ClickHouse are differs
	ID uuid.UUID `gorm:"type:UUID"`
	// Timestamp when event had been saved to database
	CreatedAt time.Time `gorm:"type:DateTime64(9)"`
	// Timestamp when event had been registered by runtime monitor
	RegisteredAt time.Time `gorm:"type:DateTime64(9)"`

	// Common event attributes
	EventType         string `gorm:"type:Enum('UNDEF'=0,'PROCESS_EXEC'=1,'PROCESS_EXIT'=5,'PROCESS_KPROBE'=9,'PROCESS_TRACEPOINT'=10,'PROCESS_LOADER'=11,'PROCESS_UPROBE'=12)"`
	NodeName          string
	SourceEvent       *RuntimeEventJSON    `gorm:"type:String"` // Raw event from tetragon. Normally should not be nil.
	Threats           RuntimeEventThreats  `gorm:"type:Nullable(String)"`
	ThreatsDetectors  []string             `gorm:"type:Array(String)"`    // Identifiers of detectors from Threats
	ThreatsSeverities []enf_model.Severity `gorm:"type:Array(Integer)"`   // Each element of severity corresponds to a detector in the ThreatsDetectors array
	DetectErrors      DetectErrors         `gorm:"type:Nullable(String)"` // Errors returned by detectors contained in chain
	TetragonVersion   string               // Version of tetragon lib that produced event

	IsIncident       bool
	IncidentSeverity enf_model.Severity
	BlockBy          []string `gorm:"type:Array(String)"`
	NotifyBy         []string `gorm:"type:Array(String)"`

	// Common process attributes
	ProcessExecID    string // Here and below field names are in CamelCase for readability and consistency with tetragon proto-fields
	ProcessPid       uint32
	ProcessUID       uint32
	ProcessCwd       string
	ProcessBinary    string
	ProcessArguments string
	ProcessFlags     string
	// Here and below we use Nullable time, because zero value of time in go is "0000-00-00 00:00:00", but clickhouse support only datetime since "1900-01-01 00:00:00".
	ProcessStartTime                  *time.Time `gorm:"type:Nullable(DateTime64(9))"`
	ProcessAuid                       uint32
	ProcessPodNamespace               string
	ProcessPodName                    string
	ProcessPodContainerID             string
	ProcessPodContainerName           string
	ProcessPodContainerImageID        string
	ProcessPodContainerImageName      string
	ProcessPodContainerStartTime      *time.Time `gorm:"type:Nullable(DateTime64(9))"`
	ProcessPodContainerPid            uint32
	ProcessPodContainerMaybeExecProbe bool
	ProcessPodPodLabels               PodLabels `gorm:"type:Nullable(String)"`
	ProcessPodWorkload                string
	ProcessPodWorkloadKind            string
	ProcessDocker                     string
	ProcessParentExecID               string
	ProcessRefcnt                     uint32
	ProcessCapPermitted               CapabilitiesTypes `gorm:"type:Array(Enum('CAP_CHOWN'=0,'DAC_OVERRIDE'=1,'CAP_DAC_READ_SEARCH'=2,'CAP_FOWNER'=3,'CAP_FSETID'=4,'CAP_KILL'=5,'CAP_SETGID'=6,'CAP_SETUID'=7,'CAP_SETPCAP'=8,'CAP_LINUX_IMMUTABLE'=9,'CAP_NET_BIND_SERVICE'=10,'CAP_NET_BROADCAST'=11,'CAP_NET_ADMIN'=12,'CAP_NET_RAW'=13,'CAP_IPC_LOCK'=14,'CAP_IPC_OWNER'=15,'CAP_SYS_MODULE'=16,'CAP_SYS_RAWIO'=17,'CAP_SYS_CHROOT'=18,'CAP_SYS_PTRACE'=19,'CAP_SYS_PACCT'=20,'CAP_SYS_ADMIN'=21,'CAP_SYS_BOOT'=22,'CAP_SYS_NICE'=23,'CAP_SYS_RESOURCE'=24,'CAP_SYS_TIME'=25,'CAP_SYS_TTY_CONFIG'=26,'CAP_MKNOD'=27,'CAP_LEASE'=28,'CAP_AUDIT_WRITE'=29,'CAP_AUDIT_CONTROL'=30,'CAP_SETFCAP'=31,'CAP_MAC_OVERRIDE'=32,'CAP_MAC_ADMIN'=33,'CAP_SYSLOG'=34,'CAP_WAKE_ALARM'=35,'CAP_BLOCK_SUSPEND'=36,'CAP_AUDIT_READ'=37,'CAP_PERFMON'=38,'CAP_BPF'=39,'CAP_CHECKPOINT_RESTORE'= 40))"`
	ProcessCapEffective               CapabilitiesTypes `gorm:"type:Array(Enum('CAP_CHOWN'=0,'DAC_OVERRIDE'=1,'CAP_DAC_READ_SEARCH'=2,'CAP_FOWNER'=3,'CAP_FSETID'=4,'CAP_KILL'=5,'CAP_SETGID'=6,'CAP_SETUID'=7,'CAP_SETPCAP'=8,'CAP_LINUX_IMMUTABLE'=9,'CAP_NET_BIND_SERVICE'=10,'CAP_NET_BROADCAST'=11,'CAP_NET_ADMIN'=12,'CAP_NET_RAW'=13,'CAP_IPC_LOCK'=14,'CAP_IPC_OWNER'=15,'CAP_SYS_MODULE'=16,'CAP_SYS_RAWIO'=17,'CAP_SYS_CHROOT'=18,'CAP_SYS_PTRACE'=19,'CAP_SYS_PACCT'=20,'CAP_SYS_ADMIN'=21,'CAP_SYS_BOOT'=22,'CAP_SYS_NICE'=23,'CAP_SYS_RESOURCE'=24,'CAP_SYS_TIME'=25,'CAP_SYS_TTY_CONFIG'=26,'CAP_MKNOD'=27,'CAP_LEASE'=28,'CAP_AUDIT_WRITE'=29,'CAP_AUDIT_CONTROL'=30,'CAP_SETFCAP'=31,'CAP_MAC_OVERRIDE'=32,'CAP_MAC_ADMIN'=33,'CAP_SYSLOG'=34,'CAP_WAKE_ALARM'=35,'CAP_BLOCK_SUSPEND'=36,'CAP_AUDIT_READ'=37,'CAP_PERFMON'=38,'CAP_BPF'=39,'CAP_CHECKPOINT_RESTORE'= 40))"`
	ProcessCapInheritable             CapabilitiesTypes `gorm:"type:Array(Enum('CAP_CHOWN'=0,'DAC_OVERRIDE'=1,'CAP_DAC_READ_SEARCH'=2,'CAP_FOWNER'=3,'CAP_FSETID'=4,'CAP_KILL'=5,'CAP_SETGID'=6,'CAP_SETUID'=7,'CAP_SETPCAP'=8,'CAP_LINUX_IMMUTABLE'=9,'CAP_NET_BIND_SERVICE'=10,'CAP_NET_BROADCAST'=11,'CAP_NET_ADMIN'=12,'CAP_NET_RAW'=13,'CAP_IPC_LOCK'=14,'CAP_IPC_OWNER'=15,'CAP_SYS_MODULE'=16,'CAP_SYS_RAWIO'=17,'CAP_SYS_CHROOT'=18,'CAP_SYS_PTRACE'=19,'CAP_SYS_PACCT'=20,'CAP_SYS_ADMIN'=21,'CAP_SYS_BOOT'=22,'CAP_SYS_NICE'=23,'CAP_SYS_RESOURCE'=24,'CAP_SYS_TIME'=25,'CAP_SYS_TTY_CONFIG'=26,'CAP_MKNOD'=27,'CAP_LEASE'=28,'CAP_AUDIT_WRITE'=29,'CAP_AUDIT_CONTROL'=30,'CAP_SETFCAP'=31,'CAP_MAC_OVERRIDE'=32,'CAP_MAC_ADMIN'=33,'CAP_SYSLOG'=34,'CAP_WAKE_ALARM'=35,'CAP_BLOCK_SUSPEND'=36,'CAP_AUDIT_READ'=37,'CAP_PERFMON'=38,'CAP_BPF'=39,'CAP_CHECKPOINT_RESTORE'= 40))"`
	ProcessNsUtsInum                  uint32
	ProcessNsUtsIsHost                bool
	ProcessNsIpcInum                  uint32
	ProcessNsIpcIsHost                bool
	ProcessNsMntInum                  uint32
	ProcessNsMntIsHost                bool
	ProcessNsPidInum                  uint32
	ProcessNsPidIsHost                bool
	ProcessNsPidForChildrenInum       uint32
	ProcessNsPidForChildrenIsHost     bool
	ProcessNsNetInum                  uint32
	ProcessNsNetIsHost                bool
	ProcessNsTimeInum                 uint32
	ProcessNsTimeIsHost               bool
	ProcessNsTimeForChildrenInum      uint32
	ProcessNsTimeForChildrenIsHost    bool
	ProcessNsCgroupInum               uint32
	ProcessNsCgroupIsHost             bool
	ProcessNsUserInum                 uint32
	ProcessNsUserIsHost               bool
	ProcessTid                        uint32

	// Process's parent attributes
	ParentExecID                     string
	ParentPid                        uint32
	ParentUID                        uint32
	ParentCwd                        string
	ParentBinary                     string
	ParentArguments                  string
	ParentFlags                      string
	ParentStartTime                  *time.Time `gorm:"type:Nullable(DateTime64(9))"`
	ParentAuid                       uint32
	ParentPodNamespace               string
	ParentPodName                    string
	ParentPodContainerID             string
	ParentPodContainerName           string
	ParentPodContainerImageID        string
	ParentPodContainerImageName      string
	ParentPodContainerStartTime      *time.Time `gorm:"type:Nullable(DateTime64(9))"`
	ParentPodContainerPid            uint32
	ParentPodContainerMaybeExecProbe bool
	ParentPodPodLabels               PodLabels `gorm:"type:Nullable(String)"`
	ParentPodWorkload                string
	ParentPodWorkloadKind            string
	ParentDocker                     string
	ParentParentExecID               string
	ParentRefcnt                     uint32
	ParentCapPermitted               CapabilitiesTypes `gorm:"type:Array(Enum('CAP_CHOWN'=0,'DAC_OVERRIDE'=1,'CAP_DAC_READ_SEARCH'=2,'CAP_FOWNER'=3,'CAP_FSETID'=4,'CAP_KILL'=5,'CAP_SETGID'=6,'CAP_SETUID'=7,'CAP_SETPCAP'=8,'CAP_LINUX_IMMUTABLE'=9,'CAP_NET_BIND_SERVICE'=10,'CAP_NET_BROADCAST'=11,'CAP_NET_ADMIN'=12,'CAP_NET_RAW'=13,'CAP_IPC_LOCK'=14,'CAP_IPC_OWNER'=15,'CAP_SYS_MODULE'=16,'CAP_SYS_RAWIO'=17,'CAP_SYS_CHROOT'=18,'CAP_SYS_PTRACE'=19,'CAP_SYS_PACCT'=20,'CAP_SYS_ADMIN'=21,'CAP_SYS_BOOT'=22,'CAP_SYS_NICE'=23,'CAP_SYS_RESOURCE'=24,'CAP_SYS_TIME'=25,'CAP_SYS_TTY_CONFIG'=26,'CAP_MKNOD'=27,'CAP_LEASE'=28,'CAP_AUDIT_WRITE'=29,'CAP_AUDIT_CONTROL'=30,'CAP_SETFCAP'=31,'CAP_MAC_OVERRIDE'=32,'CAP_MAC_ADMIN'=33,'CAP_SYSLOG'=34,'CAP_WAKE_ALARM'=35,'CAP_BLOCK_SUSPEND'=36,'CAP_AUDIT_READ'=37,'CAP_PERFMON'=38,'CAP_BPF'=39,'CAP_CHECKPOINT_RESTORE'= 40))"`
	ParentCapEffective               CapabilitiesTypes `gorm:"type:Array(Enum('CAP_CHOWN'=0,'DAC_OVERRIDE'=1,'CAP_DAC_READ_SEARCH'=2,'CAP_FOWNER'=3,'CAP_FSETID'=4,'CAP_KILL'=5,'CAP_SETGID'=6,'CAP_SETUID'=7,'CAP_SETPCAP'=8,'CAP_LINUX_IMMUTABLE'=9,'CAP_NET_BIND_SERVICE'=10,'CAP_NET_BROADCAST'=11,'CAP_NET_ADMIN'=12,'CAP_NET_RAW'=13,'CAP_IPC_LOCK'=14,'CAP_IPC_OWNER'=15,'CAP_SYS_MODULE'=16,'CAP_SYS_RAWIO'=17,'CAP_SYS_CHROOT'=18,'CAP_SYS_PTRACE'=19,'CAP_SYS_PACCT'=20,'CAP_SYS_ADMIN'=21,'CAP_SYS_BOOT'=22,'CAP_SYS_NICE'=23,'CAP_SYS_RESOURCE'=24,'CAP_SYS_TIME'=25,'CAP_SYS_TTY_CONFIG'=26,'CAP_MKNOD'=27,'CAP_LEASE'=28,'CAP_AUDIT_WRITE'=29,'CAP_AUDIT_CONTROL'=30,'CAP_SETFCAP'=31,'CAP_MAC_OVERRIDE'=32,'CAP_MAC_ADMIN'=33,'CAP_SYSLOG'=34,'CAP_WAKE_ALARM'=35,'CAP_BLOCK_SUSPEND'=36,'CAP_AUDIT_READ'=37,'CAP_PERFMON'=38,'CAP_BPF'=39,'CAP_CHECKPOINT_RESTORE'= 40))"`
	ParentCapInheritable             CapabilitiesTypes `gorm:"type:Array(Enum('CAP_CHOWN'=0,'DAC_OVERRIDE'=1,'CAP_DAC_READ_SEARCH'=2,'CAP_FOWNER'=3,'CAP_FSETID'=4,'CAP_KILL'=5,'CAP_SETGID'=6,'CAP_SETUID'=7,'CAP_SETPCAP'=8,'CAP_LINUX_IMMUTABLE'=9,'CAP_NET_BIND_SERVICE'=10,'CAP_NET_BROADCAST'=11,'CAP_NET_ADMIN'=12,'CAP_NET_RAW'=13,'CAP_IPC_LOCK'=14,'CAP_IPC_OWNER'=15,'CAP_SYS_MODULE'=16,'CAP_SYS_RAWIO'=17,'CAP_SYS_CHROOT'=18,'CAP_SYS_PTRACE'=19,'CAP_SYS_PACCT'=20,'CAP_SYS_ADMIN'=21,'CAP_SYS_BOOT'=22,'CAP_SYS_NICE'=23,'CAP_SYS_RESOURCE'=24,'CAP_SYS_TIME'=25,'CAP_SYS_TTY_CONFIG'=26,'CAP_MKNOD'=27,'CAP_LEASE'=28,'CAP_AUDIT_WRITE'=29,'CAP_AUDIT_CONTROL'=30,'CAP_SETFCAP'=31,'CAP_MAC_OVERRIDE'=32,'CAP_MAC_ADMIN'=33,'CAP_SYSLOG'=34,'CAP_WAKE_ALARM'=35,'CAP_BLOCK_SUSPEND'=36,'CAP_AUDIT_READ'=37,'CAP_PERFMON'=38,'CAP_BPF'=39,'CAP_CHECKPOINT_RESTORE'= 40))"`
	ParentNsUtsInum                  uint32
	ParentNsUtsIsHost                bool
	ParentNsIpcInum                  uint32
	ParentNsIpcIsHost                bool
	ParentNsMntInum                  uint32
	ParentNsMntIsHost                bool
	ParentNsPidInum                  uint32
	ParentNsPidIsHost                bool
	ParentNsPidForChildrenInum       uint32
	ParentNsPidForChildrenIsHost     bool
	ParentNsNetInum                  uint32
	ParentNsNetIsHost                bool
	ParentNsTimeInum                 uint32
	ParentNsTimeIsHost               bool
	ParentNsTimeForChildrenInum      uint32
	ParentNsTimeForChildrenIsHost    bool
	ParentNsCgroupInum               uint32
	ParentNsCgroupIsHost             bool
	ParentNsUserInum                 uint32
	ParentNsUserIsHost               bool
	ParentTid                        uint32

	PolicyName string

	// Extra exit attributes
	ExitSignal string
	ExitStatus uint32
	ExitTime   *time.Time `gorm:"type:Nullable(DateTime64(9))"`

	// Extra kprobe attributes
	KprobeFunctionName string
	KprobeAction       string `gorm:"type:Enum('KPROBE_ACTION_UNKNOWN'=0,'KPROBE_ACTION_POST'=1,'KPROBE_ACTION_FOLLOWFD'=2,'KPROBE_ACTION_SIGKILL'=3,'KPROBE_ACTION_UNFOLLOWFD'=4,'KPROBE_ACTION_OVERRIDE'=5, 'KPROBE_ACTION_COPYFD'=6,'KPROBE_ACTION_GETURL'=7,'KPROBE_ACTION_DNSLOOKUP'=8,'KPROBE_ACTION_NOPOST'=9,'KPROBE_ACTION_SIGNAL'=10)"`

	// Extra tracepoint attributes
	TracepointSubsys string
	TracepointEvent  string

	// Extra loader attributes
	LoaderPath    string
	LoaderBuildid []byte

	// Extra uprobe attributes
	UprobePath   string
	UprobeSymbol string
}

// BeforeCreate callback sets default values for enums and ID as newly generated UUID. If ID was set already, just go ahead and do nothing.
// Note: we cannot use Base model for this purpose because type and indices annotations of uuid type in PG and ClickHouse are differs
func (re *RuntimeEvent) BeforeCreate(*gorm.DB) error {
	if re.ID == uuid.Nil {
		re.ID = uuid.New()
	}

	// We need assign zero values for Enums to avoid key error in Clickhouse
	if re.KprobeAction == "" {
		re.KprobeAction = "KPROBE_ACTION_UNKNOWN"
	}

	return nil
}

func (ct *CapabilitiesTypes) Scan(src any) error {
	if ct == nil {
		return errors.New("unpacking failed: CapabilitiesTypes is nil")
	}

	strings, ok := src.([]string)
	if !ok {
		return fmt.Errorf("expected []string, got %T", src)
	}

	*ct = make(CapabilitiesTypes, 0, len(strings))
	for _, str := range strings {
		enumValue, ok := tetragon.CapabilitiesType_value[str]
		if !ok {
			return fmt.Errorf("unknown key %s for enum tetragon.CapabilitiesType", str)
		}
		*ct = append(*ct, tetragon.CapabilitiesType(enumValue))
	}

	return nil
}

func (t *RuntimeEventThreats) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}

	return json.Unmarshal([]byte(s), t)
}

func (t RuntimeEventThreats) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}

	marshalled, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("can't marshal json: %w", err)
	}

	return string(marshalled), nil
}

func (pl PodLabels) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}

	return json.Unmarshal([]byte(s), &pl)
}

func (pl PodLabels) Value() (driver.Value, error) {
	if pl == nil {
		return nil, nil
	}

	marshalled, err := json.Marshal(pl)
	if err != nil {
		return nil, fmt.Errorf("can't marshal json: %w", err)
	}

	return string(marshalled), nil // labels are stored as nullable string in clickhouse
}

func (r *RuntimeEventJSON) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}

	resp := (*tetragon.GetEventsResponse)(r)

	return protojson.Unmarshal([]byte(s), resp)
}

func (r *RuntimeEventJSON) Value() (driver.Value, error) {
	resp := (*tetragon.GetEventsResponse)(r)

	b, err := protojson.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (d *DetectErrors) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", src)
	}

	return json.Unmarshal([]byte(s), d)
}

func (d DetectErrors) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}

	marshalled, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("can't marshal json: %w", err)
	}

	return string(marshalled), nil
}
