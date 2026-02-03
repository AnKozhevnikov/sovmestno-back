#!/bin/sh
set -eu

until mc alias set myminio http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" 2>/dev/null; do
  echo "waiting for minio..."
  sleep 1
done

: "${MINIO_BUCKETS:=images creator-avatars venue-cover-photos venue-photos venue-logos event-covers}"
for b in $MINIO_BUCKETS; do
  mc mb --ignore-existing myminio/"$b"
done

echo "minio buckets created"