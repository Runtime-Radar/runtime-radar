package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/runtime-radar/runtime-radar/cluster-manager/api"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/database"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/helm"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/model"
	"github.com/runtime-radar/runtime-radar/cluster-manager/pkg/model/convert"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/security/cipher"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type ClusterGeneric struct {
	api.UnimplementedClusterControllerServer

	ClusterRepository        database.ClusterRepository
	Crypter                  cipher.Crypter
	UseAuth                  bool
	UseTLS                   bool
	CertCA                   string
	CertPEM                  string
	KeyPEM                   string
	TokenKey                 string
	EncryptionKey            string
	PublicAccessTokenSaltKey string
	CSVersion                string

	AdministratorUsername string
	AdministratorPassword string
}

func (cg *ClusterGeneric) Create(ctx context.Context, req *api.Cluster) (*api.CreateClusterResp, error) {
	if reason, ok := cg.validateCluster(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	var id uuid.UUID
	var err error

	if idStr := req.GetId(); idStr != "" {
		id, err = uuid.Parse(idStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
		}
	}

	cfg := (*model.ClusterConfig)(req.GetConfig())
	cfg.EncryptSensitive(cg.Crypter)

	c := &model.Cluster{
		model.Base{ID: id},
		req.GetName(),
		uuid.Nil,
		model.ClusterStatusUnregistered, // cluster should be registered explicitly
		cfg,
		nil,
		gorm.DeletedAt{},
	}

	if err := cg.ClusterRepository.Add(ctx, c); err != nil {
		if errors.Is(err, model.ErrClusterNameInUse) {
			return nil, errcommon.StatusWithReason(codes.AlreadyExists, NameMustBeUnique, "name field must be unique").Err()
		}
		return nil, status.Errorf(codes.Internal, "can't add cluster: %v", err)
	}

	resp := &api.CreateClusterResp{
		Id: c.ID.String(),
	}

	return resp, nil
}

func (cg *ClusterGeneric) Read(ctx context.Context, req *api.ReadClusterReq) (*api.ReadClusterResp, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	c, err := cg.ClusterRepository.GetByID(ctx, id, true) // preload is on
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get cluster by id: %v", err)
	}

	resp := &api.ReadClusterResp{
		Cluster: convert.ClusterToProto(c, true),
		Deleted: c.DeletedAt.Valid,
	}

	return resp, nil
}

