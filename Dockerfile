FROM golang:1.17-stretch as builder


ENV GOPRIVATE=github.com/meldyer1/ddosultraprotect/*

RUN mkdir ../home/app 
WORKDIR /../home/app
COPY . .

RUN git config --global url.git@github.com:.insteadOf https://github.com/

RUN go env -w GO111MODULE=auto

ENV GIT_TERMINAL_PROMPT=1
ENV GOARCH=amd64
ENV GOOS=linux
ENV CGO_ENABLED=0

RUN go get google.golang.org/grpc/examples/helloworld/helloworld 

ENTRYPOINT ["go", "run"]


EXPOSE 50051