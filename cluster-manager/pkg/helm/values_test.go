package helm

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestBuildHelmArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       any
		prefix      string
		expected    []string
		expectedErr error
	}{
		{
			name: "basic struct with different types",
			input: struct {
				Name    string  `json:"name"`
				Enabled bool    `json:"enabled"`
				Count   int     `json:"count"`
				Price   float64 `json:"price"`
				NoTag   string
				Skipped string `json:"-"`
				Minus   string `json:"-,"`
			}{
				Name:    "test",
				Enabled: true,
				Count:   5,
				Price:   10.5,
				NoTag:   "has value",
				Skipped: "should be skipped",
				Minus:   "should be minus",
			},
			prefix: "",
			expected: []string{
				"--set-string 'name=test'",
				"--set 'enabled=true'",
				"--set 'count=5'",
				"--set 'price=10.500000'",
				"--set-string 'NoTag=has value'",
				"--set-string '-=should be minus'",
			},
			expectedErr: nil,
		},
		{
			name: "omitempty fields",
			input: struct {
				Name       string `json:"name,omitempty"`
				EmptyName  string `json:"emptyName,omitempty"`
				Field      string `json:",omitempty"`
				EmptyField string `json:",omitempty"`
				ZeroInt    int    `json:"zeroInt,omitempty"`
				EmptySlice []int  `json:"emptySlice,omitempty"`
				Nested     struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				} `json:"nested"`
				NestedEmpty struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				} `json:"nestedEmpty,omitzero"`
				NestedZero struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				} `json:"nestedZero,omitzero"`
			}{
				Name:       "test",
				EmptyName:  "",
				Field:      "test",
				EmptyField: "",
				ZeroInt:    0,
				EmptySlice: []int{},
				Nested: struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				}{
					Field1: "has value",
					Field2: 42,
				},
				NestedEmpty: struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				}{
					Field1: "",
					Field2: 0,
				},
				NestedZero: struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				}{
					Field1: "",
					Field2: 0,
				},
			},
			prefix: "",
			expected: []string{
				"--set-string 'name=test'",
				"--set-string 'Field=test'",
				"--set-string 'nested.field1=has value'",
				"--set 'nested.field2=42'",
			},
			expectedErr: nil,
		},
		{
			name: "nested struct",
			input: struct {
				Name   string `json:"name"`
				Nested struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				} `json:"nested"`
			}{
				Name: "parent",
				Nested: struct {
					Field1 string `json:"field1"`
					Field2 int    `json:"field2"`
				}{
					Field1: "child",
					Field2: 42,
				},
			},
			prefix: "",
			expected: []string{
				"--set-string 'name=parent'",
				"--set-string 'nested.field1=child'",
				"--set 'nested.field2=42'",
			},
			expectedErr: nil,
		},
		{
			name: "with prefix",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "test",
				Age:  30,
			},
			prefix: "person",
			expected: []string{
				"--set-string 'person.name=test'",
				"--set 'person.age=30'",
			},
			expectedErr: nil,
		},
		{
			name: "with slice",
			input: struct {
				Names []string `json:"names"`
			}{
				Names: []string{"alice", "bob", "charlie"},
			},
			prefix: "",
			expected: []string{
				"--set-string 'names[0]=alice'",
				"--set-string 'names[1]=bob'",
				"--set-string 'names[2]=charlie'",
			},
			expectedErr: nil,
		},
		{
			name: "with slice of structs",
			input: struct {
				Env []Env `json:"env"`
			}{
				Env: []Env{
					{
						Name:  "ENV1",
						Value: "value1",
					},
					{
						Name:  "ENV2",
						Value: "value2",
					},
				},
			},
			prefix: "",
			expected: []string{
				"--set-string 'env[0].name=ENV1'",
				"--set-string 'env[0].value=value1'",
				"--set-string 'env[1].name=ENV2'",
				"--set-string 'env[1].value=value2'",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args, err := buildHelmArgs(tt.input, tt.prefix, false)

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error '%v', got '%v'", tt.expectedErr, err)
			}

			if diff := cmp.Diff(args, tt.expected,
				cmpopts.SortSlices(func(a, b string) bool {
					return a < b // particular way of sorting does not matter
				}),
			); diff != "" {
				t.Fatalf("expected args != actual: %s", diff)
			}
		})
	}
}

