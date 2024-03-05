FROM golang:1.22

WORKDIR /app
COPY /test-goreleaser /app/test-goreleaser

EXPOSE 8080

CMD ["./test-goreleaser"]