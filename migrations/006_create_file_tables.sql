-- File Management Tables

-- Files metadata table
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Basic info
    filename VARCHAR(255) NOT NULL,                  -- Original filename
    storage_key VARCHAR(500) NOT NULL,               -- MinIO object key (unique path)
    size BIGINT NOT NULL,                            -- File size in bytes
    mime_type VARCHAR(255) NOT NULL,                 -- MIME type (e.g., image/jpeg)
    content_type VARCHAR(255),                       -- Content-Type header

    -- Security & integrity
    checksum VARCHAR(64) NOT NULL,                   -- SHA-256 hash
    virus_scanned BOOLEAN DEFAULT FALSE,             -- Virus scan status
    virus_scan_result VARCHAR(50),                   -- clean/infected/error
    virus_scanned_at TIMESTAMP,                      -- Scan timestamp

    -- Metadata
    extension VARCHAR(50),                           -- File extension (.pdf, .jpg)
    bucket VARCHAR(255) NOT NULL,                    -- MinIO bucket name
    category VARCHAR(100),                           -- File category (document/image/video)
    tags TEXT[],                                     -- Searchable tags
    metadata JSONB,                                  -- Additional metadata

    -- Status & flags
    status VARCHAR(50) DEFAULT 'active',             -- active/archived/deleted
    is_temporary BOOLEAN DEFAULT FALSE,              -- Temporary file flag
    is_public BOOLEAN DEFAULT FALSE,                 -- Public access flag

    -- Version control
    version_number INT DEFAULT 1,                    -- Current version number
    parent_file_id UUID REFERENCES files(id),        -- Parent file for versions

    -- Compression & watermark
    is_compressed BOOLEAN DEFAULT FALSE,             -- Compression flag
    has_watermark BOOLEAN DEFAULT FALSE,             -- Watermark flag
    watermark_text VARCHAR(255),                     -- Watermark content

    -- Access control
    uploaded_by UUID NOT NULL REFERENCES users(id),  -- Uploader
    access_level VARCHAR(50) DEFAULT 'private',      -- private/tenant/public

    -- Preview & thumbnail
    thumbnail_key VARCHAR(500),                      -- Thumbnail storage key
    preview_url VARCHAR(1000),                       -- Preview URL (presigned)
    preview_expires_at TIMESTAMP,                    -- Preview URL expiration

    -- Lifecycle
    expires_at TIMESTAMP,                            -- Auto-delete timestamp (for temp files)
    archived_at TIMESTAMP,                           -- Archive timestamp

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP                             -- Soft delete
);

CREATE INDEX IF NOT EXISTS idx_files_tenant ON files(tenant_id);
CREATE INDEX IF NOT EXISTS idx_files_storage_key ON files(storage_key);
CREATE INDEX IF NOT EXISTS idx_files_checksum ON files(checksum);
CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_files_parent ON files(parent_file_id);
CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);
CREATE INDEX IF NOT EXISTS idx_files_is_temporary ON files(is_temporary) WHERE is_temporary = TRUE;
CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_category ON files(category);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
CREATE INDEX IF NOT EXISTS idx_files_tags ON files USING GIN(tags);

-- File versions history table
CREATE TABLE IF NOT EXISTS file_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Version info
    version_number INT NOT NULL,                     -- Version number
    storage_key VARCHAR(500) NOT NULL,               -- Storage key for this version
    size BIGINT NOT NULL,                            -- File size
    checksum VARCHAR(64) NOT NULL,                   -- SHA-256 hash

    -- Metadata
    filename VARCHAR(255) NOT NULL,                  -- Original filename at version
    mime_type VARCHAR(255) NOT NULL,
    comment TEXT,                                    -- Version comment

    -- Change tracking
    changed_by UUID NOT NULL REFERENCES users(id),   -- Who created this version
    change_type VARCHAR(50) DEFAULT 'update',        -- create/update/revert

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_file_versions_unique ON file_versions(file_id, version_number);
CREATE INDEX IF NOT EXISTS idx_file_versions_file ON file_versions(file_id);
CREATE INDEX IF NOT EXISTS idx_file_versions_tenant ON file_versions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_file_versions_created_at ON file_versions(created_at);

