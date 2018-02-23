FROM golang:1.10

WORKDIR /app

ADD . /app

RUN go build -o out

EXPOSE 3000

CMD [ "/app/out" ]