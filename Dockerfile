# syntax=docker/dockerfile:1

FROM golang:1.16-c

WORKDIR /app

COPY start-service /app/service
RUN chmod +x /app/service
RUN mkdir /app/logs

ENV SR_START_PORT=8080
ENV TZ=Europe/Berlin

EXPOSE 8080

ENTRYPOINT [ "./service" ]
