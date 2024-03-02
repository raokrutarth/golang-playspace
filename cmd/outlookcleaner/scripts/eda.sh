#!/bin/bash -ex

# https://clickhouse.com/docs/en/interfaces/formats/

./clickhouse local \
    -q "SELECT from, count(id) as cnt FROM file(INBOX_stats.csv, CSVWithNames) GROUP BY from ORDER BY cnt DESC LIMIT 20 FORMAT Pretty"
