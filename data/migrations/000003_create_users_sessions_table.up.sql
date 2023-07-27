CREATE TABLE IF NOT EXISTS
  users_sessions_created (
    id uuid NOT NULL,
    userid uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
  );


CREATE TABLE IF NOT EXISTS
  users_sessions_deleted (
    id uuid NOT NULL,
    userid uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
  );
