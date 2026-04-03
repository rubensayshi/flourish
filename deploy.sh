#!/usr/bin/env bash
set -e

echo "Building frontend..."
(cd frontend && npm run build)

echo "Deploying to Fly.io..."
fly deploy
