#!/bin/bash
set -e

echo "Getting Garage node ID..."
NODE_ID=$(docker exec garage /garage node id 2>/dev/null | head -1)
echo "Node ID: $NODE_ID"

echo "Applying layout..."
docker exec garage /garage layout assign -z dc1 -c 1G $NODE_ID
docker exec garage /garage layout apply --version 1

echo "Creating bucket..."
docker exec garage /garage bucket create films

echo "Creating access key..."
docker exec garage /garage key create cinema-key

echo "Granting permissions..."
KEY_ID=$(docker exec garage /garage key list | grep cinema-key | awk '{print $1}')
docker exec garage /garage bucket allow films --read --write --key $KEY_ID

echo ""
echo "Done! Save the Access Key ID and Secret Key above to your .env file:"
echo "STORAGE_ACCESS_KEY=..."
echo "STORAGE_SECRET_KEY=..."