package service

import (
"context"
"fmt"

"github.com/google/uuid"
"github.com/lk2023060901/go-next-erp/internal/file/model"
"github.com/lk2023060901/go-next-erp/internal/file/repository"
"github.com/lk2023060901/go-next-erp/pkg/logger"
"go.uber.org/zap"
)

// FileRelationService 文件关联服务接口
type FileRelationService interface {
	// 创建关联
	AttachFileToEntity(ctx context.Context, req *AttachFileRequest) error
	AttachMultipleFiles(ctx context.Context, req *AttachMultipleFilesRequest) error

	// 查询关联
	GetEntityFiles(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.File, error)
	GetEntityFilesByField(ctx context.Context, entityType model.EntityType, entityID uuid.UUID, fieldName string) ([]*model.File, error)
	GetFileRelations(ctx context.Context, fileID uuid.UUID) ([]*model.FileRelation, error)

	// 删除关联
	DetachFile(ctx context.Context, relationID uuid.UUID) error
	DetachFileFromEntity(ctx context.Context, fileID uuid.UUID, entityType model.EntityType, entityID uuid.UUID) error
	DetachAllFilesFromEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) error
}

// AttachFileRequest 附加文件请求
type AttachFileRequest struct {
	FileID       uuid.UUID
	TenantID     uuid.UUID
	EntityType   model.EntityType
	EntityID     uuid.UUID
	FieldName    *string
	RelationType model.RelationType
	Description  *string
	SortOrder    int
	CreatedBy    uuid.UUID
}

// AttachMultipleFilesRequest 批量附加文件请求
type AttachMultipleFilesRequest struct {
	FileIDs      []uuid.UUID
	TenantID     uuid.UUID
	EntityType   model.EntityType
	EntityID     uuid.UUID
	FieldName    *string
	RelationType model.RelationType
	CreatedBy    uuid.UUID
}

type fileRelationService struct {
	fileRepo         repository.FileRepository
	relationRepo     repository.FileRelationRepository
	logger           *logger.Logger
}

// NewFileRelationService 创建文件关联服务
func NewFileRelationService(
fileRepo repository.FileRepository,
relationRepo repository.FileRelationRepository,
logger *logger.Logger,
) FileRelationService {
	return &fileRelationService{
		fileRepo:     fileRepo,
		relationRepo: relationRepo,
		logger:       logger.With(zap.String("service", "file_relation")),
	}
}

// AttachFileToEntity 将文件附加到实体
func (s *fileRelationService) AttachFileToEntity(ctx context.Context, req *AttachFileRequest) error {
	// 验证文件是否存在
	file, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 验证租户匹配
	if file.TenantID != req.TenantID {
		return fmt.Errorf("tenant mismatch")
	}

	// 创建关联
	relation := &model.FileRelation{
		FileID:       req.FileID,
		TenantID:     req.TenantID,
		EntityType:   string(req.EntityType),
		EntityID:     req.EntityID,
		FieldName:    req.FieldName,
		RelationType: string(req.RelationType),
		Description:  req.Description,
		SortOrder:    req.SortOrder,
		CreatedBy:    req.CreatedBy,
	}

	if err := s.relationRepo.Create(ctx, relation); err != nil {
		return fmt.Errorf("failed to create file relation: %w", err)
	}

	s.logger.Info("File attached to entity",
zap.String("file_id", req.FileID.String()),
		zap.String("entity_type", string(req.EntityType)),
		zap.String("entity_id", req.EntityID.String()))

	return nil
}