func (cg *ClusterGeneric) Update(ctx context.Context, req *api.Cluster) (*emptypb.Empty, error) {
	if reason, ok := cg.validateCluster(req); !ok {
		return nil, status.Error(codes.InvalidArgument, reason)
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	cfg := (*model.ClusterConfig)(req.GetConfig())
	cfg.EncryptSensitive(cg.Crypter)

	m := map[string]any{
		"Name":   req.GetName(),
		"Config": cfg,
	}

	if err := cg.ClusterRepository.UpdateWithMap(ctx, id, m); err != nil {
		if errors.Is(err, model.ErrClusterNameInUse) {
			return nil, errcommon.StatusWithReason(codes.AlreadyExists, NameMustBeUnique, "name field must be unique").Err()
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		}

		return nil, status.Errorf(codes.Internal, "can't update cluster: %v", err)
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (cg *ClusterGeneric) Delete(ctx context.Context, req *api.DeleteClusterReq) (*emptypb.Empty, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	// this method only allows to delete unregistered clusters
	// deleting registered clusters is performed as a part of unregistration process
	filter := gorm.Expr("id = ? and status = ?", id, model.ClusterStatusUnregistered)

	if err := cg.ClusterRepository.DeleteByFilter(ctx, filter); err != nil {
		return nil, status.Errorf(codes.Internal, "can't delete cluster: %v", err)
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (cg *ClusterGeneric) ListPage(ctx context.Context, req *api.ListClusterPageReq) (*api.ListClusterPageResp, error) {
	total, err := cg.ClusterRepository.GetCount(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get cluster count: %v", err)
	}

	order, pageSize := req.GetOrder(), req.GetPageSize()
	if order == "" {
		order = defaultOrder
	}
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	cs, err := cg.ClusterRepository.GetPage(ctx, nil, order, int(pageSize), int(req.GetPageNum()), true) // preload is on
	if err != nil {
		if errors.Is(err, database.ErrInvalidOrder) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "can't get cluster page: %v", err)
	}

	resp := &api.ListClusterPageResp{
		Total:    uint32(total),
		Clusters: convert.ClustersToProto(cs, true),
	}

	return resp, nil
}

func (cg *ClusterGeneric) Register(ctx context.Context, req *api.RegisterClusterReq) (*emptypb.Empty, error) {
	token, err := uuid.Parse(req.GetToken())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse token: %v", err)
	}

	cluster, err := cg.ClusterRepository.GetByToken(ctx, token, false) // preload is off
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get cluster by token: %v", err)
	}

	if cluster.Status == model.ClusterStatusRegistered {
		return nil, status.Error(codes.Canceled, "cluster already registered")
	}

	m := map[string]any{
		"Status":       model.ClusterStatusRegistered,
		"RegisteredAt": time.Now(),
	}

	if err := cg.ClusterRepository.UpdateWithMap(ctx, cluster.ID, m); err != nil {
		return nil, status.Errorf(codes.Internal, "can't update cluster: %v", err)
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (cg *ClusterGeneric) Unregister(ctx context.Context, req *api.UnregisterClusterReq) (*emptypb.Empty, error) {
	token, err := uuid.Parse(req.GetToken())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse token: %v", err)
	}

	// this method only allows to delete registered clusters
	// deleting unregistered clusters is performed with Delete method
	filter := gorm.Expr("token = ? and status = ?", token, model.ClusterStatusRegistered)

	if err := cg.ClusterRepository.DeleteByFilter(ctx, filter); err != nil {
		return nil, status.Errorf(codes.Internal, "can't delete cluster: %v", err)
	}

	resp := &emptypb.Empty{}

	return resp, nil
}

func (cg *ClusterGeneric) GenerateUninstallCmd(ctx context.Context, req *api.GenerateUninstallCmdReq) (*api.GenerateUninstallCmdResp, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	cluster, err := cg.ClusterRepository.GetByID(ctx, id, true) // preload is on
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get cluster by id: %v", err)
	}

	cfg := cluster.Config

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = helm.DefaultNamespace
	}

	cmd := helm.UninstallCmd(namespace)

	resp := &api.GenerateUninstallCmdResp{
		Cmd: cmd,
	}

	return resp, nil
}

func (cg *ClusterGeneric) GenerateInstallCmd(ctx context.Context, req *api.GenerateInstallCmdReq) (*api.GenerateInstallCmdResp, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	cluster, err := cg.ClusterRepository.GetByID(ctx, id, true) // preload is on
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get cluster by id: %v", err)
	}

	cfg := cluster.Config
	cfg.DecryptSensitive(cg.Crypter)

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = helm.DefaultNamespace
	}

	var v *helm.Values

	if !req.GetUseValuesFile() {
		v = cg.buildValues(cfg, cluster.Token.String())
	}

	cmd, err := helm.UpgradeCmd(cfg.Registry.Address, cg.CSVersion, namespace, v)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get upgrade cmd: %v", err)
	}

	resp := &api.GenerateInstallCmdResp{
		Cmd: cmd,
	}

	return resp, nil
}

func (cg *ClusterGeneric) GenerateValuesYAML(ctx context.Context, req *api.GenerateValuesReq) (*api.GenerateValuesResp, error) {
	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse ID: %v", err)
	}

	cluster, err := cg.ClusterRepository.GetByID(ctx, id, true) // preload is on
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "cluster not found")
		}
		return nil, status.Errorf(codes.Internal, "can't get cluster by id: %v", err)
	}

	cfg := cluster.Config
	cfg.DecryptSensitive(cg.Crypter)

	v := cg.buildValues(cfg, cluster.Token.String())

	values, err := v.ToYAML()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't marshal values.yaml: %v", err)
	}

	resp := &api.GenerateValuesResp{
		Yaml: values,
	}

	return resp, nil
}

