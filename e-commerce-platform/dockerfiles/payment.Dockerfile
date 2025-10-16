FROM golang:1.24 AS build
WORKDIR /src/services/payment-service
COPY . /src
RUN apt-get update && apt-get install -y git build-essential && \
	CGO_ENABLED=0 go build -o /payment-service .

FROM gcr.io/distroless/base-debian11
COPY --from=build /payment-service /payment-service
EXPOSE 8085
ENTRYPOINT ["/payment-service"]
