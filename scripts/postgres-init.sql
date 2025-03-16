-- Create users table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);

-- Create an admin user with password 'admin123' (bcrypt hashed)
INSERT INTO users (id, email, username, password, first_name, last_name, role, status)
VALUES (
    uuid_generate_v4(),
    'admin@example.com',
    'admin',
    '$2a$12$tLUB1UBHhUaJmXKDOyJEJuVeZDiEu9wcUuDmO2i6gvYqfM1qg7yLe', -- admin123
    'Admin',
    'User',
    'admin',
    'active'
) ON CONFLICT (email) DO NOTHING;

-- Create a test user with password 'test123' (bcrypt hashed)
INSERT INTO users (id, email, username, password, first_name, last_name, role, status)
VALUES (
    uuid_generate_v4(),
    'test@example.com',
    'testuser',
    '$2a$12$9/KQPljPTQK4rdR1MgQ8DetkJPg8GXf3wkYbYNdRLGJYxlFTiX.S2', -- test123
    'Test',
    'User',
    'user',
    'active'
) ON CONFLICT (email) DO NOTHING;