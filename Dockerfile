FROM golang:alpine3.14 as builder
RUN mkdir /src
ADD . /src/
WORKDIR /src
RUN go build -ldflags "-s -w -X main.version=$(cat VERSION)" -o cnskunkworks-operator
FROM alpine
COPY --from=builder /src/cnskunkworks-operator /app/cnskunkworks-operator
WORKDIR /app
ENTRYPOINT ["/app/cnskunkworks-operator"]