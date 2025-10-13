-- Download statistics table (track download metrics)

CREATE TABLE IF NOT EXISTS download_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id),

    -- Download metadata
    downloaded_by UUID REFERENCES users(id),         -- User who downloaded (NULL for public)
    ip_address VARCHAR(50),                          -- Download IP
    user_agent TEXT,                                 -- Browser/client info

    -- Download details
    bytes_downloaded BIGINT NOT NULL,                -- Actual bytes transferred
    download_duration_ms INT,                        -- Download time in milliseconds
    is_completed BOOLEAN DEFAULT TRUE,               -- Whether download completed

    -- Source tracking
    source_type VARCHAR(50),                         -- web/api/mobile/integration
    referrer VARCHAR(1000),                          -- HTTP referrer

    -- Geographic info (optional)
    country_code VARCHAR(2),                         -- ISO country code
    city VARCHAR(100),                               -- City name

    -- Timestamp
    downloaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Metadata
    metadata JSONB                                   -- Additional tracking data
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_download_stats_file ON download_stats(file_id);
CREATE INDEX IF NOT EXISTS idx_download_stats_tenant ON download_stats(tenant_id);
CREATE INDEX IF NOT EXISTS idx_download_stats_user ON download_stats(downloaded_by);
CREATE INDEX IF NOT EXISTS idx_download_stats_downloaded_at ON download_stats(downloaded_at);
CREATE INDEX IF NOT EXISTS idx_download_stats_file_user ON download_stats(file_id, downloaded_by);

-- Partitioning hint: Consider partitioning by downloaded_at for better performance
-- Example (commented out, apply manually if needed):
-- CREATE TABLE download_stats_2024_01 PARTITION OF download_stats
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

COMMENT ON TABLE download_stats IS '文件下载统计表，记录每次下载的详细信息';
COMMENT ON COLUMN download_stats.file_id IS '关联的文件ID';
COMMENT ON COLUMN download_stats.downloaded_by IS '下载用户ID（公开访问时为NULL）';
COMMENT ON COLUMN download_stats.bytes_downloaded IS '实际下载字节数';
COMMENT ON COLUMN download_stats.download_duration_ms IS '下载耗时（毫秒）';
COMMENT ON COLUMN download_stats.is_completed IS '下载是否完成（用于区分中断的下载）';
COMMENT ON COLUMN download_stats.source_type IS '下载来源类型（web/api/mobile等）';