func generateValues() Values {
	values := Values{}

	// Global settings
	values.Global.CSVersion = "1.0.0"
	values.Global.IsChildCluster = true
	values.Global.OwnCSURL = "https://cs.local"
	values.Global.CentralCSURL = "https://central-cs.local"
	values.Global.ImageRegistry = "registry.example.com"
	values.Global.ImageShortNames = false
	values.Global.Keys.Encryption = "encryption-key"
	values.Global.Keys.Token = "token-key"

	values.TLS.CertCA = "ca-cert-data"
	values.TLS.Cert = "cert-data"
	values.TLS.CertKey = "cert-key-data"

	// Image pull secret
	values.ImagePullSecret.Username = "registry-user"
	values.ImagePullSecret.Password = "registry-password"

	// PostgreSQL
	values.Postgresql.Deploy = true
	values.Postgresql.Auth.Username = "postgres"
	values.Postgresql.Auth.Password = "postgres-password"
	values.Postgresql.Persistence.Enabled = true
	values.Postgresql.Persistence.StorageClass = "standard-1"
	values.Global.Postgresql.TLS.Enabled = true
	values.Global.Postgresql.TLS.Verify = true

	// Redis
	values.Redis.Deploy = true
	values.Redis.Auth.Username = "redis"
	values.Redis.Auth.Password = "redis-password"

	// RabbitMQ
	values.Rabbitmq.Deploy = true
	values.Rabbitmq.Auth.Username = "rabbitmq"
	values.Rabbitmq.Auth.Password = "rabbitmq-password"
	values.Rabbitmq.Persistence.Enabled = true
	values.Rabbitmq.Persistence.StorageClass = "standard-2"

	// Clickhouse
	values.Clickhouse.Deploy = false
	values.Clickhouse.ExternalHost = "clickhouse.local"

	// Reverse proxy
	values.ReverseProxy.Ingress.Enabled = true
	values.ReverseProxy.Ingress.Class = "nginx"
	values.ReverseProxy.Ingress.Hostname = "cs.example.com"
	values.ReverseProxy.Ingress.TLS.ExistingSecret = "cs-tls"
	values.ReverseProxy.Service.Type = "ClusterIP"

	// Notifier
	values.Notifier.Env = []Env{
		{Name: "HTTP_PROXY", Value: "http://proxy.local"},
		{Name: "HTTPS_PROXY", Value: "https://proxy.local"},
	}

	// CS Manager
	values.CSManager.RegistrationToken = "registration-token"

	// AuthAPI
	values.AuthAPI.Administrator.Username = "user"
	values.AuthAPI.Administrator.Password = "pass"

	return values
}

func TestValuesToHelmArgs(t *testing.T) {
	t.Parallel()

	values := generateValues()

	// Convert to helm args
	args, err := values.ToHelmArgs()
	if err != nil {
		t.Fatalf("failed to convert Values to helm args: %v", err)
	}

	expectedArgs := []string{
		"--set-string 'global.csVersion=1.0.0'",
		"--set 'global.isChildCluster=true'",
		"--set-string 'global.ownCsUrl=https://cs.local'",
		"--set-string 'global.centralCsUrl=https://central-cs.local'",
		"--set-string 'global.imageRegistry=registry.example.com'",
		"--set-string 'global.keys.encryption=encryption-key'",
		"--set-string 'global.keys.token=token-key'",
		"--set 'tls.verify=false'",
		"--set-string 'tls.certCA=ca-cert-data'",
		"--set-string 'tls.cert=cert-data'",
		"--set-string 'tls.certKey=cert-key-data'",
		"--set-string 'imagePullSecret.username=registry-user'",
		"--set-string 'imagePullSecret.password=registry-password'",
		"--set 'postgresql.deploy=true'",
		"--set-string 'postgresql.auth.username=postgres'",
		"--set-string 'postgresql.auth.password=postgres-password'",
		"--set 'postgresql.persistence.enabled=true'",
		"--set 'global.postgresql.tls.enabled=true'",
		"--set 'global.postgresql.tls.verify=true'",
		"--set-string 'postgresql.persistence.storageClass=standard-1'",
		"--set 'redis.deploy=true'",
		"--set-string 'redis.auth.username=redis'",
		"--set-string 'redis.auth.password=redis-password'",
		"--set 'redis.persistence.enabled=false'",
		"--set 'global.redis.tls.enabled=false'",
		"--set 'global.redis.tls.verify=false'",
		"--set 'metrics.enabled=false'",
		"--set 'rabbitmq.deploy=true'",
		"--set-string 'rabbitmq.auth.username=rabbitmq'",
		"--set-string 'rabbitmq.auth.password=rabbitmq-password'",
		"--set 'rabbitmq.persistence.enabled=true'",
		"--set-string 'rabbitmq.persistence.storageClass=standard-2'",
		"--set 'clickhouse.deploy=false'",
		"--set-string 'clickhouse.externalHost=clickhouse.local'",
		"--set 'clickhouse.persistence.enabled=false'",
		"--set 'global.clickhouse.tls.enabled=false'",
		"--set 'global.clickhouse.tls.verify=false'",
		"--set 'reverse-proxy.ingress.enabled=true'",
		"--set-string 'reverse-proxy.ingress.class=nginx'",
		"--set-string 'reverse-proxy.ingress.hostname=cs.example.com'",
		"--set-string 'reverse-proxy.ingress.tls.existingSecret=cs-tls'",
		"--set-string 'reverse-proxy.service.type=ClusterIP'",
		"--set-string 'notifier.env[0].name=HTTP_PROXY'",
		"--set-string 'notifier.env[0].value=http://proxy.local'",
		"--set-string 'notifier.env[1].name=HTTPS_PROXY'",
		"--set-string 'notifier.env[1].value=https://proxy.local'",
		"--set-string 'cs-manager.registrationToken=registration-token'",
		"--set-string 'auth-center.administrator.username=user'",
		"--set-string 'auth-center.administrator.password=pass'",
	}

	if diff := cmp.Diff(args, expectedArgs,
		cmpopts.SortSlices(func(a, b string) bool {
			return a < b // particular way of sorting does not matter
		}),
	); diff != "" {
		t.Fatalf("expected args != actual: %s", diff)
	}
}

