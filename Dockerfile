FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:latest as builder
RUN mkdir /build
WORKDIR /build
COPY main.go .
COPY go.mod . 
COPY go.sum .
COPY internal ./internal 
RUN go mod tidy && go build .

FROM --platform=${BUILDPLATFORM:-linux/amd64} chromedp/headless-shell:latest as headless 
RUN ls -alh  /usr/lib/

FROM --platform=${BUILDPLATFORM:-linux/amd64} ipfs/kubo:latest
# https://github.com/ipfs/kubo/blob/master/Dockerfile
# 2025/06/26 13:15:14 exec: "google-chrome": executable file not found in $PATH

# [INFO] Scanning './upload' for new files...
# 2025/06/26 13:43:05 chrome failed to start:
# /opt/headless-shell/headless-shell: error while loading shared libraries: libdl.so.2: cannot open shared object file: No such file or directory
# ./headless-shell: error while loading shared libraries: libnspr4.so: cannot open shared object file: No such file or directory
RUN mkdir -p /opt/pinshare/bin
COPY entrypoint.sh /opt/pinshare/bin/entrypoint.sh
COPY --from=builder /build/pinshare /opt/pinshare/bin/pinshare

COPY --from=headless /headless-shell /opt/headless-shell 
COPY --from=headless /usr/lib/aarch64-linux-gnu/ /lib/


ENTRYPOINT ["/sbin/tini", "--", "sh"]
CMD [ "/opt/pinshare/bin/entrypoint.sh" ]

ENV PATH /opt/headless-shell:$PATH

ENV VT_TOKEN=REDACTED
ENV GITHUB_TIMELINE_ACCESS_TOKEN=REDACTED

ENV PORT-IPFS=5001
ENV PORT-API=9090
ENV PORT-ADMIN-API=10000

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