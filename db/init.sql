CREATE TABLE IF NOT EXISTS users (
    username VARCHAR(255) PRIMARY KEY,
    hashed_password VARCHAR(255),
    github_username VARCHAR(255) UNIQUE,
    github_access_token VARCHAR(500),
    auth_method VARCHAR(50) DEFAULT 'local',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rooms (
    code VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT 'Untitled Hub',
    public BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL DEFAULT 'Untitled Document',
    room_code VARCHAR(10) NOT NULL REFERENCES rooms(code) ON DELETE CASCADE,
    content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(room_code, title)
);

CREATE TABLE IF NOT EXISTS pdfs (
    id SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    room_code VARCHAR(10) NOT NULL REFERENCES rooms(code) ON DELETE CASCADE,
    github_url TEXT NOT NULL,
    uploaded_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(room_code, filename)
);

INSERT INTO rooms (code, name, public) VALUES ('default', 'Default Hub', TRUE) ON CONFLICT (code) DO NOTHING;
INSERT INTO documents (title, content, room_code) VALUES ('Untitled Document', '# Welcome!', 'default') ON CONFLICT (room_code, title) DO NOTHING;