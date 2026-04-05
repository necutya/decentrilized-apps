package grpcserver

import (
	"context"

	pb "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb"
	"github.com/necutya/decentrilized_apps/lab1/internal/model"
	"github.com/necutya/decentrilized_apps/lab1/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TicketServer struct {
	pb.UnimplementedTicketServiceServer
	svc  *service.TicketService
	auth *service.AuthService
}

func NewTicketServer(svc *service.TicketService, auth *service.AuthService) *TicketServer {
	return &TicketServer{svc: svc, auth: auth}
}

func (s *TicketServer) ListEvents(ctx context.Context, _ *pb.ListEventsRequest) (*pb.ListEventsResponse, error) {
	events, err := s.svc.ListEvents()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.ListEventsResponse{Events: toProtoEvents(events)}, nil
}

func (s *TicketServer) GetEvent(ctx context.Context, req *pb.GetEventRequest) (*pb.Event, error) {
	e, err := s.svc.GetEvent(req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "event not found")
	}
	return toProtoEvent(*e), nil
}

func (s *TicketServer) BookTicket(ctx context.Context, req *pb.BookTicketRequest) (*pb.Booking, error) {
	userID, _, err := s.tokenFromMeta(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	b, err := s.svc.BookTicket(userID, req.EventId, req.Seats)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return toProtoBooking(*b), nil
}

func (s *TicketServer) ListMyBookings(ctx context.Context, _ *pb.ListBookingsRequest) (*pb.ListBookingsResponse, error) {
	userID, _, err := s.tokenFromMeta(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	bookings, err := s.svc.ListMyBookings(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pbs []*pb.Booking
	for _, b := range bookings {
		pbs = append(pbs, toProtoBooking(b))
	}
	return &pb.ListBookingsResponse{Bookings: pbs}, nil
}

func (s *TicketServer) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*pb.CancelBookingResponse, error) {
	userID, _, err := s.tokenFromMeta(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.svc.CancelBooking(userID, req.Id); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &pb.CancelBookingResponse{Ok: true}, nil
}

func (s *TicketServer) tokenFromMeta(ctx context.Context) (uint, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return 0, "", status.Error(codes.Unauthenticated, "missing authorization")
	}
	return s.auth.ValidateToken(vals[0])
}

// ─── converters ──────────────────────────────────────────────────────────────

func toProtoEvent(e model.Event) *pb.Event {
	return &pb.Event{
		Id:             e.ID,
		Title:          e.Title,
		Venue:          e.Venue,
		Date:           e.Date.Format("2006-01-02 15:04"),
		AvailableSeats: e.AvailableSeats,
		TotalSeats:     e.TotalSeats,
		Price:          e.Price,
	}
}

func toProtoEvents(events []model.Event) []*pb.Event {
	out := make([]*pb.Event, len(events))
	for i, e := range events {
		out[i] = toProtoEvent(e)
	}
	return out
}

func toProtoBooking(b model.Booking) *pb.Booking {
	return &pb.Booking{
		Id:         b.ID,
		EventId:    b.EventID,
		EventTitle: b.Event.Title,
		Seats:      b.Seats,
		Status:     b.Status,
		TotalPrice: b.TotalPrice,
	}
}
