FROM golang:1.12 AS build
WORKDIR /go
ADD . ./src/
RUN CGO_ENABLED=0 GOOS=linux go build -o helb cli/balance

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=build /go/helb /helb
ENTRYPOINT ["/helb"]
