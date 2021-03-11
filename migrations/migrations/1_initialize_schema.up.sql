create table if not exists users
(
    id         serial not null,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name       text,
    email      text   not null,
    constraint users_pkey
    primary key (id, email),
    constraint unique_email
    unique (email)
    );


create table if not exists channels
(
    id                serial not null,
    created_at        timestamp with time zone,
    updated_at        timestamp with time zone,
    deleted_at        timestamp with time zone,
    title             text,
    name              text,
    secret            text,
    host_passphrase   text,
    viewer_passphrase text,
    dtmf              text,
    constraint channels_pkey
    primary key (id)
    );

create table if not exists tokens
(
    id         serial not null,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    token_id   text,
    user_email text,
    constraint tokens_pkey
    primary key (id),
    constraint tokens_user_email_fkey
    foreign key (user_email) references users (email)
    );

