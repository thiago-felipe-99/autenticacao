CREATE TABLE IF NOT EXISTS
  users(
    id uuid NOT NULL,
    name VARCHAR(256) NOT NULL,
    username VARCHAR(256) NOT NULL,
    email VARCHAR(256) NOT NULL,
    password VARCHAR(256) NOT NULL,
    is_active boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    created_by uuid NOT NULL,
    deleted_at timestamp with time zone NOT NULL,
    deleted_by uuid NOT NULL,
    PRIMARY KEY (id)
  );
