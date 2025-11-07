#!/bin/bash
set -e

HOST="${1:-localhost:50051}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROTO_DIR="$SCRIPT_DIR/../proto"

echo "Testing log ingestion to $HOST..."

grpcurl -plaintext -import-path "$PROTO_DIR" -proto logs.proto -d '{
  "logs": [
    {
      "timestamp": "2024-11-04T12:00:00Z",
      "service": "web-api",
      "level": "INFO",
      "message": "Some example log",
      "host": "web-01",
      "tags": ["startup", "test"]
    },
    {
      "timestamp": "2024-11-04T12:01:00Z",
      "service": "web-api",
      "level": "ERROR",
      "message": "Some example error",
      "host": "web-01",
      "tags": ["database", "error"]
    }
  ]
}' $HOST logstream.LogIngestion/Ingest

echo "Test complete!"
