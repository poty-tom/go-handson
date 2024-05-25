drop schema if exists sample;
create schema sample;
use sample;

drop table if exists articles;
drop table if exists comments;

create table if not exists articles (
    article_id integer unsigned auto_increment primary key,
    title varchar(100) not null,
    contents text not null,
    username varchar(100) not null,
    nice integer not null,
    created_at datetime
);

create table if not exists comments (
    comment_id integer unsigned auto_increment primary key,
    article_id integer unsigned not null,
    message text not null,
    created_at datetime,
    foreign key (article_id) references articles(article_id)
);

insert into articles 
values 
(
    1, 
    "devcontainerが立った件", 
    "devcontainerが立つまでに1週間を要した。そのくせただのネットの情報を流用しただけという。",
    "tarou nakajima",
    100,
    CURDATE()
);

insert into comments
values
(
    1,
    1,
    "いい年して言い訳ばかりする悲しい記事",
    CURDATE()
);
