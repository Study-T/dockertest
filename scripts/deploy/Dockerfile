# 部署用 Dockerfile（可自定义）
FROM alpine:3.19
COPY app/server /app/server
COPY app/etc/app.yaml /app/etc/app.yaml
WORKDIR /app
CMD ["./server", "-f", "etc/app.yaml"]
