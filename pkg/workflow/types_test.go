package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowStatus_Constants(t *testing.T) {
	assert.Equal(t, WorkflowStatus("draft"), WorkflowStatusDraft)
	assert.Equal(t, WorkflowStatus("active"), WorkflowStatusActive)
	assert.Equal(t, WorkflowStatus("inactive"), WorkflowStatusInactive)
	assert.Equal(t, WorkflowStatus("archived"), WorkflowStatusArchived)
}

func TestExecutionStatus_Constants(t *testing.T) {
	assert.Equal(t, ExecutionStatus("pending"), ExecutionStatusPending)
	assert.Equal(t, ExecutionStatus("running"), ExecutionStatusRunning)
	assert.Equal(t, ExecutionStatus("completed"), ExecutionStatusCompleted)
	assert.Equal(t, ExecutionStatus("failed"), ExecutionStatusFailed)
	assert.Equal(t, ExecutionStatus("cancelled"), ExecutionStatusCancelled)
	assert.Equal(t, ExecutionStatus("timeout"), ExecutionStatusTimeout)
}

func TestNodeStatus_Constants(t *testing.T) {
	assert.Equal(t, NodeStatus("pending"), NodeStatusPending)
	assert.Equal(t, NodeStatus("running"), NodeStatusRunning)
	assert.Equal(t, NodeStatus("completed"), NodeStatusCompleted)
	assert.Equal(t, NodeStatus("failed"), NodeStatusFailed)
	assert.Equal(t, NodeStatus("skipped"), NodeStatusSkipped)
}

func TestWorkflowDefinition_Creation(t *testing.T) {
	now := time.Now()
	workflow := &WorkflowDefinition{
		ID:          "wf-001",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Version:     1,
		Status:      WorkflowStatusActive,
		Nodes: []*NodeDefinition{
			{ID: "node-1", Name: "Start", Type: "trigger"},
			{ID: "node-2", Name: "Process", Type: "action"},
		},
		Edges: []*Edge{
			{ID: "edge-1", Source: "node-1", Target: "node-2"},
		},
		Variables: map[string]interface{}{
			"key1": "value1",
		},
		Settings: &WorkflowSettings{
			ExecutionTimeout: 30 * time.Second,
			MaxRetries:       3,
			RetryDelay:       5 * time.Second,
			OnError:          "stop",
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: "user-001",
	}

	assert.NotNil(t, workflow)
	assert.Equal(t, "wf-001", workflow.ID)
	assert.Equal(t, "Test Workflow", workflow.Name)
	assert.Equal(t, 1, workflow.Version)
	assert.Equal(t, WorkflowStatusActive, workflow.Status)
	assert.Len(t, workflow.Nodes, 2)
	assert.Len(t, workflow.Edges, 1)
	assert.NotNil(t, workflow.Settings)
}

func TestExecutionContext_Creation(t *testing.T) {
	now := time.Now()
	ctx := &ExecutionContext{
		ID:         "exec-001",
		WorkflowID: "wf-001",
		Status:     ExecutionStatusRunning,
		Input: map[string]interface{}{
			"userId": "user-123",
		},
		Output:        map[string]interface{}{},
		Variables:     map[string]interface{}{},
		NodeStates:    map[string]*NodeState{},
		CurrentNodeID: "node-1",
		StartedAt:     now,
		TriggerBy:     "user-001",
	}

	assert.NotNil(t, ctx)
	assert.Equal(t, "exec-001", ctx.ID)
	assert.Equal(t, "wf-001", ctx.WorkflowID)
	assert.Equal(t, ExecutionStatusRunning, ctx.Status)
	assert.Equal(t, "node-1", ctx.CurrentNodeID)
	assert.Nil(t, ctx.CompletedAt)
}

func TestExecutionMetrics_AtomicOperations(t *testing.T) {
	metrics := &ExecutionMetrics{}

	metrics.TotalExecutions.Add(1)
	metrics.SuccessExecutions.Add(1)
	metrics.FailedExecutions.Add(0)
	metrics.AverageDuration.Store(150)

	assert.Equal(t, int64(1), metrics.TotalExecutions.Load())
	assert.Equal(t, int64(1), metrics.SuccessExecutions.Load())
	assert.Equal(t, int64(0), metrics.FailedExecutions.Load())
	assert.Equal(t, int64(150), metrics.AverageDuration.Load())

	metrics.TotalExecutions.Add(2)
	metrics.FailedExecutions.Add(1)

	assert.Equal(t, int64(3), metrics.TotalExecutions.Load())
	assert.Equal(t, int64(1), metrics.FailedExecutions.Load())
}
