-- Links created by the shortener. The code is the primary key, so the
-- database enforces uniqueness even between concurrent requests.
CREATE TABLE links (
    code       text PRIMARY KEY,
    url        text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