-- File relations table (link files to business entities)
CREATE TABLE IF NOT EXISTS file_relations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Related entity
    entity_type VARCHAR(100) NOT NULL,               -- approval_task/form_data/workflow_instance/employee
    entity_id UUID NOT NULL,                         -- Related entity ID

    -- Relation metadata
    field_name VARCHAR(100),                         -- Field name (for form attachments)
    relation_type VARCHAR(50) DEFAULT 'attachment',  -- attachment/avatar/evidence/report
    description TEXT,                                -- Relation description
    sort_order INT DEFAULT 0,                        -- Display order

    -- Access control
    created_by UUID NOT NULL REFERENCES users(id),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_file_relations_file ON file_relations(file_id);
CREATE INDEX IF NOT EXISTS idx_file_relations_entity ON file_relations(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_file_relations_tenant ON file_relations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_file_relations_created_by ON file_relations(created_by);
CREATE UNIQUE INDEX IF NOT EXISTS idx_file_relations_unique
    ON file_relations(file_id, entity_type, entity_id, field_name)
    WHERE field_name IS NOT NULL;

-- Storage quotas table (tenant/user level quotas)
CREATE TABLE IF NOT EXISTS storage_quotas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Quota subject
    subject_type VARCHAR(50) NOT NULL,               -- tenant/user/department
    subject_id UUID,                                 -- NULL for tenant-level quota

    -- Quota limits (in bytes)
    quota_limit BIGINT NOT NULL,                     -- Total quota limit
    quota_used BIGINT DEFAULT 0,                     -- Current usage
    quota_reserved BIGINT DEFAULT 0,                 -- Reserved space (uploading)

    -- File count limits
    file_count_limit INT,                            -- Max file count (optional)
    file_count_used INT DEFAULT 0,                   -- Current file count

    -- Metadata
    settings JSONB,                                  -- Additional settings

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_storage_quotas_unique
    ON storage_quotas(tenant_id, subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_storage_quotas_tenant ON storage_quotas(tenant_id);
CREATE INDEX IF NOT EXISTS idx_storage_quotas_subject ON storage_quotas(subject_type, subject_id);

-- Multipart upload tracking table (for large file uploads)
CREATE TABLE IF NOT EXISTS multipart_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Upload info
    upload_id VARCHAR(255) NOT NULL,                 -- MinIO upload ID
    filename VARCHAR(255) NOT NULL,
    storage_key VARCHAR(500) NOT NULL,               -- Target storage key
    total_size BIGINT,                               -- Expected total size
    part_size BIGINT NOT NULL,                       -- Part size (default 5MB)

    -- Progress tracking
    uploaded_parts JSONB DEFAULT '[]'::JSONB,        -- Array of completed part numbers
    total_parts INT,                                 -- Total number of parts

    -- Metadata
    mime_type VARCHAR(255),
    metadata JSONB,

    -- Status
    status VARCHAR(50) DEFAULT 'in_progress',        -- in_progress/completed/aborted

    -- Owner
    created_by UUID NOT NULL REFERENCES users(id),

    -- Timestamps
    expires_at TIMESTAMP NOT NULL,                   -- Upload expiration (24 hours)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_multipart_uploads_upload_id ON multipart_uploads(upload_id);
CREATE INDEX IF NOT EXISTS idx_multipart_uploads_tenant ON multipart_uploads(tenant_id);
CREATE INDEX IF NOT EXISTS idx_multipart_uploads_created_by ON multipart_uploads(created_by);
CREATE INDEX IF NOT EXISTS idx_multipart_uploads_status ON multipart_uploads(status);
CREATE INDEX IF NOT EXISTS idx_multipart_uploads_expires_at ON multipart_uploads(expires_at);

-- File access logs table (audit trail for file access)
CREATE TABLE IF NOT EXISTS file_access_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Access info
    action VARCHAR(50) NOT NULL,                     -- download/preview/delete/update
    user_id UUID REFERENCES users(id),
    ip_address VARCHAR(50),
    user_agent TEXT,

    -- Result
    success BOOLEAN NOT NULL,
    error_message TEXT,

    -- Metadata
    metadata JSONB,                                  -- Additional context

    -- Timestamp
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_file_access_logs_file ON file_access_logs(file_id);
CREATE INDEX IF NOT EXISTS idx_file_access_logs_tenant ON file_access_logs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_file_access_logs_user ON file_access_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_file_access_logs_action ON file_access_logs(action);
CREATE INDEX IF NOT EXISTS idx_file_access_logs_created_at ON file_access_logs(created_at);

-- Initialize default tenant quota (10GB)
INSERT INTO storage_quotas (tenant_id, subject_type, quota_limit)
VALUES ('00000000-0000-0000-0000-000000000001', 'tenant', 10737418240) -- 10GB in bytes
ON CONFLICT DO NOTHING;
