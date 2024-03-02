select "from",
       from_name,
       count(message_id) as cnt
from messages
where not is_seen
group by ("from",
          from_name)
order by cnt desc
limit 50;

----

select "from",
       from_name,
       mail_box_folder sum(size_bytes) / 1000000 as size_mb
from messages
group by ("from",
          from_name,
          mail_box_folder)
order by size_mb desc -----------------------------------

select "from",
       from_name,
       count(id) as cnt
from messages
group by ("from",
          from_name)
order by cnt desc