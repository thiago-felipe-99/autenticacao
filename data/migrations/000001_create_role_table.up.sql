CREATE TABLE IF NOT EXISTS
  role (
    name VARCHAR(256) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    created_by uuid NOT NULL,
    deleted_at timestamp with time zone NOT NULL,
    deleted_by uuid NOT NULL,
    PRIMARY KEY (name)
  );
