create table if not exists public.links (
  id bigint primary key,
  short_id varchar(16) not null unique,
  long_url text not null,
  created_at timestamptz not null default now()
);

create index if not exists idx_links_short_id on public.links (short_id);
