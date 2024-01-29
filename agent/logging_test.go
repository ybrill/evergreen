package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/evergreen-ci/evergreen/agent/command"
	"github.com/evergreen-ci/evergreen/agent/internal"
	"github.com/evergreen-ci/evergreen/agent/internal/client"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/task"
	_ "github.com/evergreen-ci/evergreen/testutil"
	"github.com/evergreen-ci/evergreen/util"
	"github.com/mongodb/jasper"
	"github.com/mongodb/jasper/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func TestGetSenderLocal(t *testing.T) {
	assert := assert.New(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := (&Agent{}).GetSender(ctx, LogOutputStdout, "agent", "task_id", 2)
	assert.NoError(err)
}

func TestAgentFileLogging(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	tmpDirName := t.TempDir()
	require.NoError(os.Mkdir(fmt.Sprintf("%s/tmp", tmpDirName), 0666))

	agt := &Agent{
		opts: Options{
			HostID:           "host",
			HostSecret:       "secret",
			StatusPort:       2286,
			LogOutput:        LogOutputStdout,
			LogPrefix:        "agent",
			WorkingDirectory: tmpDirName,
		},
		comm:   client.NewHostCommunicator("www.example.com", "host", "secret"),
		tracer: otel.GetTracerProvider().Tracer("noop_tracer"),
	}
	jpm, err := jasper.NewSynchronizedManager(false)
	require.NoError(err)
	agt.jasper = jpm
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run a task with a command logger specified as a file.
	taskID := "logging"
	taskSecret := "mock_task_secret"
	task := &task.Task{
		Project:        "project",
		Id:             "t1",
		Execution:      0,
		DisplayName:    "task1",
		TaskOutputInfo: initializeTaskOutput(t),
	}
	tc := &taskContext{
		task: client.TaskData{
			ID:     taskID,
			Secret: taskSecret,
		},
		ranSetupGroup: false,
		oomTracker:    &mock.OOMTracker{},
		taskConfig: &internal.TaskConfig{
			Task:         *task,
			BuildVariant: model.BuildVariant{Name: "bv"},
			Project: model.Project{
				Tasks: []model.ProjectTask{
					{Name: "task1", Commands: []model.PluginCommandConf{
						{
							Command: "shell.exec",
							Params: map[string]interface{}{
								"shell":  "bash",
								"script": "echo 'hello world'",
							},
							Loggers: &model.LoggerConfig{
								Agent:  []model.LogOpts{{Type: model.FileLogSender}},
								System: []model.LogOpts{{Type: model.FileLogSender}},
								Task:   []model.LogOpts{{Type: model.FileLogSender}},
							},
						},
					}},
				},
				Loggers: &model.LoggerConfig{
					Agent:  []model.LogOpts{{Type: model.FileLogSender}},
					System: []model.LogOpts{{Type: model.FileLogSender}},
					Task:   []model.LogOpts{{Type: model.FileLogSender}},
				},
				BuildVariants: model.BuildVariants{
					{Name: "bv", Tasks: []model.BuildVariantTaskUnit{{Name: "task1", Variant: "bv"}}},
				},
			},
			Timeout:    internal.Timeout{IdleTimeoutSecs: 15, ExecTimeoutSecs: 15},
			WorkDir:    tmpDirName,
			Expansions: *util.NewExpansions(nil),
		},
	}
	require.NoError(agt.startLogging(ctx, tc))
	defer func() {
		agt.removeTaskDirectory(tc)
		assert.NoError(tc.logger.Close())
	}()
	require.NoError(agt.runTaskCommands(ctx, tc))

	// Verify log contents.
	f, err := os.Open(fmt.Sprintf("%s/%s/%s/task.log", tmpDirName, taskLogDirectory, "shell.exec"))
	require.NoError(err)
	bytes, err := io.ReadAll(f)
	assert.NoError(f.Close())
	require.NoError(err)
	assert.Contains(string(bytes), "hello world")
}

func TestStartLogging(t *testing.T) {
	tmpDirName := t.TempDir()
	agt := &Agent{
		opts: Options{
			HostID:           "host",
			HostSecret:       "secret",
			StatusPort:       2286,
			LogOutput:        LogOutputStdout,
			LogPrefix:        "agent",
			WorkingDirectory: tmpDirName,
		},
		comm: client.NewMock("url"),
	}
	tc := &taskContext{
		task: client.TaskData{
			ID:     "logging",
			Secret: "task_secret",
		},
		oomTracker: &mock.OOMTracker{},
	}

	ctx := context.Background()
	config, err := agt.makeTaskConfig(ctx, tc)
	require.NoError(t, err)
	tc.taskConfig = config

	assert.EqualValues(t, model.EvergreenLogSender, tc.taskConfig.Project.Loggers.Agent[0].Type)
	assert.EqualValues(t, model.SplunkLogSender, tc.taskConfig.Project.Loggers.System[0].Type)
	assert.EqualValues(t, model.FileLogSender, tc.taskConfig.Project.Loggers.Task[0].Type)

	require.NoError(t, agt.startLogging(ctx, tc))
	tc.logger.Execution().Info("foo")
	assert.NoError(t, tc.logger.Close())
	lines := agt.comm.(*client.Mock).GetTaskLogs(tc.taskConfig.Task.Id)
	require.Len(t, lines, 1)
	assert.Equal(t, "foo", lines[0].Data)

	// Check that expansions are correctly populated.
	logConfig := agt.prepLogger(tc, tc.taskConfig.Project.Loggers, "")
	assert.Equal(t, "bar", logConfig.System[0].SplunkToken)
}

