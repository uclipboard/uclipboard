# linux amd64 version image
FROM debian:stable-slim
LABEL maintainer="dangjinghaoemail@163.com"
WORKDIR /app/

ARG BUILD_DIR=build
ARG PLATFORM=linux-amd64

EXPOSE 4533

RUN mkdir /data && chmod 777 /data

COPY --chmod=755 docker-entrypoint.sh /entrypoint.sh

COPY --chmod=755 ${BUILD_DIR}/uclipboard-${PLATFORM} ./uclipboard

ENTRYPOINT ["/entrypoint.sh"]