package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew 测试引擎创建
func TestNew(t *testing.T) {
	t.Run("Create with default config", func(t *testing.T) {
		engine, err := New()
		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.NotNil(t, engine.config)
		assert.NotNil(t, engine.logger)
		assert.NotNil(t, engine.registry)
		assert.NotNil(t, engine.evaluator)
		assert.NotNil(t, engine.ctxMgr)
		assert.NotNil(t, engine.executor)
		assert.NotNil(t, engine.persistence)
	})

	t.Run("Create with custom config", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.MaxConcurrentExecutions = 50
		cfg.DefaultExecutionTimeout = time.Minute * 10
		cfg.EnableMetrics = true
		cfg.EnableTracing = true

		engine, err := New(WithConfig(cfg))
		require.NoError(t, err)
		assert.Equal(t, 50, engine.config.MaxConcurrentExecutions)
		assert.Equal(t, time.Minute*10, engine.config.DefaultExecutionTimeout)
	})
}

// TestRegisterNodeType 测试节点类型注册
func TestRegisterNodeType(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	t.Run("Register custom node type", func(t *testing.T) {
		factory := func(cfg *NodeDefinition) (Node, error) {
			return &mockNode{}, nil
		}
		err := engine.RegisterNodeType("custom", factory)
		assert.NoError(t, err)
	})

	t.Run("Register duplicate node type", func(t *testing.T) {
		factory := func(cfg *NodeDefinition) (Node, error) {
			return &mockNode{}, nil
		}
		err := engine.RegisterNodeType("duplicate", factory)
		require.NoError(t, err)

		err = engine.RegisterNodeType("duplicate", factory)
		assert.Error(t, err)
	})
}

// TestCreateWorkflow 测试创建工作流
func TestCreateWorkflow(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)
	registerTestNodes(t, engine)

	t.Run("Create valid workflow", func(t *testing.T) {
		def := &WorkflowDefinition{
			ID:      uuid.New().String(),
			Name:    "Test Workflow",
			Version: 1,
			Nodes: []*NodeDefinition{
				{
					ID:   "start",
					Type: "start",
					Name: "Start",
				},
			},
			Variables: map[string]interface{}{
				"test": "value",
			},
		}

		err := engine.CreateWorkflow(def)
		assert.NoError(t, err)
	})

	t.Run("Create workflow with duplicate ID", func(t *testing.T) {
		id := uuid.New().String()
		def1 := &WorkflowDefinition{
			ID:      id,
			Name:    "Test 1",
			Version: 1,
			Nodes: []*NodeDefinition{
				{
					ID:   "start",
					Type: "start",
					Name: "Start",
				},
			},
		}

		err := engine.CreateWorkflow(def1)
		require.NoError(t, err)

		def2 := &WorkflowDefinition{
			ID:      id,
			Name:    "Test 2",
			Version: 1,
			Nodes: []*NodeDefinition{
				{
					ID:   "start",
					Type: "start",
					Name: "Start",
				},
			},
		}

		err = engine.CreateWorkflow(def2)
		assert.ErrorIs(t, err, ErrWorkflowAlreadyExists)
	})

	t.Run("Create workflow with empty ID", func(t *testing.T) {
		def := &WorkflowDefinition{
			ID:      "",
			Name:    "No ID",
			Version: 1,
			Nodes:   []*NodeDefinition{},
		}

		err := engine.CreateWorkflow(def)
		assert.Error(t, err)
	})

	t.Run("Create workflow with no nodes", func(t *testing.T) {
		def := &WorkflowDefinition{
			ID:      uuid.New().String(),
			Name:    "No Nodes",
			Version: 1,
			Nodes:   []*NodeDefinition{},
		}

		err := engine.CreateWorkflow(def)
		assert.Error(t, err)
	})
}

// TestGetWorkflow 测试获取工作流
func TestGetWorkflow(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)
	registerTestNodes(t, engine)

	def := &WorkflowDefinition{
		ID:      uuid.New().String(),
		Name:    "Test Workflow",
		Version: 1,
		Nodes: []*NodeDefinition{
			{
				ID:   "start",
				Type: "start",
				Name: "Start",
			},
		},
	}

	err = engine.CreateWorkflow(def)
	require.NoError(t, err)

	t.Run("Get existing workflow", func(t *testing.T) {
		found, err := engine.GetWorkflow(def.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, def.ID, found.ID)
		assert.Equal(t, def.Name, found.Name)
	})

	t.Run("Get non-existent workflow", func(t *testing.T) {
		_, err := engine.GetWorkflow("non-existent")
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})
}

