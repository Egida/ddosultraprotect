FROM golang:1.17-stretch as builder


ENV GOPRIVATE=github.com/meldyer1/ddosultraprotect/*

RUN mkdir ../home/app 
WORKDIR /../home/app
COPY . .

RUN git config --global url.git@github.com:.insteadOf https://github.com/

RUN GIT_TERMINAL_PROMPT=1 \
    GOARCH=amd64 \
    GOOS=linux \
    CGO_ENABLED=0 \
    go build --installsuffix 'static' -o /home/app/my-app

FROM alpine 

COPY --from=builder /home/app/my-app/ /

FROM golang:1.17 as go2

WORKDIR .

RUN mkdir /dockerapp

COPY --from=builder /home/app/ /dockerapp/

RUN go env -w GO111MODULE=auto

ENTRYPOINT ["go", "run", "/dockerapp/examples/proposedAlgLB/server/main.go"]

EXPOSE 50051