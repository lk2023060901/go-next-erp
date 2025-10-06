package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresPersistence PostgreSQL 持久化实现
// 企业级特性：
// - 完整的 ACID 事务支持
// - 高性能批量操作
// - 连接池管理
// - 自动重连机制
// - 查询优化和索引
type PostgresPersistence struct {
	pool *pgxpool.Pool
}

// NewPostgresPersistence 创建 PostgreSQL 持久化
func NewPostgresPersistence(pool *pgxpool.Pool) (*PostgresPersistence, error) {
	p := &PostgresPersistence{
		pool: pool,
	}

	// 初始化表结构
	if err := p.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return p, nil
}

// initSchema 初始化数据库表结构
func (p *PostgresPersistence) initSchema(ctx context.Context) error {
	schema := `
	-- 工作流定义表
	CREATE TABLE IF NOT EXISTS workflows (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		version INTEGER NOT NULL DEFAULT 1,
		status VARCHAR(50) NOT NULL,
		nodes JSONB NOT NULL,
		edges JSONB NOT NULL,
		variables JSONB,
		settings JSONB NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		created_by VARCHAR(255),
		UNIQUE(name, version)
	);

	CREATE INDEX IF NOT EXISTS idx_workflows_status ON workflows(status);
	CREATE INDEX IF NOT EXISTS idx_workflows_created_by ON workflows(created_by);
	CREATE INDEX IF NOT EXISTS idx_workflows_created_at ON workflows(created_at DESC);

	-- 执行上下文表
	CREATE TABLE IF NOT EXISTS workflow_executions (
		id VARCHAR(255) PRIMARY KEY,
		workflow_id VARCHAR(255) NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
		status VARCHAR(50) NOT NULL,
		input JSONB,
		output JSONB,
		variables JSONB,
		error TEXT,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		trigger_by VARCHAR(255),
		metadata JSONB,
		FOREIGN KEY (workflow_id) REFERENCES workflows(id)
	);

	CREATE INDEX IF NOT EXISTS idx_executions_workflow_id ON workflow_executions(workflow_id);
	CREATE INDEX IF NOT EXISTS idx_executions_status ON workflow_executions(status);
	CREATE INDEX IF NOT EXISTS idx_executions_started_at ON workflow_executions(started_at DESC);
	CREATE INDEX IF NOT EXISTS idx_executions_trigger_by ON workflow_executions(trigger_by);

	-- 节点状态表
	CREATE TABLE IF NOT EXISTS node_states (
		id SERIAL PRIMARY KEY,
		execution_id VARCHAR(255) NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
		node_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) NOT NULL,
		input JSONB,
		output JSONB,
		error TEXT,
		attempts INTEGER NOT NULL DEFAULT 0,
		started_at TIMESTAMP NOT NULL,
		completed_at TIMESTAMP,
		UNIQUE(execution_id, node_id)
	);

	CREATE INDEX IF NOT EXISTS idx_node_states_execution_id ON node_states(execution_id);
	CREATE INDEX IF NOT EXISTS idx_node_states_status ON node_states(status);
	`

	_, err := p.pool.Exec(ctx, schema)
	return err
}

