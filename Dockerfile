FROM golang:alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o go-connector-bedrock .

FROM alpine:edge

WORKDIR /app

COPY --from=build /app/go-connector-bedrock .

RUN apk --no-cache add ca-certificates tzdata

ENTRYPOINT ["/app/go-connector-bedrock"]