func TestValuesToYaml(t *testing.T) {
	t.Parallel()

	values := generateValues()

	// Convert to helm args
	yaml, err := values.ToYAML()
	if err != nil {
		t.Fatalf("failed to convert Values to yaml: %v", err)
	}
	lines := strings.Split(yaml, "\n")

	expectedLines := []string{
		"auth-center:",
		"  administrator:",
		"    password: pass",
		"    username: user",
		"clickhouse:",
		"  deploy: false",
		"  externalHost: clickhouse.local",
		"  persistence:",
		"    enabled: false",
		"cs-manager:",
		"  registrationToken: registration-token",
		"global:",
		"  centralCsUrl: https://central-cs.local",
		"  clickhouse:",
		"    tls:",
		"      enabled: false",
		"      verify: false",
		"  csVersion: 1.0.0",
		"  imageRegistry: registry.example.com",
		"  isChildCluster: true",
		"  keys:",
		"    encryption: encryption-key",
		"    token: token-key",
		"  ownCsUrl: https://cs.local",
		"  postgresql:",
		"    tls:",
		"      enabled: true",
		"      verify: true",
		"  redis:",
		"    tls:",
		"      enabled: false",
		"      verify: false",
		"imagePullSecret:",
		"  password: registry-password",
		"  username: registry-user",
		"metrics:",
		"  enabled: false",
		"notifier:",
		"  env:",
		"  - name: HTTP_PROXY",
		"    value: http://proxy.local",
		"  - name: HTTPS_PROXY",
		"    value: https://proxy.local",
		"postgresql:",
		"  auth:",
		"    password: postgres-password",
		"    username: postgres",
		"  deploy: true",
		"  persistence:",
		"    enabled: true",
		"    storageClass: standard-1",
		"rabbitmq:",
		"  auth:",
		"    password: rabbitmq-password",
		"    username: rabbitmq",
		"  deploy: true",
		"  persistence:",
		"    enabled: true",
		"    storageClass: standard-2",
		"redis:",
		"  auth:",
		"    password: redis-password",
		"    username: redis",
		"  deploy: true",
		"  persistence:",
		"    enabled: false",
		"reverse-proxy:",
		"  ingress:",
		"    class: nginx",
		"    enabled: true",
		"    hostname: cs.example.com",
		"    tls:",
		"      existingSecret: cs-tls",
		"  service:",
		"    type: ClusterIP",
		"tls:",
		"  cert: cert-data",
		"  certCA: ca-cert-data",
		"  certKey: cert-key-data",
		"  verify: false",
		"",
	}

	if diff := cmp.Diff(lines, expectedLines); diff != "" {
		t.Fatalf("expected args != actual: %s", diff)
	}
}
