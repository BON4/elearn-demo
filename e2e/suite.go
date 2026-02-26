package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

type TestSuite struct {
	Ctx       context.Context
	Postgres  *postgres.PostgresContainer
	Rabbit    *rabbitmq.RabbitMQContainer
	AppURL    string
	DBUrl     string
	RabbitURL string
	AppCmd    *exec.Cmd
}

func SetupSuite(t *testing.T) *TestSuite {
	ctx := context.Background()

	// Get the workspace root from the current file location
	_, currentFile, _, _ := runtime.Caller(0)
	workspaceRoot := filepath.Dir(filepath.Dir(currentFile))

	// Build binary
	binaryPath := filepath.Join(t.TempDir(), "course-service")
	buildCmd := exec.Command(
		"go", "build",
		"-o", binaryPath,
		filepath.Join(workspaceRoot, "course-service/cmd/main.go"),
	)
	buildCmd.Dir = workspaceRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, string(output))
	}

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("courses"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatal(err)
	}

	pgURL, _ := pgContainer.ConnectionString(ctx)

	rabbitContainer, err := rabbitmq.Run(ctx,
		"rabbitmq:3.13-management",
	)
	if err != nil {
		t.Fatal(err)
	}

	rabbitURL, _ := rabbitContainer.AmqpURL(ctx)

	// Start app binary
	cmd := exec.Command(binaryPath)
	cmd.Env = append(os.Environ(),
		"DATABASE_URL="+pgURL,
		"REBBITMQ_URL="+rabbitURL,
		"HTTP_PORT=8083",
		"OUTBOX_WORKER_INTERVAL=333ms",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start service: %v", err)
	}

	time.Sleep(3 * time.Second)

	return &TestSuite{
		Ctx:       ctx,
		Postgres:  pgContainer,
		Rabbit:    rabbitContainer,
		AppURL:    "http://localhost:8083",
		DBUrl:     pgURL,
		RabbitURL: rabbitURL,
		AppCmd:    cmd,
	}
}

func (s *TestSuite) TearDown(t *testing.T) {
	if s.AppCmd != nil && s.AppCmd.Process != nil {
		_ = s.AppCmd.Process.Kill()
		_ = s.AppCmd.Wait()
	}
	if s.Postgres != nil {
		_ = s.Postgres.Terminate(s.Ctx)
	}
	if s.Rabbit != nil {
		_ = s.Rabbit.Terminate(s.Ctx)
	}
}
