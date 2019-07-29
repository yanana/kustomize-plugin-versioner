FROM golang:1.12-alpine AS builder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=1

COPY . /app
WORKDIR /app

RUN apk add --update git make gcc g++ \
 && go install sigs.k8s.io/kustomize/v3/cmd/kustomize

RUN make build

FROM gcr.io/distroless/static
FROM alpine:3.10

COPY --from=builder /go/bin/kustomize /usr/bin/kustomize
COPY --from=builder /root/.config/kustomize/plugin /root/.config/kustomize/plugin/

ENTRYPOINT ["/usr/bin/kustomize"]
