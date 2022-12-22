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
    go build --installsuffix 'static' -o /my-app

FROM alpine:3.14 

COPY --from=builder /my-app /

EXPOSE 50051

CMD ["/my-app"]

