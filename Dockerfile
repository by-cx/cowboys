FROM golang:1.18.0-alpine3.15 AS builder

RUN apk add make gcc musl-dev
ADD . /src/
RUN cd /src && make build

FROM alpine:3.15 AS runner
COPY --from=builder /src/bin/cowboy-v1-amd64 /usr/local/bin/cowboy
COPY --from=builder /src/bin/universe-v1-amd64 /usr/local/bin/universe
COPY --from=builder /src/bin/timetraveler-v1-amd64 /usr/local/bin/timetraveler

WORKDIR /srv/

ENTRYPOINT ["/bin/sh", "-c"]
