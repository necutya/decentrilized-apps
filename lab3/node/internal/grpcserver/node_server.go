package grpcserver

import (
	"context"

	"github.com/necutya/decentrilized_apps/lab3/node/gen/nodepb"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeServer struct {
	nodepb.UnimplementedNodeServiceServer
	svc *service.Service
}

func NewNodeServer(svc *service.Service) *NodeServer {
	return &NodeServer{svc: svc}
}

func (s *NodeServer) ListEvents(_ context.Context, _ *nodepb.ListEventsRequest) (*nodepb.ListEventsResponse, error) {
	events, err := s.svc.ListEvents()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list events: %v", err)
	}
	resp := &nodepb.ListEventsResponse{}
	for _, e := range events {
		resp.Events = append(resp.Events, &nodepb.Event{
			Id:             e.ID,
			Title:          e.Title,
			Venue:          e.Venue,
			Date:           e.Date.String(),
			AvailableSeats: e.AvailableSeats,
			TotalSeats:     e.TotalSeats,
			Price:          e.Price,
		})
	}
	return resp, nil
}

func (s *NodeServer) BookTicket(_ context.Context, req *nodepb.BookTicketRequest) (*nodepb.BookTicketResponse, error) {
	b, err := s.svc.BookTicket(req.EventId, req.Seats, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "book ticket: %v", err)
	}
	return &nodepb.BookTicketResponse{
		Booking: &nodepb.Booking{
			Id:         b.ID,
			EventId:    b.EventID,
			UserId:     b.UserID,
			Seats:      b.Seats,
			Status:     b.Status,
			TotalPrice: b.TotalPrice,
		},
	}, nil
}

func (s *NodeServer) CancelBooking(_ context.Context, req *nodepb.CancelBookingRequest) (*nodepb.CancelBookingResponse, error) {
	if err := s.svc.CancelBooking(req.Id, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "cancel booking: %v", err)
	}
	return &nodepb.CancelBookingResponse{Ok: true}, nil
}

func (s *NodeServer) ListBookings(_ context.Context, req *nodepb.ListBookingsRequest) (*nodepb.ListBookingsResponse, error) {
	bookings, err := s.svc.ListBookings(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list bookings: %v", err)
	}
	resp := &nodepb.ListBookingsResponse{}
	for _, b := range bookings {
		resp.Bookings = append(resp.Bookings, &nodepb.Booking{
			Id:         b.ID,
			EventId:    b.EventID,
			UserId:     b.UserID,
			Seats:      b.Seats,
			Status:     b.Status,
			TotalPrice: b.TotalPrice,
		})
	}
	return resp, nil
}
