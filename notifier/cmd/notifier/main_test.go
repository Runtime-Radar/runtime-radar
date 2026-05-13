package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	history "github.com/runtime-radar/runtime-radar/history-api/pkg/model"
	"github.com/runtime-radar/runtime-radar/lib/errcommon"
	"github.com/runtime-radar/runtime-radar/lib/logger"
	"github.com/runtime-radar/runtime-radar/lib/security"
	"github.com/runtime-radar/runtime-radar/lib/security/cipher"
	"github.com/runtime-radar/runtime-radar/lib/security/jwt"
	"github.com/runtime-radar/runtime-radar/lib/server/interceptor"
	"github.com/runtime-radar/runtime-radar/lib/util/retry"
	"github.com/runtime-radar/runtime-radar/notifier/api"
	"github.com/runtime-radar/runtime-radar/notifier/internal/mailpit"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/client"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/config"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/database"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/model"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/notifier/syslog"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/server"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/service"
	"github.com/runtime-radar/runtime-radar/notifier/pkg/template"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	listenGRPCAddr = "127.0.0.1:8000"
	sysLogFile     = "../../testdata/syslog/logs/syslog.log"
)

var (
	notifier               api.NotifierClient
	notificationController api.NotificationControllerClient
	integrationController  api.IntegrationControllerClient

	deniedNotifier               api.NotifierClient
	deniedNotificationController api.NotificationControllerClient
	deniedIntegrationController  api.IntegrationControllerClient

	unauthNotifier               api.NotifierClient
	unauthNotificationController api.NotificationControllerClient
	unauthIntegrationController  api.IntegrationControllerClient

	cfg *config.Config

	mailpitClient *mailpit.Client
)

func TestMain(m *testing.M) {
	cfg = config.New()
	cfg.EncryptionKey = hex.EncodeToString(security.Rand(32))
	cfg.TemplatesHTMLFolder = "../../templates/html"
	cfg.TemplatesTextFolder = "../../templates/text"
	syslog.TestConnectionTimeout = 1 * time.Second

	template.Init(cfg.TemplatesHTMLFolder, cfg.TemplatesTextFolder)

	if testing.Verbose() {
		logger.Init("", "DEBUG")
	} else {
		logger.Init("", "INFO")
	}

	mailpitClient = mailpit.NewClient(cfg.TestMailpitHTTPAddr)

	lis, err := net.Listen("tcp", listenGRPCAddr)
	if err != nil {
		log.Fatal().Msgf("### Failed to listen: %v", err)
	}

	crypter, err := cipher.NewCrypt(cfg.EncryptionKey)
	if err != nil {
		log.Fatal().Msgf("### Failed to parse encryption key: %v", err)
	}

	var verifier jwt.Verifier
	var tokenKey []byte
	if cfg.Auth {
		// cfg.TokenKey can be set to certain value to test some edge cases.
		// In other cases random value should be used.
		if cfg.TokenKey == "" {
			tokenBytes := security.Rand(32)
			cfg.TokenKey = hex.EncodeToString(tokenBytes)
		}

		verifier, tokenKey, err = jwt.NewKeyVerifier(cfg.TokenKey)
		if err != nil {
			log.Fatal().Msgf("### Failed to parse token key: %v", err)
		}
	}

	db, closeDB, err := database.New(
		cfg.PostgresAddr,
		cfg.PostgresDB+"_test", // <-- use test DB
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresSSLMode,
		cfg.PostgresSSLCheckCert,
	)
	if err != nil {
		log.Fatal().Msgf("### Failed to open DB: %v", err)
	}
	if err := database.Migrate(
		db,
		true, // <-- recreate test DB from scratch
	); err != nil {
		log.Fatal().Msgf("### Failed to migrate DB: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptor.Recovery, interceptor.Correlation),
		grpc.MaxRecvMsgSize(server.MaxRecvMsgSize),
	}

	var tlsConfig *tls.Config
	if cfg.TLS {
		// Load TLS config
		tlsConfig, err = security.LoadTLS(caFile, certFile, keyFile)
		if err != nil {
			log.Fatal().Msgf("### Failed to load TLS config: %v", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	ruleController, closeRC, err := client.NewRuleController(cfg.PolicyEnforcerGRPCAddr, tlsConfig, tokenKey)
	if err != nil {
		log.Fatal().Msgf("### Failed to connect to Policy Enforcer: %v", err)
	}

	grpcSrv := grpc.NewServer(opts...)
	notifier, notification, email := composeServices(db, ruleController, crypter, verifier, cfg.Auth, cfg.CSVersion)

	api.RegisterNotifierServer(grpcSrv, notifier)
	api.RegisterNotificationControllerServer(grpcSrv, notification)
	api.RegisterIntegrationControllerServer(grpcSrv, email)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal().Msgf("### Can't serve gRPC requests: %v", err)
		}
	}()

	closeClients, err := initClients(listenGRPCAddr, tlsConfig, tokenKey)
	if err != nil {
		log.Fatal().Msgf("### Can't init gRPC clients")
	}

	res := m.Run() // <-- run tests

	// This kind of tier down is not required in tests, but we want to keep everything as clean as possible
	closeClients()
	grpcSrv.GracefulStop()
	closeDB()
	closeRC()

	os.Exit(res)
}

func initClients(address string, tlsConfig *tls.Config, tokenKey []byte) (func() error, error) {
	creds := insecure.NewCredentials()
	if tlsConfig != nil {
		creds = credentials.NewTLS(tlsConfig)
	}

	if len(tokenKey) != 0 {
		return initClientsWithAuth(address, creds, tokenKey)
	}
	return initClientsWithoutAuth(address, creds)
}

func initClientsWithoutAuth(address string, creds credentials.TransportCredentials) (func() error, error) {
	goodClients, err := newClients(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	notifier = goodClients.notifier
	notificationController = goodClients.notificationController
	integrationController = goodClients.integrationController

	return goodClients.closer, nil
}

func initClientsWithAuth(address string, transportCreds credentials.TransportCredentials, tokenKey []byte) (func() error, error) {
	creds, err := generateServiceCredentials(tokenKey, []jwt.Action{jwt.ActionCreate, jwt.ActionRead, jwt.ActionUpdate, jwt.ActionDelete, jwt.ActionExecute})
	if err != nil {
		return nil, err
	}

	unauthorizedCreds, err := generateServiceCredentials(tokenKey, []jwt.Action{})
	if err != nil {
		return nil, err
	}

	goodClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds), grpc.WithPerRPCCredentials(creds))
	if err != nil {
		return nil, err
	}

	notifier = goodClients.notifier
	notificationController = goodClients.notificationController
	integrationController = goodClients.integrationController

	deniedClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds), grpc.WithPerRPCCredentials(unauthorizedCreds))
	if err != nil {
		return nil, err
	}

	deniedNotifier = deniedClients.notifier
	deniedNotificationController = deniedClients.notificationController
	deniedIntegrationController = deniedClients.integrationController

	unauthClients, err := newClients(address, grpc.WithTransportCredentials(transportCreds))
	if err != nil {
		return nil, err
	}

	unauthNotifier = unauthClients.notifier
	unauthNotificationController = unauthClients.notificationController
	unauthIntegrationController = unauthClients.integrationController

	closeClients := func() error {
		goodClients.closer()
		deniedClients.closer()
		unauthClients.closer()
		return nil
	}

	return closeClients, nil
}

