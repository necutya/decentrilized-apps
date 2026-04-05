package main

import (
	"log"
	"net"
	"os"
	"time"

	pb "github.com/necutya/decentrilized_apps/lab1/gen/ticketpb"
	"github.com/necutya/decentrilized_apps/lab1/internal/grpcserver"
	"github.com/necutya/decentrilized_apps/lab1/internal/model"
	"github.com/necutya/decentrilized_apps/lab1/internal/repo"
	"github.com/necutya/decentrilized_apps/lab1/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// ─── Database ────────────────────────────────────────────────────────────
	dbPath := envOr("DB_PATH", "tickets.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Event{}, &model.Booking{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	seed(db)

	// ─── Repos & services ────────────────────────────────────────────────────
	userRepo    := repo.NewUserRepo(db)
	eventRepo   := repo.NewEventRepo(db)
	bookingRepo := repo.NewBookingRepo(db)

	authSvc   := service.NewAuthService(userRepo)
	ticketSvc := service.NewTicketService(eventRepo, bookingRepo)

	// ─── gRPC server ─────────────────────────────────────────────────────────
	grpcAddr := envOr("GRPC_ADDR", ":50051")
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcSrv, grpcserver.NewAuthServer(authSvc))
	pb.RegisterTicketServiceServer(grpcSrv, grpcserver.NewTicketServer(ticketSvc, authSvc))
	reflection.Register(grpcSrv)

	log.Printf("gRPC listening on %s", grpcAddr)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("grpc serve: %v", err)
	}
}

func seed(db *gorm.DB) {
	var count int64
	db.Model(&model.Event{}).Count(&count)
	if count > 0 {
		return
	}
	events := []model.Event{
		{Title: "Rock Legends Concert", Venue: "Madison Square Garden", Date: time.Date(2025, 7, 20, 19, 0, 0, 0, time.UTC), TotalSeats: 200, AvailableSeats: 200, Price: 79.99},
		{Title: "Jazz Night", Venue: "Blue Note Club", Date: time.Date(2025, 8, 5, 20, 30, 0, 0, time.UTC), TotalSeats: 80, AvailableSeats: 80, Price: 45.00},
		{Title: "Classical Symphony", Venue: "Carnegie Hall", Date: time.Date(2025, 9, 12, 18, 0, 0, 0, time.UTC), TotalSeats: 300, AvailableSeats: 300, Price: 120.00},
		{Title: "Stand-Up Comedy Night", Venue: "Laugh Factory", Date: time.Date(2025, 8, 18, 21, 0, 0, 0, time.UTC), TotalSeats: 120, AvailableSeats: 120, Price: 35.00},
		{Title: "Electronic Music Festival", Venue: "Warehouse 23", Date: time.Date(2025, 10, 1, 22, 0, 0, 0, time.UTC), TotalSeats: 500, AvailableSeats: 500, Price: 60.00},
	}
	db.Create(&events)
	log.Println("seeded events")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
