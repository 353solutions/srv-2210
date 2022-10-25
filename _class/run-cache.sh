#!/bin/bash

docker run \
    -d \
    -p 6379:6379 \
    redis:7-alpine
