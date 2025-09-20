FROM golang:1.25.1-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o outproxys main.go

EXPOSE 9313

CMD ["./outproxys"]