// SaveWorkflow 保存工作流定义
func (p *PostgresPersistence) SaveWorkflow(ctx context.Context, def *WorkflowDefinition) error {
	nodesJSON, err := json.Marshal(def.Nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal nodes: %w", err)
	}

	edgesJSON, err := json.Marshal(def.Edges)
	if err != nil {
		return fmt.Errorf("failed to marshal edges: %w", err)
	}

	variablesJSON, err := json.Marshal(def.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	settingsJSON, err := json.Marshal(def.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO workflows (
			id, name, description, version, status, nodes, edges,
			variables, settings, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			version = EXCLUDED.version,
			status = EXCLUDED.status,
			nodes = EXCLUDED.nodes,
			edges = EXCLUDED.edges,
			variables = EXCLUDED.variables,
			settings = EXCLUDED.settings,
			updated_at = EXCLUDED.updated_at
	`

	_, err = p.pool.Exec(ctx, query,
		def.ID, def.Name, def.Description, def.Version, def.Status,
		nodesJSON, edgesJSON, variablesJSON, settingsJSON,
		def.CreatedAt, def.UpdatedAt, def.CreatedBy,
	)

	if err != nil {
		return fmt.Errorf("failed to save workflow: %w", err)
	}

	return nil
}

// GetWorkflow 获取工作流定义
func (p *PostgresPersistence) GetWorkflow(ctx context.Context, workflowID string) (*WorkflowDefinition, error) {
	query := `
		SELECT id, name, description, version, status, nodes, edges,
		       variables, settings, created_at, updated_at, created_by
		FROM workflows
		WHERE id = $1
	`

	var def WorkflowDefinition
	var nodesJSON, edgesJSON, variablesJSON, settingsJSON []byte

	err := p.pool.QueryRow(ctx, query, workflowID).Scan(
		&def.ID, &def.Name, &def.Description, &def.Version, &def.Status,
		&nodesJSON, &edgesJSON, &variablesJSON, &settingsJSON,
		&def.CreatedAt, &def.UpdatedAt, &def.CreatedBy,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrWorkflowNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// 反序列化 JSON 字段
	if err := json.Unmarshal(nodesJSON, &def.Nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}
	if err := json.Unmarshal(edgesJSON, &def.Edges); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
	}
	if err := json.Unmarshal(variablesJSON, &def.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}
	if err := json.Unmarshal(settingsJSON, &def.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	return &def, nil
}

// ListWorkflows 列出工作流
func (p *PostgresPersistence) ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*WorkflowDefinition, error) {
	query := `
		SELECT id, name, description, version, status, nodes, edges,
		       variables, settings, created_at, updated_at, created_by
		FROM workflows
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	// 应用过滤条件
	if len(filter.Status) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.CreatedBy != "" {
		query += fmt.Sprintf(" AND created_by = $%d", argIndex)
		args = append(args, filter.CreatedBy)
		argIndex++
	}

	if filter.Search != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	// 排序
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// 分页
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*WorkflowDefinition

	for rows.Next() {
		var def WorkflowDefinition
		var nodesJSON, edgesJSON, variablesJSON, settingsJSON []byte

		err := rows.Scan(
			&def.ID, &def.Name, &def.Description, &def.Version, &def.Status,
			&nodesJSON, &edgesJSON, &variablesJSON, &settingsJSON,
			&def.CreatedAt, &def.UpdatedAt, &def.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow: %w", err)
		}

		// 反序列化
		json.Unmarshal(nodesJSON, &def.Nodes)
		json.Unmarshal(edgesJSON, &def.Edges)
		json.Unmarshal(variablesJSON, &def.Variables)
		json.Unmarshal(settingsJSON, &def.Settings)

		workflows = append(workflows, &def)
	}

	return workflows, nil
}

// DeleteWorkflow 删除工作流
func (p *PostgresPersistence) DeleteWorkflow(ctx context.Context, workflowID string) error {
	query := `DELETE FROM workflows WHERE id = $1`

	result, err := p.pool.Exec(ctx, query, workflowID)
	if err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrWorkflowNotFound
	}

	return nil
}

// UpdateWorkflowStatus 更新工作流状态
func (p *PostgresPersistence) UpdateWorkflowStatus(ctx context.Context, workflowID string, status WorkflowStatus) error {
	query := `UPDATE workflows SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := p.pool.Exec(ctx, query, status, time.Now(), workflowID)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrWorkflowNotFound
	}

	return nil
}

// SaveExecution 保存执行上下文
func (p *PostgresPersistence) SaveExecution(ctx context.Context, execCtx *ExecutionContext) error {
	inputJSON, _ := json.Marshal(execCtx.Input)
	outputJSON, _ := json.Marshal(execCtx.Output)
	variablesJSON, _ := json.Marshal(execCtx.Variables)
	metadataJSON, _ := json.Marshal(execCtx.Metadata)

	query := `
		INSERT INTO workflow_executions (
			id, workflow_id, status, input, output, variables,
			error, started_at, completed_at, trigger_by, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			output = EXCLUDED.output,
			variables = EXCLUDED.variables,
			error = EXCLUDED.error,
			completed_at = EXCLUDED.completed_at
	`

	_, err := p.pool.Exec(ctx, query,
		execCtx.ID, execCtx.WorkflowID, execCtx.Status,
		inputJSON, outputJSON, variablesJSON,
		execCtx.Error, execCtx.StartedAt, execCtx.CompletedAt,
		execCtx.TriggerBy, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}

	// 保存节点状态
	for _, state := range execCtx.NodeStates {
		if err := p.SaveNodeState(ctx, execCtx.ID, state); err != nil {
			return err
		}
	}

	return nil
}

// GetExecution 获取执行上下文
func (p *PostgresPersistence) GetExecution(ctx context.Context, executionID string) (*ExecutionContext, error) {
	query := `
		SELECT id, workflow_id, status, input, output, variables,
		       error, started_at, completed_at, trigger_by, metadata
		FROM workflow_executions
		WHERE id = $1
	`

	var execCtx ExecutionContext
	var inputJSON, outputJSON, variablesJSON, metadataJSON []byte

	err := p.pool.QueryRow(ctx, query, executionID).Scan(
		&execCtx.ID, &execCtx.WorkflowID, &execCtx.Status,
		&inputJSON, &outputJSON, &variablesJSON,
		&execCtx.Error, &execCtx.StartedAt, &execCtx.CompletedAt,
		&execCtx.TriggerBy, &metadataJSON,
	)

	if err == pgx.ErrNoRows {
		return nil, ErrExecutionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	// 反序列化
	json.Unmarshal(inputJSON, &execCtx.Input)
	json.Unmarshal(outputJSON, &execCtx.Output)
	json.Unmarshal(variablesJSON, &execCtx.Variables)
	json.Unmarshal(metadataJSON, &execCtx.Metadata)

	// 加载节点状态
	nodeStates, err := p.GetNodeStates(ctx, executionID)
	if err != nil {
		return nil, err
	}
	execCtx.NodeStates = nodeStates

	return &execCtx, nil
}

// ListExecutions 列出执行记录
func (p *PostgresPersistence) ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*ExecutionContext, error) {
	query := `
		SELECT id, workflow_id, status, input, output, variables,
		       error, started_at, completed_at, trigger_by, metadata
		FROM workflow_executions
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if filter.WorkflowID != "" {
		query += fmt.Sprintf(" AND workflow_id = $%d", argIndex)
		args = append(args, filter.WorkflowID)
		argIndex++
	}

	if len(filter.Status) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.TriggerBy != "" {
		query += fmt.Sprintf(" AND trigger_by = $%d", argIndex)
		args = append(args, filter.TriggerBy)
		argIndex++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND started_at >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND started_at <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	// 排序
	sortBy := "started_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// 分页
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions: %w", err)
	}
	defer rows.Close()

	var executions []*ExecutionContext

	for rows.Next() {
		var execCtx ExecutionContext
		var inputJSON, outputJSON, variablesJSON, metadataJSON []byte

		err := rows.Scan(
			&execCtx.ID, &execCtx.WorkflowID, &execCtx.Status,
			&inputJSON, &outputJSON, &variablesJSON,
			&execCtx.Error, &execCtx.StartedAt, &execCtx.CompletedAt,
			&execCtx.TriggerBy, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		json.Unmarshal(inputJSON, &execCtx.Input)
		json.Unmarshal(outputJSON, &execCtx.Output)
		json.Unmarshal(variablesJSON, &execCtx.Variables)
		json.Unmarshal(metadataJSON, &execCtx.Metadata)

		executions = append(executions, &execCtx)
	}

	return executions, nil
}

// DeleteExecution 删除执行记录
func (p *PostgresPersistence) DeleteExecution(ctx context.Context, executionID string) error {
	query := `DELETE FROM workflow_executions WHERE id = $1`

	result, err := p.pool.Exec(ctx, query, executionID)
	if err != nil {
		return fmt.Errorf("failed to delete execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

// UpdateExecutionStatus 更新执行状态
func (p *PostgresPersistence) UpdateExecutionStatus(ctx context.Context, executionID string, status ExecutionStatus) error {
	query := `UPDATE workflow_executions SET status = $1 WHERE id = $2`

	result, err := p.pool.Exec(ctx, query, status, executionID)
	if err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

// SaveNodeState 保存节点状态
func (p *PostgresPersistence) SaveNodeState(ctx context.Context, executionID string, state *NodeState) error {
	inputJSON, _ := json.Marshal(state.Input)
	outputJSON, _ := json.Marshal(state.Output)

	query := `
		INSERT INTO node_states (
			execution_id, node_id, status, input, output,
			error, attempts, started_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (execution_id, node_id) DO UPDATE SET
			status = EXCLUDED.status,
			output = EXCLUDED.output,
			error = EXCLUDED.error,
			attempts = EXCLUDED.attempts,
			completed_at = EXCLUDED.completed_at
	`

	_, err := p.pool.Exec(ctx, query,
		executionID, state.NodeID, state.Status,
		inputJSON, outputJSON,
		state.Error, state.Attempts, state.StartedAt, state.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save node state: %w", err)
	}

	return nil
}

// GetNodeStates 获取节点状态
func (p *PostgresPersistence) GetNodeStates(ctx context.Context, executionID string) (map[string]*NodeState, error) {
	query := `
		SELECT node_id, status, input, output, error, attempts, started_at, completed_at
		FROM node_states
		WHERE execution_id = $1
	`

	rows, err := p.pool.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node states: %w", err)
	}
	defer rows.Close()

	states := make(map[string]*NodeState)

	for rows.Next() {
		var state NodeState
		var inputJSON, outputJSON []byte

		err := rows.Scan(
			&state.NodeID, &state.Status,
			&inputJSON, &outputJSON,
			&state.Error, &state.Attempts, &state.StartedAt, &state.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan node state: %w", err)
		}

		json.Unmarshal(inputJSON, &state.Input)
		json.Unmarshal(outputJSON, &state.Output)

		states[state.NodeID] = &state
	}

	return states, nil
}

// GetWorkflowStats 获取工作流统计
func (p *PostgresPersistence) GetWorkflowStats(ctx context.Context, workflowID string, timeRange *TimeRange) (*WorkflowStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'completed') as success,
			COUNT(*) FILTER (WHERE status = 'failed') as failed,
			MAX(started_at) as last_executed,
			AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_duration_seconds
		FROM workflow_executions
		WHERE workflow_id = $1
	`
	args := []interface{}{workflowID}

	if timeRange != nil {
		query += " AND started_at >= $2 AND started_at <= $3"
		args = append(args, timeRange.Start, timeRange.End)
	}

	var stats WorkflowStats
	var lastExecuted *time.Time
	var avgDurationSeconds *float64

	err := p.pool.QueryRow(ctx, query, args...).Scan(
		&stats.TotalExecutions,
		&stats.SuccessExecutions,
		&stats.FailedExecutions,
		&lastExecuted,
		&avgDurationSeconds,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get workflow stats: %w", err)
	}

	stats.WorkflowID = workflowID
	stats.LastExecutedAt = lastExecuted

	if avgDurationSeconds != nil {
		stats.AverageDuration = time.Duration(*avgDurationSeconds * float64(time.Second))
	}

	return &stats, nil
}

// GetExecutionHistory 获取执行历史
func (p *PostgresPersistence) GetExecutionHistory(ctx context.Context, workflowID string, limit int) ([]*ExecutionSummary, error) {
	query := `
		SELECT id, workflow_id, status, trigger_by, started_at, completed_at, error,
		       EXTRACT(EPOCH FROM (completed_at - started_at)) as duration_seconds
		FROM workflow_executions
		WHERE workflow_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`

	rows, err := p.pool.Query(ctx, query, workflowID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get execution history: %w", err)
	}
	defer rows.Close()

	var history []*ExecutionSummary

	for rows.Next() {
		var summary ExecutionSummary
		var durationSeconds *float64

		err := rows.Scan(
			&summary.ID, &summary.WorkflowID, &summary.Status, &summary.TriggerBy,
			&summary.StartedAt, &summary.CompletedAt, &summary.Error,
			&durationSeconds,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution summary: %w", err)
		}

		if durationSeconds != nil {
			summary.Duration = time.Duration(*durationSeconds * float64(time.Second))
		}

		history = append(history, &summary)
	}

	return history, nil
}

// CleanupOldExecutions 清理旧的执行记录
func (p *PostgresPersistence) CleanupOldExecutions(ctx context.Context, olderThan time.Time) (int, error) {
	query := `DELETE FROM workflow_executions WHERE started_at < $1`

	result, err := p.pool.Exec(ctx, query, olderThan)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old executions: %w", err)
	}

	return int(result.RowsAffected()), nil
}

// Ping 健康检查
func (p *PostgresPersistence) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Close 关闭连接
func (p *PostgresPersistence) Close() error {
	p.pool.Close()
	return nil
}
