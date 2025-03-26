#!/bin/bash

VERSION=$(head -n 1 VERSION)
echo $VERSION

docker build -t registry.yfb.sunline.cn/justsong/one-api:${VERSION} .
docker save registry.yfb.sunline.cn/justsong/one-api:${VERSION} | gzip > one-api__${VERSION}.tgz