// TestUpdateWorkflow 测试更新工作流
func TestUpdateWorkflow(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)
	registerTestNodes(t, engine)
	require.NoError(t, err)

	id := uuid.New().String()
	def := &WorkflowDefinition{
		ID:      id,
		Name:    "Original",
		Version: 1,
		Nodes: []*NodeDefinition{
			{
				ID:   "start",
				Type: "start",
				Name: "Start",
			},
		},
	}

	err = engine.CreateWorkflow(def)
	require.NoError(t, err)

	t.Run("Update existing workflow", func(t *testing.T) {
		updated := &WorkflowDefinition{
			ID:      id,
			Name:    "Updated",
			Version: 2,
			Nodes: []*NodeDefinition{
				{
					ID:   "start",
					Type: "start",
					Name: "Start",
				},
			},
		}

		err := engine.UpdateWorkflow(updated)
		assert.NoError(t, err)

		found, err := engine.GetWorkflow(id)
		require.NoError(t, err)
		assert.Equal(t, "Updated", found.Name)
		assert.Equal(t, 2, found.Version)
	})

	t.Run("Update non-existent workflow", func(t *testing.T) {
		nonExistent := &WorkflowDefinition{
			ID:      "non-existent",
			Name:    "Does Not Exist",
			Version: 1,
			Nodes: []*NodeDefinition{
				{
					ID:   "start",
					Type: "start",
					Name: "Start",
				},
			},
		}

		err := engine.UpdateWorkflow(nonExistent)
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})
}

// TestDeleteWorkflow 测试删除工作流
func TestDeleteWorkflow(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)
	registerTestNodes(t, engine)

	def := &WorkflowDefinition{
		ID:      uuid.New().String(),
		Name:    "To Delete",
		Version: 1,
		Nodes: []*NodeDefinition{
			{
				ID:   "start",
				Type: "start",
				Name: "Start",
			},
		},
	}

	err = engine.CreateWorkflow(def)
	require.NoError(t, err)

	t.Run("Delete existing workflow", func(t *testing.T) {
		err := engine.DeleteWorkflow(def.ID)
		assert.NoError(t, err)

		_, err = engine.GetWorkflow(def.ID)
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})

	t.Run("Delete non-existent workflow", func(t *testing.T) {
		err := engine.DeleteWorkflow("non-existent")
		assert.ErrorIs(t, err, ErrWorkflowNotFound)
	})
}

// TestListWorkflows 测试列出工作流
func TestListWorkflows(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	t.Run("List empty workflows", func(t *testing.T) {
		workflows := engine.ListWorkflows()
		assert.Empty(t, workflows)
	})

	t.Run("List multiple workflows", func(t *testing.T) {
		engine2, err := New()
		require.NoError(t, err)
		registerTestNodes(t, engine2)

		for i := 0; i < 3; i++ {
			def := &WorkflowDefinition{
				ID:      uuid.New().String(),
				Name:    "Workflow " + string(rune('A'+i)),
				Version: 1,
				Nodes: []*NodeDefinition{
					{
						ID:   "start",
						Type: "start",
						Name: "Start",
					},
				},
			}
			err := engine2.CreateWorkflow(def)
			require.NoError(t, err)
		}

		workflows := engine2.ListWorkflows()
		assert.Len(t, workflows, 3)
	})
}

// TestEngineStructure 测试引擎结构
func TestEngineStructure(t *testing.T) {
	engine, err := New()
	require.NoError(t, err)

	t.Run("Middleware slice exists", func(t *testing.T) {
		// Just verify the middlewares slice exists
		assert.NotNil(t, engine.middlewares)
	})

	t.Run("Metrics map exists", func(t *testing.T) {
		// Verify metrics map exists
		assert.NotNil(t, engine.metrics)
	})
}

// registerTestNodes 注册测试用节点类型
func registerTestNodes(t *testing.T, engine *Engine) {
	t.Helper()
	factory := func(cfg *NodeDefinition) (Node, error) {
		return &mockNode{}, nil
	}
	err := engine.RegisterNodeType("start", factory)
	require.NoError(t, err)
}

// mockNode 用于测试的模拟节点
type mockNode struct{}

func (m *mockNode) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	return input, nil
}

func (m *mockNode) Type() string {
	return "mock"
}

func (m *mockNode) Validate() error {
	return nil
}