// AttachMultipleFiles 批量附加文件
func (s *fileRelationService) AttachMultipleFiles(ctx context.Context, req *AttachMultipleFilesRequest) error {
	for i, fileID := range req.FileIDs {
		attachReq := &AttachFileRequest{
			FileID:       fileID,
			TenantID:     req.TenantID,
			EntityType:   req.EntityType,
			EntityID:     req.EntityID,
			FieldName:    req.FieldName,
			RelationType: req.RelationType,
			SortOrder:    i, // 按照数组顺序排序
			CreatedBy:    req.CreatedBy,
		}

		if err := s.AttachFileToEntity(ctx, attachReq); err != nil {
			s.logger.Error("Failed to attach file",
zap.String("file_id", fileID.String()),
				zap.Error(err))
			// 继续处理其他文件
			continue
		}
	}

	s.logger.Info("Multiple files attached",
zap.Int("count", len(req.FileIDs)),
zap.String("entity_type", string(req.EntityType)),
zap.String("entity_id", req.EntityID.String()))

	return nil
}

// GetEntityFiles 获取实体的所有关联文件
func (s *fileRelationService) GetEntityFiles(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) ([]*model.File, error) {
	// 查找关联
	relations, err := s.relationRepo.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to find relations: %w", err)
	}

	// 获取文件详情
	files := make([]*model.File, 0, len(relations))
	for _, relation := range relations {
		file, err := s.fileRepo.FindByID(ctx, relation.FileID)
		if err != nil {
			s.logger.Warn("Failed to find file for relation",
zap.String("relation_id", relation.ID.String()),
				zap.String("file_id", relation.FileID.String()),
				zap.Error(err))
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

// GetEntityFilesByField 获取实体特定字段的关联文件
func (s *fileRelationService) GetEntityFilesByField(ctx context.Context, entityType model.EntityType, entityID uuid.UUID, fieldName string) ([]*model.File, error) {
	// 查找关联
	relations, err := s.relationRepo.FindByEntityAndField(ctx, entityType, entityID, fieldName)
	if err != nil {
		return nil, fmt.Errorf("failed to find relations: %w", err)
	}

	// 获取文件详情
	files := make([]*model.File, 0, len(relations))
	for _, relation := range relations {
		file, err := s.fileRepo.FindByID(ctx, relation.FileID)
		if err != nil {
			s.logger.Warn("Failed to find file for relation",
zap.String("relation_id", relation.ID.String()),
				zap.String("file_id", relation.FileID.String()),
				zap.Error(err))
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

// GetFileRelations 获取文件的所有关联
func (s *fileRelationService) GetFileRelations(ctx context.Context, fileID uuid.UUID) ([]*model.FileRelation, error) {
	return s.relationRepo.FindByFileID(ctx, fileID)
}

// DetachFile 解除文件关联
func (s *fileRelationService) DetachFile(ctx context.Context, relationID uuid.UUID) error {
	if err := s.relationRepo.Delete(ctx, relationID); err != nil {
		return fmt.Errorf("failed to delete file relation: %w", err)
	}

	s.logger.Info("File relation deleted", zap.String("relation_id", relationID.String()))
	return nil
}

// DetachFileFromEntity 解除特定文件与实体的关联
func (s *fileRelationService) DetachFileFromEntity(ctx context.Context, fileID uuid.UUID, entityType model.EntityType, entityID uuid.UUID) error {
	// 查找关联
	relations, err := s.relationRepo.FindByEntity(ctx, entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to find relations: %w", err)
	}

	// 删除匹配的关联
	for _, relation := range relations {
		if relation.FileID == fileID {
			if err := s.relationRepo.Delete(ctx, relation.ID); err != nil {
				s.logger.Error("Failed to delete relation",
zap.String("relation_id", relation.ID.String()),
					zap.Error(err))
				continue
			}
		}
	}

	return nil
}

// DetachAllFilesFromEntity 解除实体的所有文件关联
func (s *fileRelationService) DetachAllFilesFromEntity(ctx context.Context, entityType model.EntityType, entityID uuid.UUID) error {
	if err := s.relationRepo.DeleteByEntity(ctx, entityType, entityID); err != nil {
		return fmt.Errorf("failed to delete entity relations: %w", err)
	}

	s.logger.Info("All files detached from entity",
zap.String("entity_type", string(entityType)),
zap.String("entity_id", entityID.String()))

	return nil
}
