FROM golang:latest as builder
RUN mkdir /build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
COPY internal ./internal
RUN ls -alh && go build .

FROM ipfs/kubo:latest as ipfs 

FROM debian:stable-slim

RUN set -eux; \
	apt-get update; \
	apt-get install -y \
		tini \
        chromium \
    # Using gosu (~2MB) instead of su-exec (~20KB) because it's easier to
    # install on Debian. Useful links:
    # - https://github.com/ncopa/su-exec#why-reinvent-gosu
    # - https://github.com/tianon/gosu/issues/52#issuecomment-441946745
		gosu \
    # This installs fusermount which we later copy over to the target image.
    fuse \
    ca-certificates \
	; \
	rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/pinshare/bin
COPY entrypoint.sh /opt/pinshare/bin/entrypoint.sh
COPY --from=builder /build/pinshare /opt/pinshare/bin/pinshare

### Test our chromium headless browser works performantly.
RUN date && cd /opt/pinshare/bin && ./pinshare testcdp none && date
RUN date && cd /opt/pinshare/bin && ./pinshare testsl 3df79d34abbca99308e79cb94461c1893582604d68329a41fd4bec1885e6adb4 && date

### ref:https://github.com/ipfs/kubo/blob/master/Dockerfile
COPY --from=ipfs /sbin/gosu /sbin/gosu
COPY --from=ipfs /sbin/tini /sbin/tini
COPY --from=ipfs /usr/local/bin/fusermount /usr/local/bin/fusermount
COPY --from=ipfs /usr/local/bin/ipfs /usr/local/bin/ipfs
COPY --from=ipfs /usr/local/bin/start_ipfs /usr/local/bin/start_ipfs
COPY --from=ipfs /usr/local/bin/container_init_run /usr/local/bin/container_init_run

# Add suid bit on fusermount so it will run properly
RUN chmod 4755 /usr/local/bin/fusermount

# Fix permissions on start_ipfs (ignore the build machine's permissions)
RUN chmod 0755 /usr/local/bin/start_ipfs

# Create the fs-repo directory and switch to a non-privileged user.
ENV IPFS_PATH /data/ipfs
RUN mkdir -p $IPFS_PATH \
  && adduser --disabled-password --home $IPFS_PATH --uid 1000 --ingroup users ipfs \
  && chown ipfs:users $IPFS_PATH

# Create mount points for `ipfs mount` command
RUN mkdir /ipfs /ipns /mfs \
  && chown ipfs:users /ipfs /ipns /mfs

# Create the init scripts directory
RUN mkdir /container-init.d \
  && chown ipfs:users /container-init.d

# Expose the fs-repo as a volume.
# start_ipfs initializes an fs-repo if none is mounted.
# Important this happens after the USER directive so permissions are correct.
VOLUME $IPFS_PATH

# The default logging level
ENV GOLOG_LOG_LEVEL ""
# Healthcheck for the container
# QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn is the CID of empty folder
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ipfs --api=/ip4/127.0.0.1/tcp/5001 dag stat /ipfs/QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn || exit 1

ENTRYPOINT ["/sbin/tini", "--", "sh"]
CMD [ "/opt/pinshare/bin/entrypoint.sh" ]

ENV PATH /opt/pinshare/bin:$PATH

ENV VT_TOKEN=REDACTED
ENV GITHUB_TIMELINE_ACCESS_TOKEN=REDACTED

ENV PORT-IPFS=5001
ENV PORT-API=9090
ENV PORT-ADMIN-API=10000

ENV PS_FF_MOVE_UPLOAD=false 
ENV PS_FF_SENDFILE_VT=false 
ENV PS_FF_SKIP_VT=false 
ENV PS_FF_IGNORE_UPLOADS_IN_METADATA=true

#  content scanner api
EXPOSE 9090 
# admin api content moderation
EXPOSE 10000
# Swarm TCP; should be exposed to the public
EXPOSE 4001
# Swarm UDP; should be exposed to the public
EXPOSE 4001/udp
# Daemon API; must not be exposed publicly but to client services under you control
EXPOSE 5001
# Web Gateway; can be exposed publicly with a proxy, e.g. as https://ipfs.example.org
EXPOSE 8080
# Swarm Websockets; must be exposed publicly when the node is listening using the websocket transport (/ipX/.../tcp/8081/ws).
EXPOSE 8081
# Our LibP2P Port
EXPOSE 50001