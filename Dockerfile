
FROM golang:1.22.5-alpine AS build
WORKDIR /app
COPY . /app
RUN go build .

FROM scratch
COPY --from=build /app/GoBroker /GoBroker
CMD ["/GoBroker"]
