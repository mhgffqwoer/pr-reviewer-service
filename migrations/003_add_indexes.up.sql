-- Индексы для pull_requests
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_status ON pull_requests(author_id, status);

-- Индексы для users  
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Составной индекс для оптимизации JOIN операций
CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name, is_active);