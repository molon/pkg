#!/bin/sh

# 这个只是测试使用，生产环境的需要持久化啊，队列大小控制啊等等需要考虑
docker run -d --name jaeger -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 -p 5775:5775/udp -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 14268:14268 -p 9411:9411 --restart=always  jaegertracing/all-in-one:latest