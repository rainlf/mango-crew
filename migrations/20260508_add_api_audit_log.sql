CREATE TABLE IF NOT EXISTS api_audit_log (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    request_id VARCHAR(64) NOT NULL,
    user_id INT NULL,
    http_method VARCHAR(16) NOT NULL,
    path VARCHAR(1024) NOT NULL,
    http_status INT NOT NULL,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    client_ip VARCHAR(64) NULL,
    user_agent VARCHAR(255) NULL,
    request LONGTEXT NULL,
    response LONGTEXT NULL,
    error TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
