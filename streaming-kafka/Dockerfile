FROM golang:latest AS build

WORKDIR /go/src/app
COPY . .

ENV CGO_ENABLED=0
RUN go get && go build -o streaming *.go

FROM alpine:latest

RUN apk --update add tzdata

WORKDIR /app
COPY --from=build /go/src/app/streaming /app/

CMD ["/app/streaming"]
