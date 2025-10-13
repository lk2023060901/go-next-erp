package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lk2023060901/go-next-erp/internal/approval/dto"
	"github.com/lk2023060901/go-next-erp/internal/approval/model"
)

// DeleteProcessDefinition 删除流程定义
func (s *approvalService) DeleteProcessDefinition(ctx context.Context, id uuid.UUID) error {
	_, err := s.processDefRepo.FindByID(ctx, id)
	if err != nil {
		return ErrProcessNotFound
	}

	// 检查是否有活跃实例
	instances, err := s.processInstRepo.ListByProcessDef(ctx, id, 10, 0)
	if err != nil {
		return fmt.Errorf("failed to check instances: %w", err)
	}

	for _, inst := range instances {
		if inst.Status == model.ProcessStatusPending {
			return ErrProcessHasInstances
		}
	}

	return s.processDefRepo.Delete(ctx, id)
}

// SetProcessDefinitionStatus 设置流程定义状态
func (s *approvalService) SetProcessDefinitionStatus(ctx context.Context, id uuid.UUID, enabled bool) error {
	processDef, err := s.processDefRepo.FindByID(ctx, id)
	if err != nil {
		return ErrProcessNotFound
	}

	processDef.Enabled = enabled
	processDef.UpdatedAt = time.Now()
	return s.processDefRepo.Update(ctx, processDef)
}

// GetProcessStats 获取流程统计
func (s *approvalService) GetProcessStats(ctx context.Context, processDefID uuid.UUID) (*dto.ProcessStatsResponse, error) {
	processDef, err := s.processDefRepo.FindByID(ctx, processDefID)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	instances, err := s.processInstRepo.ListByProcessDef(ctx, processDefID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}

	stats := &dto.ProcessStatsResponse{
		ProcessDefID: processDef.ID,
		ProcessCode:  processDef.Code,
		ProcessName:  processDef.Name,
	}

	var totalDuration int64
	completedCount := 0

	for _, inst := range instances {
		stats.TotalInstances++
		switch inst.Status {
		case model.ProcessStatusPending:
			stats.PendingInstances++
		case model.ProcessStatusApproved:
			stats.ApprovedInstances++
		case model.ProcessStatusRejected:
			stats.RejectedInstances++
		}

		if inst.CompletedAt != nil {
			duration := inst.CompletedAt.Sub(inst.StartedAt).Seconds()
			totalDuration += int64(duration)
			completedCount++
		}
	}

	if completedCount > 0 {
		stats.AvgDuration = totalDuration / int64(completedCount)
	}

	return stats, nil
}

// ListProcessInstances 列出流程实例（带筛选）
func (s *approvalService) ListProcessInstances(ctx context.Context, tenantID uuid.UUID, processDefID *uuid.UUID, status *model.ProcessStatus, applicantID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*dto.ProcessInstanceResponse, int, error) {
	// 简化实现：根据不同条件调用不同的 repository 方法
	var instances []*model.ProcessInstance
	var err error

	if processDefID != nil {
		instances, err = s.processInstRepo.ListByProcessDef(ctx, *processDefID, limit*10, 0)
	} else if status != nil {
		instances, err = s.processInstRepo.ListByStatus(ctx, tenantID, *status, limit*10, 0)
	} else {
		// 默认返回空列表（实际应该实现通用查询方法）
		instances = []*model.ProcessInstance{}
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list instances: %w", err)
	}

	// 内存过滤（简化实现）
	filtered := make([]*model.ProcessInstance, 0)
	for _, inst := range instances {
		if applicantID != nil && inst.ApplicantID != *applicantID {
			continue
		}
		if status != nil && inst.Status != *status {
			continue
		}
		if startDate != nil && inst.StartedAt.Before(*startDate) {
			continue
		}
		if endDate != nil && inst.StartedAt.After(*endDate) {
			continue
		}
		filtered = append(filtered, inst)
	}

	total := len(filtered)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	result := make([]*dto.ProcessInstanceResponse, 0)
	for i := start; i < end; i++ {
		result = append(result, dto.ToProcessInstanceResponse(filtered[i]))
	}

	return result, total, nil
}

