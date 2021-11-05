FROM golang:1.17.3-alpine AS builder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=1

COPY . /app
WORKDIR /app

RUN apk add --update git make gcc g++ \
 && go install sigs.k8s.io/kustomize/kustomize/v3

RUN make build

FROM alpine:3.14.2

COPY --from=builder /go/bin/kustomize /usr/bin/kustomize
COPY --from=builder /root/.config/kustomize/plugin /root/.config/kustomize/plugin/

ENTRYPOINT ["/usr/bin/kustomize"]
