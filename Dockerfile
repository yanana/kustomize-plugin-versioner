FROM golang:1.17.5-alpine AS builder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=1

COPY . /app
WORKDIR /app

RUN apk add --no-cache git make gcc g++ bash curl \
  && curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
  # && go install sigs.k8s.io/kustomize/kustomize/v3@latest

RUN make build

FROM alpine:3.15.0

COPY --from=builder /app/kustomize /usr/bin/kustomize
COPY --from=builder /root/.config/kustomize/plugin /root/.config/kustomize/plugin/

ENTRYPOINT ["/usr/bin/kustomize"]
