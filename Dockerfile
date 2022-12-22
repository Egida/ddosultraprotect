FROM golang:1.17-stretch as builder


ENV GOPRIVATE=github.com/meldyer1/ddosultraprotect/*

RUN mkdir ../home/app 
WORKDIR /../home/app
COPY . .


RUN git config --global url.git@github.com:.insteadOf https://github.com/

RUN GIT_TERMINAL_PROMPT=1 \
    GOARCH=amd64 \
    GOOS=windows \
    CGO_ENABLED=0 \
    go build --installsuffix 'static' -o app

WORKDIR /../home/app

COPY ./go.mod ./proposedAlgLB/server/go.mod

COPY ./go.sum ./proposedAlgLB/server/go.sum

EXPOSE 50051