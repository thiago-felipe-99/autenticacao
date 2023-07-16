CREATE TABLE IF NOT EXISTS
  users (
    id uuid NOT NULL,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    is_active boolean NOT NULL,
    roles VARCHAR(255) [] NOT NULL,
    created_at timestamp with time zone NOT NULL,
    created_by uuid NOT NULL,
    deleted_at timestamp with time zone NOT NULL,
    deleted_by uuid NOT NULL,
    PRIMARY KEY (id)
  );
