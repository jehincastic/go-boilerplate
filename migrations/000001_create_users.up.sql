CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT uuidv7(),
  created_at timestamptz NOT NULL DEFAULT now(),
  name text NOT NULL,
  email citext NOT NULL UNIQUE,
  password_hash text NOT NULL
);
