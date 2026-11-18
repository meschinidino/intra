FROM golang:1.26.5-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/intrad ./cmd/intrad && CGO_ENABLED=0 go build -o /out/intra ./cmd/intra && CGO_ENABLED=0 go build -o /out/labservice ./cmd/labservice

FROM alpine:3.20
RUN adduser -D -H intra
COPY --from=build /out/ /usr/local/bin/
USER intra
ENTRYPOINT ["intrad"]
