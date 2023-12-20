# Building the binary of the App
FROM golang:1.21 AS build

WORKDIR /go/src/ctfEngine
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o ctfEngine .


# Moving the binary to the 'final Image' to make it smaller
FROM golang:1.21 as release
WORKDIR /app
COPY --from=build /go/src/ctfEngine/ctfEngine .

RUN chmod +x /app/ctfEngine

EXPOSE 3000

CMD ["/app/ctfEngine"]