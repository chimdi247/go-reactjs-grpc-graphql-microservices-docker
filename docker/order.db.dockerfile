FROM postgres:16-alpine
COPY docker/order-db/schema.sql /docker-entrypoint-initdb.d/1-schema.sql
CMD ["postgres"]
