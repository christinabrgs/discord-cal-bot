FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o discord-bot

FROM alpine
ENV SKIP_ENV=true 
COPY --from=builder /app/discord-bot /

ENTRYPOINT [ "/discord-bot" ]