// CancelProcess 取消流程
func (s *approvalService) CancelProcess(ctx context.Context, instanceID uuid.UUID, operatorID uuid.UUID, reason *string) error {
	instance, err := s.processInstRepo.FindByID(ctx, instanceID)
	if err != nil {
		return ErrProcessInstanceNotFound
	}

	if instance.ApplicantID != operatorID {
		return ErrUnauthorized
	}

	if instance.Status != model.ProcessStatusPending {
		return fmt.Errorf("cannot cancel process in %s status", instance.Status)
	}

	instance.Status = model.ProcessStatusCancelled
	now := time.Now()
	instance.CompletedAt = &now
	return s.processInstRepo.Update(ctx, instance)
}

// GetInstanceStatsSummary 获取实例统计汇总
func (s *approvalService) GetInstanceStatsSummary(ctx context.Context, tenantID uuid.UUID) (*dto.InstanceStatsSummary, error) {
	stats := &dto.InstanceStatsSummary{
		ByStatus: make(map[string]int),
	}

	// 按状态分别统计
	statuses := []model.ProcessStatus{
		model.ProcessStatusPending,
		model.ProcessStatusApproved,
		model.ProcessStatusRejected,
		model.ProcessStatusWithdrawn,
		model.ProcessStatusCancelled,
	}

	for _, status := range statuses {
		count, err := s.processInstRepo.CountByStatus(ctx, tenantID, status)
		if err != nil {
			continue
		}
		stats.Total += count
		stats.ByStatus[string(status)] = count

		switch status {
		case model.ProcessStatusPending:
			stats.Pending = count
		case model.ProcessStatusApproved:
			stats.Approved = count
		case model.ProcessStatusRejected:
			stats.Rejected = count
		case model.ProcessStatusWithdrawn:
			stats.Withdrawn = count
		case model.ProcessStatusCancelled:
			stats.Cancelled = count
		}
	}

	return stats, nil
}

// GetInstanceStatsByStatus 获取按状态分组的实例统计
func (s *approvalService) GetInstanceStatsByStatus(ctx context.Context, tenantID uuid.UUID, processDefID *uuid.UUID, startDate, endDate *time.Time) (map[string]int, error) {
	stats := make(map[string]int)

	statuses := []model.ProcessStatus{
		model.ProcessStatusPending,
		model.ProcessStatusApproved,
		model.ProcessStatusRejected,
		model.ProcessStatusWithdrawn,
		model.ProcessStatusCancelled,
	}

	for _, status := range statuses {
		count, err := s.processInstRepo.CountByStatus(ctx, tenantID, status)
		if err != nil {
			continue
		}
		stats[string(status)] = count
	}

	return stats, nil
}

// BatchProcessTasks 批量处理审批任务
func (s *approvalService) BatchProcessTasks(ctx context.Context, taskIDs []uuid.UUID, operatorID uuid.UUID, action model.ApprovalAction, comment *string) ([]*dto.BatchProcessResult, error) {
	results := make([]*dto.BatchProcessResult, 0, len(taskIDs))

	for _, taskID := range taskIDs {
		result := &dto.BatchProcessResult{
			TaskID:  taskID,
			Success: true,
		}

		req := &dto.ProcessTaskRequest{
			TaskID:     taskID,
			OperatorID: operatorID,
			Action:     action,
			Comment:    comment,
		}

		if err := s.ProcessTask(ctx, req); err != nil {
			result.Success = false
			errMsg := err.Error()
			result.Error = &errMsg
		}

		results = append(results, result)
	}

	return results, nil
}

// TransferTask 转审任务
func (s *approvalService) TransferTask(ctx context.Context, taskID uuid.UUID, fromUserID, toUserID uuid.UUID, comment *string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return ErrTaskNotFound
	}

	if task.AssigneeID != fromUserID {
		return ErrUnauthorized
	}

	if task.Status != model.TaskStatusPending {
		return ErrTaskAlreadyProcessed
	}

	task.AssigneeID = toUserID
	task.TransferToID = &toUserID
	task.UpdatedAt = time.Now()
	return s.taskRepo.Update(ctx, task)
}

