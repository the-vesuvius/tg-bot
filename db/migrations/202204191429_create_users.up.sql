CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    external_id VARCHAR(255) NOT NULL,
    chat_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_external_id ON users (external_id);