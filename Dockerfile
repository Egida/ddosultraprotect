#
# copy go code:
#
FROM meldyer/ddosultraprotect
RUN go mod download
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]

