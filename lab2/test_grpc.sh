#!/usr/bin/env bash
# gRPC flow test for lab2
# Requires: grpcurl (go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest)
# Start stack first: docker compose up -d  OR run both services manually.
#
#   api-service    → localhost:50052  (DeviceService)
#   worker-service → localhost:50053  (StatsService)

set -e

API="localhost:50052"
WORKER="localhost:50053"

echo "════════════════════════════════════════════════"
echo "  api-service  ($API)  — DeviceService"
echo "════════════════════════════════════════════════"

echo ""
echo "=== 1. List devices (seeded on startup) ==="
grpcurl -plaintext -d '{}' $API devicepb.DeviceService/ListDevices

echo ""
echo "=== 2. Create device ==="
CREATE=$(grpcurl -plaintext -d '{
  "name": "RTX 4090",
  "origin": "Taiwan",
  "price": 1599.99,
  "critical": true,
  "device_type": {
    "peripheral": false,
    "power_watts": 450,
    "has_cooler": true,
    "group": "multimedia",
    "ports": ["PCIe"]
  }
}' $API devicepb.DeviceService/CreateDevice)
echo "$CREATE"

DEVICE_ID=$(echo "$CREATE" | grep -o '"id": *"[0-9]*"' | grep -o '[0-9]*' | head -1)
echo "→ Created device id: $DEVICE_ID"

echo ""
echo "=== 3. Get device (id=$DEVICE_ID) ==="
grpcurl -plaintext -d "{\"id\": $DEVICE_ID}" $API devicepb.DeviceService/GetDevice

echo ""
echo "=== 4. Update device (id=$DEVICE_ID) ==="
grpcurl -plaintext -d "{
  \"id\": $DEVICE_ID,
  \"name\": \"RTX 4090 Ti\",
  \"origin\": \"Taiwan\",
  \"price\": 1799.99,
  \"critical\": true,
  \"device_type\": {
    \"peripheral\": false,
    \"power_watts\": 500,
    \"has_cooler\": true,
    \"group\": \"multimedia\",
    \"ports\": [\"PCIe\"]
  }
}" $API devicepb.DeviceService/UpdateDevice

echo ""
echo "=== 5. Delete device (id=$DEVICE_ID) ==="
grpcurl -plaintext -d "{\"id\": $DEVICE_ID}" $API devicepb.DeviceService/DeleteDevice

echo ""
echo "════════════════════════════════════════════════"
echo "  worker-service ($WORKER)  — StatsService"
echo "════════════════════════════════════════════════"

echo ""
echo "=== 6. GetStats — event counters per type and group ==="
grpcurl -plaintext -d '{}' $WORKER statspb.StatsService/GetStats

echo ""
echo "=== 7. ListDevices — devices stored in worker DB ==="
grpcurl -plaintext -d '{}' $WORKER statspb.StatsService/ListDevices

echo ""
echo "=== 8. GetDevice (id=1) — single device from worker DB ==="
grpcurl -plaintext -d '{"id": 1}' $WORKER statspb.StatsService/GetDevice

echo ""
echo "Done."
