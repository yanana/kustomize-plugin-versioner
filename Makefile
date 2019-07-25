NOW := $(shell date "+%Y-%m-%dT%H:%M:%SZ")
GOCMD := go
GOBUILD := ${GOCMD} build
GOCLEAN := ${GOCMD} clean
GOTEST := ${GOCMD} test
XDG_CONFIG_HOME ?= "$(HOME)/.config"
KIND := "Versioner"
L_KIND := "versioner"
SO_NAME := "${KIND}.so"
API_VERSION := "yanana.tokyo/v1"
DESTINATION := "$(XDG_CONFIG_HOME)/kustomize/plugin/${API_VERSION}/${L_KIND}/${SO_NAME}"

default: test build

test:
	@${GOTEST} -v plugin/yanana.tokyo/v1/versioner/Versioner_test.go

build:
	${GOBUILD} -buildmode plugin -o ${DESTINATION} plugin/${API_VERSION}/${L_KIND}/${KIND}.go
