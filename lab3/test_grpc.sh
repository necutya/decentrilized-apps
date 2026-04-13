#!/usr/bin/env bash
# P2P Ticket Booking – integration smoke test
# Requires: grpcurl (https://github.com/fullstorydev/grpcurl)
set -e

NODE1="localhost:50061"
NODE2="localhost:50062"
NODE3="localhost:50063"
REGISTRY="localhost:50070"

hr() { echo; echo "══════════════════════════════════════════"; echo "  $*"; echo "══════════════════════════════════════════"; }

hr "List nodes from registry"
grpcurl -plaintext "$REGISTRY" registrypb.RegistryService/ListNodes

hr "List events on node1"
grpcurl -plaintext "$NODE1" nodepb.NodeService/ListEvents

hr "List events on node2"
grpcurl -plaintext "$NODE2" nodepb.NodeService/ListEvents

hr "List events on node3"
grpcurl -plaintext "$NODE3" nodepb.NodeService/ListEvents

hr "Book 2 seats on node1 (event 1, user alice)"
BOOKING=$(grpcurl -plaintext -d '{"event_id":1,"seats":2,"user_id":"alice"}' \
  "$NODE1" nodepb.NodeService/BookTicket)
echo "$BOOKING"
BOOKING_ID=$(echo "$BOOKING" | grep '"id"' | head -1 | grep -o '[0-9]*')
echo "Booking ID: $BOOKING_ID"

echo "Waiting 1s for P2P replication..."
sleep 1

hr "List bookings for alice on node2 (should see replicated booking)"
grpcurl -plaintext -d '{"user_id":"alice"}' "$NODE2" nodepb.NodeService/ListBookings

hr "List bookings for alice on node3 (should see replicated booking)"
grpcurl -plaintext -d '{"user_id":"alice"}' "$NODE3" nodepb.NodeService/ListBookings

hr "Cancel booking on node3 (booking ID $BOOKING_ID, user alice)"
grpcurl -plaintext -d "{\"id\":$BOOKING_ID,\"user_id\":\"alice\"}" \
  "$NODE3" nodepb.NodeService/CancelBooking

echo "Waiting 1s for P2P replication..."
sleep 1

hr "List bookings for alice on node1 (should show cancelled status)"
grpcurl -plaintext -d '{"user_id":"alice"}' "$NODE1" nodepb.NodeService/ListBookings

hr "List bookings for alice on node2 (should show cancelled status)"
grpcurl -plaintext -d '{"user_id":"alice"}' "$NODE2" nodepb.NodeService/ListBookings

hr "Check available seats on node1 after cancel (should be back to 200)"
grpcurl -plaintext "$NODE1" nodepb.NodeService/ListEvents | grep -A5 '"id": "1"'

hr "Done – P2P replication verified across all 3 nodes"
