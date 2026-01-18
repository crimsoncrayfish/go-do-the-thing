# Dockerfile for local development PostgreSQL
FROM postgres:16-alpine

# Defaults for local dev
ENV POSTGRES_USER=admin
ENV POSTGRES_PASSWORD=admin
ENV POSTGRES_DB=todo_db
ENV POSTGRES_HOST_AUTH_METHOD=trust
