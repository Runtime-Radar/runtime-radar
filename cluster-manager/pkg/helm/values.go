package helm

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"sigs.k8s.io/yaml"
)

const (
	DefaultUser              = "runtime-radar"
	DefaultNamespace         = "runtime-radar"
	DefaultRetentionInterval = time.Hour * 8460
	DefaultRegistry          = "ghcr.io/runtime-radar"
)

var (
	errUnsupported = errors.New("unsupported field type")
)

// Values represents helm values returned as a complete install command or yaml.
// `sigs.k8s.io/yaml` library ignores `yaml` tag and uses only `json`. So only
// `json` tag is used here.
// `struct` fields are followed by `json:"...,omitzero"` because `omitempty`
// does not check for struct's zero value.
type Values struct {
	Global struct {
		CSVersion      string `json:"csVersion,omitempty"`
		IsChildCluster bool   `json:"isChildCluster"`
		OwnCSURL       string `json:"ownCsUrl,omitempty"`
		CentralCSURL   string `json:"centralCsUrl,omitempty"`

		ImageRegistry   string `json:"imageRegistry,omitempty"`
		ImageShortNames bool   `json:"imageShortNames,omitempty"`

		Keys struct {
			Encryption            string `json:"encryption,omitempty"`
			Token                 string `json:"token,omitempty"`
			PublicAccessTokenSalt string `json:"publicAccessTokenSalt,omitempty"`
		} `json:"keys,omitzero"`

		Postgresql TLSGlobal `json:"postgresql"`
		Redis      TLSGlobal `json:"redis"`
		Clickhouse TLSGlobal `json:"clickhouse"`
		Grafana    TLSGlobal `json:"grafana,omitzero"`
		Loki       TLSGlobal `json:"loki,omitzero"`
	} `json:"global,omitzero"`

	TLS struct {
		Verify  bool   `json:"verify"`
		CertCA  string `json:"certCA,omitempty"`
		Cert    string `json:"cert,omitempty"`
		CertKey string `json:"certKey,omitempty"`
	} `json:"tls,omitzero"`

	AuthAPI struct {
		Administrator struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
		} `json:"administrator,omitzero"`
	} `json:"auth-center,omitzero"`

	ImagePullSecret struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"imagePullSecret,omitzero"`

	HistoryAPI struct {
		RetentionInterval string `json:"retentionInterval,omitempty"`
	} `json:"history-api,omitzero"`

	Postgresql struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		Auth struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
			Database string `json:"database,omitempty"`
		} `json:"auth,omitzero"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`

		TLS struct {
			CertCA  string `json:"certCA,omitempty"`
			Cert    string `json:"cert,omitempty"`
			CertKey string `json:"certKey,omitempty"`
		} `json:"tls,omitzero"`
	} `json:"postgresql"`

	Redis struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		Auth struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
		} `json:"auth,omitzero"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`

		TLS struct {
			CertCA  string `json:"certCA,omitempty"`
			Cert    string `json:"cert,omitempty"`
			CertKey string `json:"certKey,omitempty"`
		} `json:"tls,omitzero"`
	} `json:"redis"`

	Rabbitmq struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		Auth struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
		} `json:"auth,omitzero"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`
	} `json:"rabbitmq"`

	Clickhouse struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		Auth struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
			Database string `json:"database,omitempty"`
		} `json:"auth,omitzero"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`

		TLS struct {
			CertCA  string `json:"certCA,omitempty"`
			Cert    string `json:"cert,omitempty"`
			CertKey string `json:"certKey,omitempty"`
		} `json:"tls,omitzero"`
	} `json:"clickhouse"`

	Prometheus struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`
	} `json:"prometheus,omitzero"`

	Grafana struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		Auth struct {
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
		} `json:"auth,omitzero"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`

		TLS struct {
			CertCA  string `json:"certCA,omitempty"`
			Cert    string `json:"cert,omitempty"`
			CertKey string `json:"certKey,omitempty"`
		} `json:"tls,omitzero"`
	} `json:"grafana,omitzero"`

	Metrics struct {
		Enabled bool `json:"enabled"`
	} `json:"metrics"`

	Loki struct {
		Deploy       bool   `json:"deploy"`
		ExternalHost string `json:"externalHost,omitempty"`

		TenantID string `json:"tenant_id,omitempty"`

		Persistence struct {
			Enabled      bool   `json:"enabled"`
			StorageClass string `json:"storageClass,omitempty"`
		} `json:"persistence"`

		TLS struct {
			CertCA  string `json:"certCA,omitempty"`
			Cert    string `json:"cert,omitempty"`
			CertKey string `json:"certKey,omitempty"`
		} `json:"tls,omitzero"`
	} `json:"loki,omitzero"`

	ReverseProxy struct {
		Ingress struct {
			Enabled  bool   `json:"enabled,omitempty"`
			Class    string `json:"class,omitempty"`
			Hostname string `json:"hostname,omitempty"`
			TLS      struct {
				CertCA         string `json:"certCA,omitempty"`
				Cert           string `json:"cert,omitempty"`
				CertKey        string `json:"certKey,omitempty"`
				ExistingSecret string `json:"existingSecret,omitempty"`
			} `json:"tls,omitzero"`
		} `json:"ingress,omitzero"`

		Service struct {
			Type      string `json:"type,omitempty"`
			NodePorts struct {
				HTTP string `json:"http,omitempty"`
			} `json:"nodePorts,omitzero"`
		} `json:"service,omitzero"`
	} `json:"reverse-proxy,omitzero"`

	Notifier struct {
		Env []Env `json:"env,omitempty"`
	} `json:"notifier,omitzero"`

	CSManager struct {
		RegistrationToken string `json:"registrationToken,omitempty"`
	} `json:"cs-manager,omitzero"`
}

