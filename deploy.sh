#!/bin/bash

echo "Entering VM..."
ssh root@ << 'EOF'
  set -e
  cd

  echo "Pulling code..."
  git pull

  echo "Building Docker containers..."
  docker compose build

  echo "Starting containers..."
  docker compose up -d

  echo "Deployment complete."
EOF
