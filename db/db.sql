create extension if not exists citext;

create table "user"
(
    "id"       serial,
    "nickname" citext COLLATE "ucs_basic" not null primary key,
    "email"    citext not null unique,
    "fullname" text   not null,
    "about"    text   not null default ''
);

create index index_users_nickname_hash ON "user" USING HASH ("nickname");
create index index_users_email_hash ON "user" USING HASH ("email");
create index index_users_id ON "user" USING HASH ("id");


create table "forum"
(
    "id"      serial,
    "slug"    citext        not null primary key,
    "title"   text          not null,
    "user"    citext        not null,
    "posts"   int default 0 not null,
    "threads" int default 0 not null
);

create index index_forums ON "forum" ("slug", "title", "user", "posts", "threads");
create index index_forums_slug_hash ON "forum" USING HASH ("slug");
create index index_forums_users_foreign ON "forum" USING HASH ("user");
create index index_forums_id_hash ON "forum" USING HASH ("id");

create table "thread"
(
    "id"      serial primary key,
    "slug"    citext        not null,
    "title"   text          not null,
    "author"  text          not null,
    "forum"   text          not null,
    "message" text          not null,
    "votes"   int default 0 not null,
    "created" timestamptz   not null
);

create index index_threads_forum_created ON "thread" ("forum", "id");
create index index_threads_forum_ID ON "thread" ("forum", "created");
create index index_threads_created ON "thread" ("created");
create index index_threads_slug_hash ON "thread" USING HASH ("slug");
create index index_threads_id_hash ON "thread" USING HASH ("id");

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


create table "post"
(
    "id"       serial primary key,
    "parent"   int         not null,
    "path"     text        not null default '',
    "author"   text        not null,
    "forum"    text        not null,
    "thread"   int         not null,
    "message"  text        not null,
    "isEdited" bool        not null default false,
    "created"  timestamptz not null
);

create index index_posts_id on "post" USING HASH ("id");
create index index_posts_thread_id on "post" ("thread", "id");
create index index_posts_thread_parent_path on "post" ("thread", "parent", "path");
create index on "post" (substring("path",1,7));


create table "vote"
(
    "id"       serial primary key,
    "thread"   int  not null,
    "nickname" text not null,
    "voice"    int  not null
);
create index on "vote" ("thread", "nickname");


create table "forum_user"
(
    "forum" text not null,
    "user"  text not null
);
create unique index on "forum_user" ("user", "forum");


create function add_forum_user() returns trigger as
$$
begin
    insert into forum_user (forum, "user") values (NEW.forum, NEW.author) on conflict do nothing;
    return NEW;
end;
$$ language plpgsql;

create trigger forum_user
    after insert
    on post
    for each row
execute procedure add_forum_user();
create trigger forum_user
    after insert
    on thread
    for each row
execute procedure add_forum_user();
