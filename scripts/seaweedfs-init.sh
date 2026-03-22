#!/bin/bash
set -e

ENDPOINT="http://localhost:8333"
ACCESS_KEY="${STORAGE_ACCESS_KEY}"
SECRET_KEY="${STORAGE_SECRET_KEY}"

if [ -z "$ACCESS_KEY" ] || [ -z "$SECRET_KEY" ]; then
  echo "Error: STORAGE_ACCESS_KEY and STORAGE_SECRET_KEY must be set"
  exit 1
fi

export NO_PROXY=localhost
export AWS_ACCESS_KEY_ID=$ACCESS_KEY
export AWS_SECRET_ACCESS_KEY=$SECRET_KEY
AWS_CMD="aws --endpoint-url $ENDPOINT"

echo "Creating buckets..."
$AWS_CMD s3 mb s3://showcase 2>/dev/null || echo "Bucket 'showcase' already exists"
$AWS_CMD s3 mb s3://media 2>/dev/null || echo "Bucket 'media' already exists"

echo "Setting up lifecycle policy for showcase/tmp/..."
LIFECYCLE_FILE=$(mktemp)
cat > "$LIFECYCLE_FILE" << 'EOF'
{
  "Rules": [
    {
      "ID": "tmp/",
      "Status": "Enabled",
      "Filter": {"Prefix": "tmp/"},
      "Expiration": {"Days": 1}
    }
  ]
}
EOF

$AWS_CMD s3api put-bucket-lifecycle-configuration \
  --bucket showcase \
  --lifecycle-configuration "file://$LIFECYCLE_FILE"

echo "Setting up lifecycle policy for media/originals/..."
cat > "$LIFECYCLE_FILE" << 'EOF'
{
  "Rules": [
    {
      "ID": "originals/",
      "Status": "Enabled",
      "Filter": {"Prefix": "originals/"},
      "Expiration": {"Days": 1}
    }
  ]
}
EOF

$AWS_CMD s3api put-bucket-lifecycle-configuration \
  --bucket media \
  --lifecycle-configuration "file://$LIFECYCLE_FILE"

rm "$LIFECYCLE_FILE"

echo "Done!"