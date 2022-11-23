FROM golang:1.18.3-alpine AS builder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=1

COPY . /app
WORKDIR /app

RUN apk --no-cache add git make gcc g++ bash curl \
  && GOBIN=$(pwd)/ go install sigs.k8s.io/kustomize/kustomize/v4

RUN make build

FROM alpine:3.17.0

COPY --from=builder /app/kustomize /usr/bin/kustomize
COPY --from=builder /root/.config/kustomize/plugin /root/.config/kustomize/plugin/

ENTRYPOINT ["/usr/bin/kustomize"]