func TestDefaultSender(t *testing.T) {
	agt := &Agent{
		opts: Options{
			HostID:     "host",
			HostSecret: "secret",
			StatusPort: 2286,
			LogOutput:  LogOutputStdout,
			LogPrefix:  "agent",
		},
		comm: client.NewMock("mock"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskID := "logging"
	taskSecret := "mock_task_secret"
	task := task.Task{
		DisplayName: "task1",
	}
	tc := &taskContext{
		oomTracker: &mock.OOMTracker{},
		task: client.TaskData{
			ID:     taskID,
			Secret: taskSecret,
		},
		taskConfig: &internal.TaskConfig{
			Task:         task,
			BuildVariant: model.BuildVariant{Name: "bv"},
			Timeout:      internal.Timeout{IdleTimeoutSecs: 15, ExecTimeoutSecs: 15},
			Project:      model.Project{},
		},
	}

	t.Run("Valid", func(t *testing.T) {
		tc.taskConfig.Project.Loggers = &model.LoggerConfig{}
		tc.taskConfig.ProjectRef = model.ProjectRef{DefaultLogger: model.BuildloggerLogSender}

		assert.NoError(t, agt.startLogging(ctx, tc))
		expectedLogOpts := []model.LogOpts{{Type: model.BuildloggerLogSender}}
		assert.Equal(t, expectedLogOpts, tc.taskConfig.Project.Loggers.Agent)
		assert.Equal(t, expectedLogOpts, tc.taskConfig.Project.Loggers.System)
		assert.Equal(t, expectedLogOpts, tc.taskConfig.Project.Loggers.Task)
	})
	t.Run("Invalid", func(t *testing.T) {
		tc.taskConfig.Project.Loggers = &model.LoggerConfig{}
		tc.taskConfig.ProjectRef = model.ProjectRef{DefaultLogger: model.SplunkLogSender}

		assert.NoError(t, agt.startLogging(ctx, tc))
		expectedLogOpts := []model.LogOpts{{Type: model.EvergreenLogSender}}
		assert.Equal(t, expectedLogOpts, tc.taskConfig.Project.Loggers.Agent)
		assert.Equal(t, expectedLogOpts, tc.taskConfig.Project.Loggers.System)
		assert.Equal(t, expectedLogOpts, tc.taskConfig.Project.Loggers.Task)
	})
}

func TestTimberSender(t *testing.T) {
	agt := &Agent{
		opts: Options{
			HostID:     "host",
			HostSecret: "secret",
			StatusPort: 2286,
			LogOutput:  LogOutputStdout,
			LogPrefix:  "agent",
		},
		comm: client.NewMock("mock"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskID := "logging"
	taskSecret := "mock_task_secret"
	task := &task.Task{
		DisplayName: "task1",
	}
	tc := &taskContext{
		oomTracker: &mock.OOMTracker{},
		task: client.TaskData{
			ID:     taskID,
			Secret: taskSecret,
		},
		taskConfig: &internal.TaskConfig{
			Task:         *task,
			BuildVariant: model.BuildVariant{Name: "bv"},
			Timeout:      internal.Timeout{IdleTimeoutSecs: 15, ExecTimeoutSecs: 15},
			Project: model.Project{
				Loggers: &model.LoggerConfig{
					Agent:  []model.LogOpts{{Type: model.BuildloggerLogSender}},
					System: []model.LogOpts{{Type: model.BuildloggerLogSender}},
					Task:   []model.LogOpts{{Type: model.BuildloggerLogSender}},
				},
			},
		},
	}
	assert.NoError(t, agt.startLogging(ctx, tc))
}

func TestRedactList(t *testing.T) {
	for _, test := range []struct {
		name        string
		projectVars map[string]string
		redacted    map[string]bool
		expected    []string
	}{
		{
			name: "Defaults",
		},
		{
			name: "SuspiciousPatterns",
			projectVars: map[string]string{
				"not_suspicious":  "",
				"num_hosts":       "",
				"aws_key":         "",
				"my_secret":       "",
				"my_SECRET":       "",
				"aws_token":       "",
				"AWSTOKEN":        "",
				"my_password":     "",
				"servicePassword": "",
				"srv_pass":        "",
				"my_PASS":         "",
				"mypw":            "",
				"myPW":            "",
				"private_key":     "",
				"Some_Auth":       "",
			},
			expected: []string{
				"aws_key",
				"my_secret",
				"my_SECRET",
				"aws_token",
				"AWSTOKEN",
				"my_password",
				"servicePassword",
				"srv_pass",
				"my_PASS",
				"mypw",
				"myPW",
				"private_key",
				"Some_Auth",
			},
		},
		{
			name: "Redacted",
			projectVars: map[string]string{
				"not_suspicious": "",
				"num_hosts":      "",
				"aws_token":      "",
				"my_secret":      "",
			},
			redacted: map[string]bool{
				"aws_token": true,
				"my_secret": true,
			},
			expected: []string{
				"aws_token",
				"my_secret",
			},
		},
		{
			name: "SuspiciousPatternsAndRedacted",
			projectVars: map[string]string{
				"not_suspicious": "",
				"num_hosts":      "",
				"aws_token":      "",
				"my_secret":      "",
				"svc_PASSWORD":   "",
			},
			redacted: map[string]bool{
				"aws_token": true,
				"my_secret": true,
			},
			expected: []string{
				"aws_token",
				"my_secret",
				"svc_PASSWORD",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.expected = append(test.expected, command.ExpansionsToRedact...)

			actual := redactList(test.projectVars, test.redacted)
			assert.ElementsMatch(t, test.expected, actual)
		})
	}
}
