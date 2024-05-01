FROM golang:1.17-alpine as builder

ARG GITLAB_TOKEN
ARG CI_PIPELINE_IID

WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY src src
COPY assets assets
RUN go build -ldflags "-X github.com/PCManiac/compile_vars.version=${CI_PIPELINE_IID} -X github.com/PCManiac/compile_vars.build_time=`date +%FT%T%z`" -o server ./...

FROM alpine:3.14
COPY --from=builder /app/server .
COPY --from=builder /app/assets /assets
ENTRYPOINT ["./server"]
