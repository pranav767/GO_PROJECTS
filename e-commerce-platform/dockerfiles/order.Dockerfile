FROM golang:1.24 AS build
WORKDIR /src/services/order-service
COPY . /src
RUN apt-get update && apt-get install -y git build-essential && \
	CGO_ENABLED=0 go build -o /order-service .

FROM gcr.io/distroless/base-debian11
COPY --from=build /order-service /order-service
EXPOSE 8084
ENTRYPOINT ["/order-service"]
