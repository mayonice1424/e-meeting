FROM golang:1.24.4-alpine

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build

EXPOSE 8080

CMD [ "go", "run", "main.go"]