func (cg *ClusterGeneric) buildValues(cfg *model.ClusterConfig, token string) *helm.Values {
	var v helm.Values

	v.Global.CSVersion = cg.CSVersion
	v.Global.IsChildCluster = true
	v.Global.OwnCSURL = cfg.OwnCsUrl
	v.Global.CentralCSURL = cfg.CentralCsUrl

	// Registration token
	v.CSManager.RegistrationToken = token

	// Proxy
	if cfg.ProxyUrl != "" {
		v.Notifier.Env = []helm.Env{
			{
				Name:  "HTTP_PROXY",
				Value: cfg.ProxyUrl,
			},
			{
				Name:  "HTTPS_PROXY",
				Value: cfg.ProxyUrl,
			},
		}
	}

	// Registry
	registry := cfg.Registry
	v.Global.ImageRegistry = registry.Address
	v.Global.ImageShortNames = registry.ImageShortNames
	if registry.Password != "" {
		v.ImagePullSecret.Username = registry.User
		v.ImagePullSecret.Password = registry.Password
	}

	// Auth
	if cg.UseAuth {
		v.Global.Keys.Encryption = cg.EncryptionKey
		v.Global.Keys.Token = cg.TokenKey
		v.Global.Keys.PublicAccessTokenSalt = cg.PublicAccessTokenSaltKey

	}

	// History API
	interval := cfg.HistoryRetentionInterval
	if interval == "" {
		interval = helm.DefaultRetentionInterval.String()
	}
	v.HistoryAPI.RetentionInterval = interval

	// Postgres
	postgres := cfg.Postgres
	postgresUser := postgres.User
	if postgresUser == "" {
		postgresUser = helm.DefaultUser
	}
	v.Postgresql.Auth.Username = postgresUser
	v.Postgresql.Auth.Password = postgres.Password
	v.Postgresql.Auth.Database = postgres.Database
	v.Postgresql.Persistence.Enabled = postgres.Persistence
	v.Global.Postgresql.TLS.Enabled = postgres.UseTls
	v.Global.Postgresql.TLS.Verify = postgres.CheckCert
	v.Postgresql.Deploy = true
	if postgres.Address != "" {
		v.Postgresql.Deploy = false
		v.Postgresql.ExternalHost = postgres.Address
		v.Postgresql.TLS.CertCA = postgres.Ca
	} else if postgres.Persistence {
		v.Postgresql.Persistence.StorageClass = postgres.StorageClass
	}

	// Redis
	redis := cfg.Redis
	redisUser := redis.User
	if redisUser == "" {
		redisUser = helm.DefaultUser
	}

	v.Redis.Auth.Username = redisUser
	v.Redis.Auth.Password = redis.Password
	v.Redis.Persistence.Enabled = redis.Persistence
	v.Global.Redis.TLS.Enabled = redis.UseTls
	v.Global.Redis.TLS.Verify = redis.CheckCert
	v.Redis.Deploy = true
	if redis.Address != "" {
		v.Redis.Deploy = false
		v.Redis.ExternalHost = redis.Address
		v.Redis.TLS.CertCA = redis.Ca
	} else if redis.Persistence {
		v.Redis.Persistence.StorageClass = redis.StorageClass
	}

	// Rabbit
	rabbit := cfg.Rabbit
	rabbitUser := rabbit.User
	if rabbitUser == "" {
		rabbitUser = helm.DefaultUser
	}

	v.Rabbitmq.Auth.Username = rabbitUser
	v.Rabbitmq.Auth.Password = rabbit.Password
	v.Rabbitmq.Persistence.Enabled = rabbit.Persistence
	v.Rabbitmq.Deploy = true
	if rabbit.Address != "" {
		v.Rabbitmq.Deploy = false
		v.Rabbitmq.ExternalHost = rabbit.Address
	} else if rabbit.Persistence {
		v.Rabbitmq.Persistence.StorageClass = rabbit.StorageClass
	}

	// Clickhouse
	clickhouse := cfg.Clickhouse
	clickhouseUser := clickhouse.User
	if clickhouseUser == "" {
		clickhouseUser = helm.DefaultUser
	}

	v.Clickhouse.Auth.Username = clickhouseUser
	v.Clickhouse.Auth.Password = clickhouse.Password
	v.Clickhouse.Auth.Database = clickhouse.Database
	v.Clickhouse.Persistence.Enabled = clickhouse.Persistence
	v.Global.Clickhouse.TLS.Enabled = clickhouse.UseTls
	v.Global.Clickhouse.TLS.Verify = clickhouse.CheckCert
	v.Clickhouse.Deploy = true
	if clickhouse.Address != "" {
		v.Clickhouse.Deploy = false
		v.Clickhouse.ExternalHost = clickhouse.Address
		v.Clickhouse.TLS.CertCA = clickhouse.Ca
	} else if clickhouse.Persistence {
		v.Clickhouse.Persistence.StorageClass = clickhouse.StorageClass
	}

	if cfg.EnableMetrics {
		// Grafana
		grafana := cfg.Grafana
		grafanaUser := grafana.User
		if grafanaUser == "" {
			grafanaUser = helm.DefaultUser
		}

		v.Grafana.Auth.Username = grafanaUser
		v.Grafana.Auth.Password = grafana.Password
		v.Grafana.Deploy = true
		if grafana.Address != "" {
			v.Grafana.Deploy = false
			v.Grafana.ExternalHost = grafana.Address
		}

		// Prometheus
		prometheus := cfg.Prometheus
		if prometheus.Deploy {
			v.Prometheus.Deploy = true
			v.Prometheus.Persistence.Enabled = prometheus.Persistence
			v.Prometheus.Persistence.StorageClass = prometheus.StorageClass
		}

		// Metrics
		v.Metrics.Enabled = true
	}

	// Ingress
	if ingress := cfg.Ingress; ingress != nil {
		v.ReverseProxy.Ingress.Enabled = true
		v.ReverseProxy.Ingress.Class = ingress.IngressClass
		if ingress.Hostname != "" {
			v.ReverseProxy.Ingress.Hostname = ingress.Hostname
		}
		if ingress.Cert != "" {
			v.ReverseProxy.Ingress.TLS.CertCA = ingress.Cert
			v.ReverseProxy.Ingress.TLS.Cert = ingress.Cert
			v.ReverseProxy.Ingress.TLS.CertKey = ingress.CertKey
		}
	}

	// NodePort
	if nodePort := cfg.NodePort; nodePort != nil {
		v.ReverseProxy.Service.Type = "NodePort"
		if nodePort.Port != "" {
			v.ReverseProxy.Service.NodePorts.HTTP = nodePort.Port
		}
	}

	// AuthAPI
	v.AuthAPI.Administrator.Username = cg.AdministratorUsername
	v.AuthAPI.Administrator.Password = cg.AdministratorPassword

	return &v
}

