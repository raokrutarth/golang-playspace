#!/bin/bash -ex

curl https://clickhouse.com/ | sh
sudo ./clickhouse install
rm -rm ./clickhouse
# sudo mv ./clickhouse /usr/local/bin/clickhouse