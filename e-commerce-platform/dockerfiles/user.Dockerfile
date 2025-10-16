FROM golang:1.24 AS build
WORKDIR /src/services/user-service
COPY . /src
RUN apt-get update && apt-get install -y git build-essential && \
	CGO_ENABLED=0 go build -o /user-service .

FROM gcr.io/distroless/base-debian11
COPY --from=build /user-service /user-service
EXPOSE 8082
ENTRYPOINT ["/user-service"]
