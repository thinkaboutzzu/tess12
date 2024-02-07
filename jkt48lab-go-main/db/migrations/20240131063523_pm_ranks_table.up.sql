CREATE TABLE pm_ranks
(
    username      varchar primary key,
    name          varchar,
    type          varchar,
    since         varchar,
    until         varchar,
    points        int,
    count         int,
    text_count    int,
    image_count   int,
    voice_count   int,
    profile_image varchar
);


