#!/bin/bash

cid=$(docker run \
    -d \
    -p 5432:5432 \
    -e POSTGRES_PASSWORD=s3cr3t \
    postgres:15-alpine)
sleep 1
docker cp db/sql/schema.sql ${cid}:/tmp
docker exec ${cid} psql -U postgres -f /tmp/schema.sql