func newClients(address string, opts ...grpc.DialOption) (*clients, error) {
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}

	return &clients{
		api.NewNotifierClient(conn),
		api.NewNotificationControllerClient(conn),
		api.NewIntegrationControllerClient(conn),
		conn.Close,
	}, nil
}

type clients struct {
	notifier               api.NotifierClient
	notificationController api.NotificationControllerClient
	integrationController  api.IntegrationControllerClient
	closer                 func() error
}

func generateServiceCredentials(key []byte, actions []jwt.Action) (credentials.PerRPCCredentials, error) {
	rp := &jwt.RolePermissions{
		Scopes: &jwt.Permission{
			Actions: actions,
		},
		Rules: &jwt.Permission{
			Actions: actions,
		},
		Scanning: &jwt.Permission{
			Actions: actions,
		},
	}
	return jwt.GeneratePerRPCCredentials(key, "test", rp)
}

func TestNotifyEmailE2E(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	t.Run("Usecase: create email integration and notify about image scan", func(t *testing.T) {
		t.Parallel()

		imageName := "notifier:v0.0.15"
		from := "admin@example.com"
		to := "user@example.com"

		// Subject is random string so it can be used for searching message
		// assuming that only one message with such subject exist.
		// The test may be broken if two same subjects are generated.
		subject := security.RandAlphaNum(10)
		body := "Some message to display"

		integration := newEmailIntegration(imageName, from, cfg.TestMailpitSMTPAddr)
		integrationResp, err := integrationController.Create(ctx, integration)
		if err != nil {
			t.Fatalf("Can't create integration: %v", err)
		}

		notification := newEmailNotification(imageName, integrationResp.Id, body, subject, to)
		notificationResp, err := notificationController.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Can't create notification: %v", err)
		}

		_, err = notifier.Notify(ctx, newNotifyRuntimeScanReq(imageName, notificationResp.Id))
		if err != nil {
			t.Fatalf("Can't perform notify operation: %v", err)
		}

		m, err := messageBySubject(ctx, subject)
		if err != nil {
			t.Fatalf("Can't get message by subject: %v", err)
		}

		if reason, ok := assertEmail(from, to, subject, body, m); !ok {
			t.Fatal(reason)
		}
	})
}

