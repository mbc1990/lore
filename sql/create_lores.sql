create table lores(
  lore_id serial primary key not null,
  user_id varchar(1024) not null,
  message text not null,
  timestamp_added timestamp default current_timestamp,
  score int
)
