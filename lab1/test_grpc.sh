#!/usr/bin/env bash
# gRPC flow test for lab1
# Usage: ./test_grpc.sh
# Requires: grpcurl (go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest)

set -e

HOST="localhost:50051"

echo "=== 1. Register ==="
grpcurl -plaintext -d '{"username":"alice","password":"secret","email":"alice@example.com"}' \
  $HOST ticketpb.AuthService/Register

echo ""
echo "=== 2. Login ==="
LOGIN=$(grpcurl -plaintext -d '{"username":"alice","password":"secret"}' \
  $HOST ticketpb.AuthService/Login)
echo "$LOGIN"

TOKEN=$(echo "$LOGIN" | grep -o '"token": *"[^"]*"' | sed 's/.*": *"\(.*\)"/\1/')
echo "Token: $TOKEN"

echo ""
echo "=== 3. List Events ==="
grpcurl -plaintext -d '{}' \
  $HOST ticketpb.TicketService/ListEvents

echo ""
echo "=== 4. Get Event (id=1) ==="
grpcurl -plaintext -d '{"id":1}' \
  $HOST ticketpb.TicketService/GetEvent

echo ""
echo "=== 5. Book Ticket (event_id=1, seats=2) ==="
grpcurl -plaintext \
  -H "authorization: $TOKEN" \
  -d '{"event_id":1,"seats":2}' \
  $HOST ticketpb.TicketService/BookTicket

echo ""
echo "=== 6. List My Bookings ==="
grpcurl -plaintext \
  -H "authorization: $TOKEN" \
  -d '{}' \
  $HOST ticketpb.TicketService/ListMyBookings

echo ""
echo "=== 7. Cancel Booking (id=1) ==="
grpcurl -plaintext \
  -H "authorization: $TOKEN" \
  -d '{"id":1}' \
  $HOST ticketpb.TicketService/CancelBooking