func TestNotifySyslogE2E(t *testing.T) {
	ctx := context.Background()

	// clear logs before tests
	_ = clearSyslogLogs()
	ruleName := "syslog_" + time.Now().String()

	nonExistSyslogTCPAddr := "tcp://10.10.10.1:606"

	t.Run("Usecase: create syslog TCP integration and notification", func(t *testing.T) {
		rule := "tcp_" + ruleName

		integration := newSyslogIntegration(cfg.TestSyslogTCPAddr, "tcp_test_1", false)
		integrationResp, err := integrationController.Create(ctx, integration)
		if err != nil {
			t.Fatalf("Can't create integration: %v", err)
		}

		templateResp, err := notificationController.DefaultTemplate(ctx, &api.DefaultTemplateReq{
			IntegrationType: model.IntegrationSyslog,
			EventType:       history.EventTypeRuntimeEvent,
		})
		if err != nil {
			t.Fatalf("Can't get notification template: %v", err)
		}

		notification := newSyslogNotification(rule, integrationResp.Id, templateResp.Template)
		notificationResp, err := notificationController.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Can't create notification: %v", err)
		}

		_, err = notifier.Notify(ctx, newNotifyRuntimeScanReq(rule, notificationResp.Id))
		if err != nil {
			t.Fatalf("Can't perform notify operation: %v", err)
		}

		if err := verifySysLog(rule); err != nil {
			t.Fatalf("Can't get syslog logs: %v", err)
		}
	})

	t.Run("Usecase: create non exist syslog TCP integration and skip check", func(t *testing.T) {
		integration := newSyslogIntegration(nonExistSyslogTCPAddr, "tcp_test_not_exist", true)
		_, err := integrationController.Create(ctx, integration)
		if err != nil {
			t.Fatalf("Can't create integration: %v", err)
		}
	})

	t.Run("Usecase: create non exist syslog TCP integration and not skip check", func(t *testing.T) {
		integration := newSyslogIntegration(nonExistSyslogTCPAddr, "tcp_test_not_exist_not_skip", false)
		_, err := integrationController.Create(ctx, integration)
		if err == nil {
			t.Fatalf("Integration without verification was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		} else if !strings.Contains(st.Message(), "integration is inaccessible") {
			t.Fatalf("Incorrect status message: '%v' does not contains 'integration is inaccessible'", st.Message())
		}
	})

	t.Run("Usecase: should return error when creating integration with duplicate name", func(t *testing.T) {
		integration := newSyslogIntegration(cfg.TestSyslogTCPAddr, "tcp_double", false)

		_, err := integrationController.Create(ctx, integration)
		if err != nil {
			t.Fatalf("Can't create integration: %v", err)
		}

		_, err = integrationController.Create(ctx, integration)
		if err == nil {
			t.Fatal("Expected an error when creating a duplicate integration, but got none")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.AlreadyExists {
			t.Fatalf("Incorrect status code: 'codes.AlreadyExists' != '%v'", st.Code())
		} else if reason, ok := errcommon.ReasonFromStatus(st); !ok {
			t.Fatal("Reason should exist")
		} else if reason != service.NameMustBeUnique {
			t.Fatalf("Incorrect reason in status: 'NAME_MUST_BE_UNIQUE' != '%v'", reason)
		}
	})

	t.Run("Usecase: create syslog UDP integration and notification", func(t *testing.T) {
		image := "udp_" + ruleName

		integration := newSyslogIntegration(cfg.TestSyslogUDPAddr, "tcp_test_2", false)
		integrationResp, err := integrationController.Create(ctx, integration)
		if err != nil {
			t.Fatalf("Can't create integration: %v", err)
		}

		templateResp, err := notificationController.DefaultTemplate(ctx, &api.DefaultTemplateReq{
			IntegrationType: model.IntegrationSyslog,
			EventType:       history.EventTypeRuntimeEvent,
		})
		if err != nil {
			t.Fatalf("Can't get notification template: %v", err)
		}

		notification := newSyslogNotification(image, integrationResp.Id, templateResp.Template)
		notificationResp, err := notificationController.Create(ctx, notification)
		if err != nil {
			t.Fatalf("Can't create notification: %v", err)
		}

		_, err = notifier.Notify(ctx, newNotifyRuntimeScanReq(image, notificationResp.Id))
		if err != nil {
			t.Fatalf("Can't perform notify operation: %v", err)
		}

		if err := verifySysLog(image); err != nil {
			t.Fatalf("Can't get syslog logs: %v", err)
		}
	})

	t.Run("Usecase: unexpected port for syslog TCP connection", func(t *testing.T) {
		integration := newSyslogIntegration("tcp://127.0.0.1:670", "tcp_test_3", false)
		_, err := integrationController.Create(ctx, integration)
		if err == nil {
			t.Fatal("Integration without verification was allowed")
		} else if st, ok := status.FromError(err); !ok {
			t.Fatalf("Non-status error: %v", err)
		} else if st.Code() != codes.InvalidArgument {
			t.Fatalf("Incorrect status code: 'codes.InvalidArgument' != '%v'", st.Code())
		} else if !strings.Contains(st.Message(), "connection refused") {
			t.Fatalf("Incorrect status message: '%v' does not contains 'connection refused'", st.Message())
		}
	})
}

func newEmailIntegration(name, from, addr string) *api.Integration {
	i := &api.Integration{
		Name:      name,
		Type:      model.IntegrationEmail,
		SkipCheck: false,
		Config: &api.Integration_Email{
			Email: &api.Email{
				From:     from,
				Server:   addr,
				AuthType: api.Email_AUTH_TYPE_NONE,
			},
		},
	}
	return i
}

func newEmailNotification(name, integrationID, template, subjectTemplate, to string) *api.Notification {
	n := &api.Notification{
		Name:            name,
		IntegrationType: model.IntegrationEmail,
		IntegrationId:   integrationID,
		Recipients:      []string{to},
		EventType:       history.EventTypeRuntimeEvent,
		Template:        template,
		Config: &api.Notification_Email{
			Email: &api.EmailConfig{
				SubjectTemplate: subjectTemplate,
			},
		},
	}
	return n
}

func newNotifyRuntimeScanReq(imageName, notificationID string) *api.NotifyReq {
	req := &api.NotifyReq{
		Notifications: []*api.Message{
			{
				NotificationId: notificationID,
				Event: &api.Message_RuntimeEvent{
					RuntimeEvent: &api.RuntimeEvent{
						RuleName: imageName,
						Event:    &api.RuntimeEvent_Event{},
					},
				},
			},
		},
	}
	return req
}

func assertEmail(from, to, subj, body string, actual *mailpit.Message) (string, bool) {
	if expected, actual := fmt.Sprintf("<%s>", from), actual.From.String(); expected != actual {
		return fmt.Sprintf("expected from to be %s, got %s", expected, actual), false
	}

	if expected, actual := fmt.Sprintf("<%s>", to), actual.To[0].String(); expected != actual {
		return fmt.Sprintf("expected from to be %s, got %s", expected, actual), false
	}

	if subj != actual.Subject {
		return fmt.Sprintf("expected subject to be '%s', got '%s'", subj, actual.Subject), false
	}

	if body != actual.Text {
		return fmt.Sprintf("expected body to be '%s', got '%s'", body, actual.Text), false
	}

	return "", true
}

// messageBySubject finds message by subject assuming that only one message with given subject exists.
// If more than one message is returned by mailpit API, error is returned.
func messageBySubject(ctx context.Context, subj string) (*mailpit.Message, error) {
	ms, err := mailpitClient.MessagesBySubject(ctx, subj)
	if err != nil {
		return nil, err
	}

	if len(ms.Messages) != 1 {
		return nil, fmt.Errorf("expected 1 message in response, got %d", len(ms.Messages))
	}

	return mailpitClient.MessageByID(ctx, ms.Messages[0].ID)
}

func newSyslogIntegration(addr, name string, skip bool) *api.Integration {
	return &api.Integration{
		Name:      name,
		Type:      model.IntegrationSyslog,
		SkipCheck: skip,
		Config: &api.Integration_Syslog{
			Syslog: &api.Syslog{
				Address: addr,
			},
		},
	}
}

func newSyslogNotification(name, integrationID, template string) *api.Notification {
	n := &api.Notification{
		Name:            name,
		IntegrationType: model.IntegrationSyslog,
		IntegrationId:   integrationID,
		Recipients:      []string{},
		EventType:       history.EventTypeRuntimeEvent,
		Template:        template,
		Config: &api.Notification_Syslog{
			Syslog: &api.SyslogConfig{},
		},
	}
	return n
}

func clearSyslogLogs() error {
	f, err := os.OpenFile(sysLogFile, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Truncate(0)
}

func syslogLogs() (string, error) {
	file, err := os.Open(sysLogFile)
	if err != nil {
		return "", fmt.Errorf("can't open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("can't read file: %w", err)
	}

	return string(content), nil
}

func verifySysLog(rule string) error {
	c := retry.NewDefaultConfig()
	c.MaxAttempts = 10
	c.Delay = 200 * time.Millisecond

	return retry.Do(func() error {
		result, err := syslogLogs()
		if err != nil {
			return err
		}

		if !strings.Contains(result, `"rule_name":"`+rule+`"`) {
			return errors.New("syslog not found")
		}

		return nil
	}, c)
}
