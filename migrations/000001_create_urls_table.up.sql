create table if not exists urls (
    id serial primary key,
    short_url text not null unique,
    long_url text not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now()
);