// DelegateTask 委托任务
func (s *approvalService) DelegateTask(ctx context.Context, taskID uuid.UUID, fromUserID, toUserID uuid.UUID, comment *string) error {
	return s.TransferTask(ctx, taskID, fromUserID, toUserID, comment)
}

// ListPendingTasks 列出待处理任务
func (s *approvalService) ListPendingTasks(ctx context.Context, tenantID uuid.UUID, processDefID, assigneeID *uuid.UUID, limit, offset int) ([]*dto.ApprovalTaskResponse, int, error) {
	var tasks []*model.ApprovalTask
	var err error

	if assigneeID != nil {
		pendingStatus := model.TaskStatusPending
		tasks, err = s.taskRepo.ListByAssignee(ctx, *assigneeID, &pendingStatus, limit*10, 0)
	} else {
		// 返回空列表（简化实现）
		tasks = []*model.ApprovalTask{}
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	total := len(tasks)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	result := make([]*dto.ApprovalTaskResponse, 0)
	for i := start; i < end; i++ {
		result = append(result, dto.ToApprovalTaskResponse(tasks[i]))
	}

	return result, total, nil
}

// ListCompletedTasks 列出已完成任务
func (s *approvalService) ListCompletedTasks(ctx context.Context, tenantID uuid.UUID, processDefID, assigneeID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]*dto.ApprovalTaskResponse, int, error) {
	var tasks []*model.ApprovalTask
	var err error

	if assigneeID != nil {
		tasks, err = s.taskRepo.ListByAssignee(ctx, *assigneeID, nil, limit*10, 0)
	} else {
		tasks = []*model.ApprovalTask{}
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	// 过滤已完成的任务
	filtered := make([]*model.ApprovalTask, 0)
	for _, task := range tasks {
		if task.Status == model.TaskStatusPending {
			continue
		}
		if task.ApprovedAt != nil {
			if startDate != nil && task.ApprovedAt.Before(*startDate) {
				continue
			}
			if endDate != nil && task.ApprovedAt.After(*endDate) {
				continue
			}
		}
		filtered = append(filtered, task)
	}

	total := len(filtered)
	start := offset
	end := offset + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	result := make([]*dto.ApprovalTaskResponse, 0)
	for i := start; i < end; i++ {
		result = append(result, dto.ToApprovalTaskResponse(filtered[i]))
	}

	return result, total, nil
}

// GetTaskHistory 获取任务历史
func (s *approvalService) GetTaskHistory(ctx context.Context, taskID uuid.UUID) ([]*dto.ProcessHistoryResponse, error) {
	histories, err := s.historyRepo.ListByTaskID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task history: %w", err)
	}
	return dto.ToProcessHistoryResponseList(histories), nil
}

// GetDashboard 获取审批工作台数据
func (s *approvalService) GetDashboard(ctx context.Context, tenantID, userID uuid.UUID) (*dto.DashboardResponse, error) {
	pendingCount, _ := s.CountPendingTasks(ctx, userID)

	completedStatus := model.TaskStatusApproved
	completedTasks, _ := s.taskRepo.ListByAssignee(ctx, userID, &completedStatus, 100, 0)
	completedCount := len(completedTasks)

	myApps, _ := s.ListMyApplications(ctx, userID, 100, 0)
	appsCount := len(myApps)

	pendingApps := 0
	for _, app := range myApps {
		if app.Status == model.ProcessStatusPending {
			pendingApps++
		}
	}

	recentTasks, _ := s.ListMyTasks(ctx, userID, nil, 5, 0)
	recentApps := myApps
	if len(recentApps) > 5 {
		recentApps = recentApps[:5]
	}

	return &dto.DashboardResponse{
		MyPendingTasks:      pendingCount,
		MyCompletedTasks:    completedCount,
		MyApplications:      appsCount,
		PendingApplications: pendingApps,
		RecentTasks:         recentTasks,
		RecentApplications:  recentApps,
	}, nil
}

