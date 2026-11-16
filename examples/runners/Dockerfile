# Account-free contract images for the opt-in integration tests. For a real
# project, prepare its trusted base dependencies too; see docs/RUNNER.md.
FROM python:3.14-alpine3.24 AS python
RUN pip install --no-cache-dir pytest==9.2.0
ENV PYTHONDONTWRITEBYTECODE=1

FROM node:24-alpine3.24 AS node
WORKDIR /deps
RUN npm install --ignore-scripts --no-audit --no-fund vitest@5.1.0
