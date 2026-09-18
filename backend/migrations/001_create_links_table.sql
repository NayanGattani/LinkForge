CREATE TABLE IF NOT EXISTS links (
    id VARCHAR(64) PRIMARY KEY,
    original_url TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    click_count BIGINT NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_links_original_url ON links (original_url);
CREATE INDEX IF NOT EXISTS idx_links_expires_at ON links (expires_at);
