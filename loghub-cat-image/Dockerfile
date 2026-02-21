FROM alpine:3.20

# coreutils 提供比較完整的 tail（含 -F 行為一致）
RUN apk add --no-cache coreutils

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
