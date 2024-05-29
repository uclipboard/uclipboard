# linux amd64 version image
FROM debian:stable
LABEL maintainer="dangjinghaoemail@163.com"
WORKDIR /app/

ARG BUILD_DIR=build
ARG PLATFORM=linux-amd64

COPY docker-entrypoint.sh /entrypoint.sh
COPY ${BUILD_DIR}/uclipboard-${PLATFORM} ./uclipboard

# RUN sed -i 's/deb.debian.org/mirrors.ustc.edu.cn/g' /etc/apt/sources.list.d/debian.sources

RUN apt update && apt install bash -y && \
    apt clean && \
    chmod +x /entrypoint.sh && \
    chmod +x /app/uclipboard

ENV PLATFORM=$PLATFORM \
    PUID=0 PGID=0 UMASK=022 

EXPOSE 4533

ENTRYPOINT ["/entrypoint.sh"]