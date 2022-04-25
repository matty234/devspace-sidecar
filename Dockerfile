FROM --platform=arm64 golang:1.18.0 AS build


WORKDIR /build
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod download

COPY . .

RUN go build -o ./devspace . 

ENTRYPOINT [ "./devspace" ]