func (cg *ClusterGeneric) ListRegistered(ctx context.Context, _ *emptypb.Empty) (*api.ListRegisteredResp, error) {
	filter := &model.Cluster{Status: model.ClusterStatusRegistered}
	order := defaultOrder
	cs, err := cg.ClusterRepository.GetAll(ctx, filter, order, true)
	if err != nil {
		if errors.Is(err, database.ErrInvalidOrder) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "can't get registered clusters: %v", err)
	}

	csProto := make([]*api.ListRegisteredResp_Cluster, 0, len(cs))
	for _, c := range cs {
		csProto = append(csProto, &api.ListRegisteredResp_Cluster{
			Id:       c.ID.String(),
			Name:     c.Name,
			OwnCsUrl: c.Config.OwnCsUrl,
		})
	}

	resp := &api.ListRegisteredResp{
		Clusters: csProto,
	}

	return resp, nil
}

func (cg *ClusterGeneric) validateCluster(c *api.Cluster) (reason string, ok bool) {
	if c.GetName() == "" {
		return "empty or missing name", false
	}

	cfg := c.GetConfig()

	if cfg == nil {
		return "missing cluster config", false
	}

	if v := cfg.GetVersion(); v == "" {
		return "empty or missing config version", false
	} else if v != string(model.ClusterConfigVersion) {
		return fmt.Sprintf("config metadata version mismatch: expected %s, got %s", model.ClusterConfigVersion, v), false
	}

	if cfg.GetOwnCsUrl() == "" {
		return "empty or missing own cs url", false
	}

	if cfg.GetCentralCsUrl() == "" {
		return "empty or missing central cs url", false
	}

	if pc := cfg.GetPostgres(); pc == nil {
		return "empty or missing postgres config", false
	} else if pc.GetPassword() == "" {
		return "postgres: empty or missing password", false
	}

	if cc := cfg.GetClickhouse(); cc == nil {
		return "empty or missing clickhouse config", false
	} else if cc.GetPassword() == "" {
		return "clickhouse: empty or missing password", false
	}

	if rc := cfg.GetRedis(); rc == nil {
		return "empty or missing redis config", false
	} else if rc.GetPassword() == "" {
		return "redis: empty or missing password", false
	}

	if rc := cfg.GetRabbit(); rc == nil {
		return "empty or missing rabbit config", false
	} else if rc.GetPassword() == "" {
		return "rabbit: empty or missing password", false
	}

	if rc := cfg.GetRegistry(); rc == nil {
		return "empty or missing registry config", false
	} else if _, err := helm.ConvertRegistryToOCI(rc.GetAddress()); err != nil {
		return "registry: incorrect address", false
	} else if u := rc.GetUser(); u != "" && rc.GetPassword() == "" {
		return fmt.Sprintf("registry: missing password for user %s", u), false
	}

	if ic := cfg.GetIngress(); ic != nil {
		if ic.GetCert() != "" {
			if ic.GetCertKey() == "" {
				return "ingress: both cert and key should be presented", false
			}
			if ic.GetHostname() == "" {
				return "ingress: cert requires hostname to be presented", false
			}
		}
	}

	if npc := cfg.GetNodePort(); npc != nil && npc.GetPort() == "" {
		return "nodeport: port should be presented", false
	}

	if hri := cfg.GetHistoryRetentionInterval(); hri != "" {
		if _, err := time.ParseDuration(hri); err != nil {
			return fmt.Sprintf("invalid history retention interval: %v", err), false
		}
	}

	if cfg.GetGrafana() != nil {
		// external grafana or grafana password required
		if cfg.GetGrafana().GetAddress() == "" && cfg.GetGrafana().GetPassword() == "" {
			return "grafana: empty or missing grafana password", false
		}
	}

	return "", true
}