// GetProcessMetrics 获取流程性能指标
func (s *approvalService) GetProcessMetrics(ctx context.Context, processDefID uuid.UUID, startDate, endDate *time.Time) (*dto.ProcessMetrics, error) {
	processDef, err := s.processDefRepo.FindByID(ctx, processDefID)
	if err != nil {
		return nil, ErrProcessNotFound
	}

	instances, err := s.processInstRepo.ListByProcessDef(ctx, processDefID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}

	metrics := &dto.ProcessMetrics{
		ProcessDefID: processDef.ID,
		ProcessCode:  processDef.Code,
		ProcessName:  processDef.Name,
	}

	var totalDuration, maxDuration, minDuration int64
	minDuration = -1

	for _, inst := range instances {
		if startDate != nil && inst.StartedAt.Before(*startDate) {
			continue
		}
		if endDate != nil && inst.StartedAt.After(*endDate) {
			continue
		}

		metrics.TotalInstances++

		if inst.CompletedAt != nil {
			metrics.CompletedInstances++
			duration := int64(inst.CompletedAt.Sub(inst.StartedAt).Seconds())
			totalDuration += duration

			if duration > maxDuration {
				maxDuration = duration
			}
			if minDuration == -1 || duration < minDuration {
				minDuration = duration
			}

			switch inst.Status {
			case model.ProcessStatusApproved:
				metrics.ApprovedInstances++
			case model.ProcessStatusRejected:
				metrics.RejectedInstances++
			}
		} else if inst.Status == model.ProcessStatusPending {
			metrics.PendingInstances++
		}
	}

	if metrics.CompletedInstances > 0 {
		metrics.AvgDurationSeconds = totalDuration / int64(metrics.CompletedInstances)
		metrics.AvgApprovalTime = metrics.AvgDurationSeconds
		metrics.MaxDuration = maxDuration
		if minDuration != -1 {
			metrics.MinDuration = minDuration
		}
		metrics.ApprovalRate = float64(metrics.ApprovedInstances) / float64(metrics.CompletedInstances)
		metrics.RejectionRate = float64(metrics.RejectedInstances) / float64(metrics.CompletedInstances)
	}

	return metrics, nil
}

// GetUserWorkload 获取用户工作负载
func (s *approvalService) GetUserWorkload(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time) (*dto.UserWorkload, error) {
	tasks, err := s.taskRepo.ListByAssignee(ctx, userID, nil, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tasks: %w", err)
	}

	workload := &dto.UserWorkload{
		UserID:   userID,
		UserName: "",
	}

	var totalProcessTime int64
	processedCount := 0

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := today.AddDate(0, 0, -7)
	monthAgo := today.AddDate(0, -1, 0)

	for _, task := range tasks {
		if startDate != nil && task.CreatedAt.Before(*startDate) {
			continue
		}
		if endDate != nil && task.CreatedAt.After(*endDate) {
			continue
		}

		switch task.Status {
		case model.TaskStatusPending:
			workload.PendingTasks++
		case model.TaskStatusApproved:
			workload.ApprovedTasks++
			workload.CompletedTasks++
		case model.TaskStatusRejected:
			workload.RejectedTasks++
			workload.CompletedTasks++
		case model.TaskStatusTransferred:
			workload.TransferredTasks++
		}

		if task.ApprovedAt != nil {
			duration := int64(task.ApprovedAt.Sub(task.CreatedAt).Seconds())
			totalProcessTime += duration
			processedCount++

			if task.ApprovedAt.After(today) {
				workload.TodayTasks++
			}
			if task.ApprovedAt.After(weekAgo) {
				workload.ThisWeekTasks++
			}
			if task.ApprovedAt.After(monthAgo) {
				workload.ThisMonthTasks++
			}
		}
	}

	if processedCount > 0 {
		workload.AvgProcessTime = totalProcessTime / int64(processedCount)
	}

	return workload, nil
}
