FROM golang:latest AS build

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0
RUN go get && go build -o app main.go

FROM alpine:latest

WORKDIR /app
COPY --from=build /go/src/app/app /app/

CMD ["/app/app"]
