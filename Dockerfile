# Stage 1: Build Vue frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Python app
FROM python:3.12-slim
WORKDIR /app

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

# Install Python deps
COPY pyproject.toml uv.lock ./
RUN uv sync --no-dev --frozen

# Copy app code
COPY src/ src/
COPY config/ config/

# Copy built frontend
COPY --from=frontend /app/frontend/dist frontend/dist

EXPOSE 8080
CMD ["uv", "run", "uvicorn", "rdruid_analyzer.web.app:app", "--host", "0.0.0.0", "--port", "8080"]
