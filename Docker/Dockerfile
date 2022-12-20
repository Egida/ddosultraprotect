# Multistage build

# Build stage I: 
FROM golang AS build
ENV location /go/src/github.com/meldyer1/ddosultraprotect
WORKDIR ${location}/server

ADD ./server ${location}/server
ADD ./proto-gen ${location}/proto-gen

RUN go get -d ./... 
RUN go install ./... 

RUN CGO_ENABLED=0 go build -o /bin/grpc-productinfo-server

# Build stage II: 
FROM scratch
COPY --from=build /bin/grpc-productinfo-server /bin/grpc-productinfo-server

ENTRYPOINT ["/bin/grpc-productinfo-server"]
EXPOSE 50051