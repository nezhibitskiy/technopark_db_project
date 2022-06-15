CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users(
    nickname CITEXT COLLATE "ucs_basic" PRIMARY KEY UNIQUE,
    fullname TEXT NOT NULL,
    about TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS forum(
    slug    CITEXT CONSTRAINT forum_pk PRIMARY KEY UNIQUE,
    title   TEXT,
    author  CITEXT NOT NULL
--     CONSTRAINT forum_author_fk FOREIGN KEY (author) REFERENCES users(nickname)
);

CREATE TABLE IF NOT EXISTS thread(
    id SERIAL PRIMARY KEY UNIQUE,
    title TEXT NOT NULL,
    author CITEXT NOT NULL,
    forum CITEXT NOT NULL,
    message TEXT NOT NULL,
    votes   INT DEFAULT 0 NOT NULL,
    slug CITEXT,
    created_at TIMESTAMP NOT NULL
--     CONSTRAINT thread_forum_fk FOREIGN KEY (forum) REFERENCES forum(slug),
--     CONSTRAINT thread_user_fk FOREIGN KEY (author) REFERENCES users(nickname)
);

CREATE INDEX ON thread (slug);
CREATE INDEX ON thread (created_at, forum);
CREATE INDEX ON thread (forum, author);


CREATE TABLE IF NOT EXISTS posts(
    id SERIAL PRIMARY KEY UNIQUE,
    parent int8 NOT NULL,
    path TEXT NOT NULL DEFAULT '',
    author CITEXT NOT NULL,
    message TEXT NOT NULL,
    is_edited BOOL DEFAULT FALSE,
    thread_id INT NOT NULL,
    created_at timestamp NOT NULL,
    CONSTRAINT posts_user_fk FOREIGN KEY (author) REFERENCES users(nickname)
--     CONSTRAINT posts_thread_fk FOREIGN KEY (thread_id) REFERENCES thread(id)
--     CONSTRAINT posts_post_fk FOREIGN KEY (parent) REFERENCES posts(id)
);

CREATE INDEX ON posts (thread_id);
CREATE INDEX ON posts (substring("path",1,7));


CREATE TABLE IF NOT EXISTS votes(
    id SERIAL PRIMARY KEY UNIQUE,
    thread_id INT NOT NULL,
    author CITEXT NOT NULL,
    value INT NOT NULL
);
CREATE INDEX ON votes (thread_id, author);


CREATE TABLE IF NOT EXISTS forum_users(
    id SERIAL8 PRIMARY KEY UNIQUE,
    author CITEXT NOT NULL,
    forum CITEXT NOT NULL
--     CONSTRAINT forum_users_forum_fk FOREIGN KEY (forum) REFERENCES forum(slug),
--     CONSTRAINT forum_users_users_fk FOREIGN KEY (author) REFERENCES users(nickname)
);
