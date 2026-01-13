FROM debian:trixie-slim

WORKDIR /app

COPY backend-service .

EXPOSE 8080

ENV DATABASE_URL=postgresql://udd-mirror:udd-mirror@udd-mirror.debian.net:5432/udd

CMD ["./backend-service"]
