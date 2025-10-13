package workflow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrWorkflowNotFound", ErrWorkflowNotFound, "workflow not found"},
		{"ErrWorkflowAlreadyExists", ErrWorkflowAlreadyExists, "workflow already exists"},
		{"ErrWorkflowInvalidState", ErrWorkflowInvalidState, "workflow in invalid state"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.err, tt.expected)
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestExecutionErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrExecutionNotFound", ErrExecutionNotFound, "execution not found"},
		{"ErrExecutionAlreadyDone", ErrExecutionAlreadyDone, "execution already completed or failed"},
		{"ErrExecutionTimeout", ErrExecutionTimeout, "execution timeout"},
		{"ErrExecutionCancelled", ErrExecutionCancelled, "execution cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.EqualError(t, tt.err, tt.expected)
		})
	}
}

func TestNodeErrors(t *testing.T) {
	assert.NotNil(t, ErrNodeNotFound)
	assert.NotNil(t, ErrNodeTypeNotRegistered)
	assert.NotNil(t, ErrNodeExecutionFailed)
	assert.NotNil(t, ErrInvalidNodeConfig)
}

func TestAllErrorsDefined(t *testing.T) {
	allErrors := []error{
		ErrWorkflowNotFound,
		ErrWorkflowAlreadyExists,
		ErrWorkflowInvalidState,
		ErrExecutionNotFound,
		ErrExecutionAlreadyDone,
		ErrExecutionTimeout,
		ErrExecutionCancelled,
		ErrNodeNotFound,
		ErrNodeTypeNotRegistered,
		ErrNodeExecutionFailed,
		ErrInvalidNodeConfig,
		ErrInvalidEdge,
		ErrCyclicDependency,
		ErrDisconnectedGraph,
		ErrInvalidWorkflowDef,
		ErrMissingTriggerNode,
		ErrInvalidCondition,
		ErrPersistenceFailed,
	}

	for _, err := range allErrors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}
