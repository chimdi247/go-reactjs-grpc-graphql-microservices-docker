FROM postgres:16-alpine
COPY docker/account-db/schema.sql /docker-entrypoint-initdb.d/1-schema.sql
CMD ["postgres"]
