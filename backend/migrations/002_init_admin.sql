-- Initialize admin account
-- Default username: admin
-- Default password: admin123 (Please change in production)
-- Password hash is bcrypt encrypted, cost=10

-- Note: This password hash is for 'admin123'
-- In production, use Go's golang.org/x/crypto/bcrypt to generate
-- Example: bcrypt.GenerateFromPassword([]byte("admin123"), 10)

INSERT INTO users (username, password_hash, role, is_active, created_at, updated_at)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'admin',
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO NOTHING;