type Env struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TLSGlobal struct {
	TLS struct {
		Enabled bool `json:"enabled"`
		Verify  bool `json:"verify"`
	} `json:"tls"`
}

// buildHelmArgs recursively converts a value to Helm command-line arguments.
// It traverses the value and generates the appropriate --set or --set-string arguments
// based on the value's kind.
//
// Each value is processed according to its kind:
//   - Strings use --set-string
//   - Bool, numeric types use --set
//   - Arrays/Slices are expanded into per-element args using the `name[i]` form
//   - Structs are processed recursively over their fields
//
// Struct fields with `json:"-"` are skipped.
// Struct fields with `json:",omitempty"` or `json:",omitzero"` are skipped if they
// contain zero values; this is communicated through the `hasOmit` argument when
// recursing.
//
// Returns an error if JSON marshaling fails for array/slice fields.
func buildHelmArgs(v any, prefix string, hasOmit bool) ([]string, error) {
	var res []string

	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// skip field on `omitempty` and `omitzero`
	// theoretically `field.IsZero()` can possibly panic
	// but currently it should never happen
	if hasOmit && val.IsZero() {
		return nil, nil
	}

	switch val.Kind() {
	case reflect.String:
		res = append(res, fmt.Sprintf("--set-string '%s=%s'", prefix, val.String()))

	case reflect.Bool:
		res = append(res, fmt.Sprintf("--set '%s=%t'", prefix, val.Bool()))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		res = append(res, fmt.Sprintf("--set '%s=%d'", prefix, val.Int()))

	case reflect.Float32, reflect.Float64:
		res = append(res, fmt.Sprintf("--set '%s=%f'", prefix, val.Float()))

	case reflect.Array, reflect.Slice:
		if hasOmit && val.Len() == 0 {
			return nil, nil
		}
		for i := range val.Len() {
			nestedRes, err := buildHelmArgs(val.Index(i).Interface(), fmt.Sprintf("%s[%d]", prefix, i), false)
			if err != nil {
				return nil, err
			}
			res = append(res, nestedRes...)
		}
	case reflect.Struct:
		t := val.Type()

		for i := range val.NumField() {
			field := val.Field(i)
			fieldType := t.Field(i)

			tag := fieldType.Tag.Get("json")
			if tag == "-" {
				continue
			}

			parts := strings.Split(tag, ",")
			fieldName := parts[0]
			if fieldName == "" {
				fieldName = fieldType.Name
			}

			hasOmit := (slices.Index(parts, "omitempty") != -1) || (slices.Index(parts, "omitzero") != -1)

			currentPrefix := fieldName
			if prefix != "" {
				currentPrefix = prefix + "." + fieldName
			}

			nestedRes, err := buildHelmArgs(field.Interface(), currentPrefix, hasOmit)
			if err != nil {
				return nil, err
			}
			res = append(res, nestedRes...)
		}
	default:
		return nil, errUnsupported
	}

	return res, nil
}

func (v *Values) ToYAML() (string, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return "", nil
	}

	return string(b), nil
}

func (v *Values) ToHelmArgs() ([]string, error) {
	return buildHelmArgs(v, "", false)
}
