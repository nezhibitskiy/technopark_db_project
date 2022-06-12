FROM golang:latest AS builder

WORKDIR /app

COPY . ./
RUN GOAMD64=v3 go build -ldflags "-w -s" ./main.go

FROM ubuntu:20.04

RUN apt-get -y update && apt-get install -y tzdata
ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get -y update && apt-get install -y postgresql-12 && rm -rf /var/lib/apt/lists/*
USER postgres

RUN /etc/init.d/postgresql start && \
  psql --command "CREATE USER defaultuser WITH SUPERUSER PASSWORD 'password';" && \
  createdb -O defaultuser sqlhw && \
  /etc/init.d/postgresql stop

EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /cmd

RUN mkdir /cmd/configs
VOLUME ["/cmd/configs"]

COPY ./db/db.sql ./db.sql
COPY ./.env ./.env
COPY --from=builder /app/main .

EXPOSE 5000
ENV PGPASSWORD password
CMD service postgresql start && psql -h localhost -d sqlhw -U defaultuser -p 5432 -a -q -f ./db.sql && ./main