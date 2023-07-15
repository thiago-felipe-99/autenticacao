CREATE TABLE IF NOT EXISTS
  users_roles (
    user_id uuid references users (id),
    name VARCHAR(256) references role (name)
  );
