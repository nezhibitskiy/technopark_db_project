CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users(
    id SERIAL,
    nickname CITEXT COLLATE "ucs_basic" PRIMARY KEY UNIQUE,
    fullname TEXT NOT NULL,
    about TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS forum(
    id      SERIAL,
    slug    CITEXT CONSTRAINT forum_pk PRIMARY KEY UNIQUE,
    title   TEXT,
    author  CITEXT NOT NULL,
    posts   INT DEFAULT 0 NOT NULL,
    threads INT DEFAULT 0 NOT NULL
);

CREATE TABLE IF NOT EXISTS thread(
    id SERIAL PRIMARY KEY,
    slug CITEXT NOT NULL,
    title TEXT NOT NULL,
    author CITEXT NOT NULL,
    forum CITEXT NOT NULL,
    message TEXT NOT NULL,
    votes   INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);


CREATE TABLE IF NOT EXISTS posts(
    id SERIAL PRIMARY KEY UNIQUE,
    parent INT NOT NULL,
    path TEXT NOT NULL DEFAULT '',
    author CITEXT NOT NULL,
    forum CITEXT NOT NULL,
    thread INT NOT NULL,
    message TEXT NOT NULL,
    is_edited BOOL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS votes(
    id SERIAL PRIMARY KEY UNIQUE,
    thread_id INT NOT NULL,
    author TEXT NOT NULL,
    value INT NOT NULL
);

CREATE TABLE IF NOT EXISTS forum_users(
    id SERIAL PRIMARY KEY UNIQUE,
    author CITEXT NOT NULL,
    forum CITEXT NOT NULL
);


CREATE INDEX IF NOT EXISTS forum_users_index ON forum_users (author, forum);

CREATE INDEX IF NOT EXISTS thread_slug_index ON thread (slug);
CREATE INDEX IF NOT EXISTS thread_created_at_index ON thread (created_at, forum);
CREATE INDEX IF NOT EXISTS thread_forum_author_index ON thread (forum, author);

CREATE INDEX IF NOT EXISTS post_thread_index ON posts (thread);
CREATE INDEX IF NOT EXISTS post_substring_index ON posts (substring(path,1,7));
create index IF NOT EXISTS post_forum_index on posts (forum, author);

CREATE INDEX IF NOT EXISTS votes_index ON votes (thread_id, author);

create unique index IF NOT EXISTS forum_users_index on forum_users (author, forum);


create function inc_forum_thread() returns trigger as
$$
begin
    update forum set threads = threads + 1 where slug=NEW.forum;
    return NEW;
end;
$$ language plpgsql;

create trigger thread_insert
    after insert
    on thread
    for each row
execute procedure inc_forum_thread();

create function inc_forum_posts() returns trigger as
$$
begin
    update forum set posts = posts + 1 where slug=NEW.forum;
    return NEW;
end;
$$ language plpgsql;

create trigger thread_insert
    after insert
    on posts
    for each row
execute procedure inc_forum_posts();


create function add_forum_user() returns trigger as
$$
begin
    insert into forum_users (forum, author) values (NEW.forum, NEW.author) on conflict do nothing;
    return NEW;
end;
$$ language plpgsql;

create trigger forum_user
    after insert
    on posts
    for each row
execute procedure add_forum_user();

create trigger forum_user
    after insert
    on thread
    for each row
execute procedure add_forum_user();