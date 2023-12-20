FROM alpine:3.18.5

RUN apk add socat bash
WORKDIR /app
COPY bind.sh ./
RUN chmod +x bind.sh

CMD ["/app/bind.sh"]

EXPOSE 42023