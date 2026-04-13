package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/necutya/decentrilized_apps/lab3/node/gen/nodepb"
	registrypb "github.com/necutya/decentrilized_apps/lab3/node/gen/registrypb"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/crypto"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/grpcserver"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/model"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/p2p"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/repo"
	"github.com/necutya/decentrilized_apps/lab3/node/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	nodeID := getenv("NODE_ID", "node1")
	nodeAddr := getenv("NODE_ADDR", "localhost:50060")
	grpcAddr := getenv("GRPC_ADDR", ":50060")
	registryAddr := getenv("REGISTRY_ADDR", "localhost:50070")
	dbPath := getenv("DB_PATH", "node.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Event{}, &model.Booking{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	r := repo.New(db)

	if r.EventsEmpty() {
		seedEvents(db)
	}

	signer := crypto.New()

	// Register with registry — retry 5 × 2s
	if err := registerWithRetry(registryAddr, nodeID, nodeAddr, signer.PublicKeyPEM(), 5, 2*time.Second); err != nil {
		log.Fatalf("register: %v", err)
	}

	broadcaster := p2p.NewBroadcaster(registryAddr, nodeID, signer)
	svc := service.New(r, broadcaster)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcSrv := grpc.NewServer()
	nodepb.RegisterNodeServiceServer(grpcSrv, grpcserver.NewNodeServer(svc))
	nodepb.RegisterPeerServiceServer(grpcSrv, grpcserver.NewPeerServer(r, registryAddr))
	reflection.Register(grpcSrv)

	log.Printf("node %s listening on %s", nodeID, grpcAddr)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func registerWithRetry(registryAddr, nodeID, nodeAddr, pubKey string, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		conn, err := grpc.NewClient(registryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			lastErr = err
			log.Printf("register attempt %d: dial error: %v", i+1, err)
			time.Sleep(delay)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = registrypb.NewRegistryServiceClient(conn).Register(ctx, &registrypb.RegisterRequest{
			Node: &registrypb.NodeInfo{
				Id:        nodeID,
				Address:   nodeAddr,
				PublicKey: pubKey,
			},
		})
		cancel()
		conn.Close()
		if err == nil {
			log.Printf("registered as %s at %s", nodeID, nodeAddr)
			return nil
		}
		lastErr = err
		log.Printf("register attempt %d: %v", i+1, err)
		time.Sleep(delay)
	}
	return lastErr
}

func seedEvents(db *gorm.DB) {
	events := []model.Event{
		{ID: 1, Title: "Rock Night Live", Venue: "Stadium Arena", Date: time.Date(2025, 6, 15, 20, 0, 0, 0, time.UTC), AvailableSeats: 200, TotalSeats: 200, Price: 49.99},
		{ID: 2, Title: "Jazz Evening", Venue: "Blue Note Club", Date: time.Date(2025, 7, 20, 19, 0, 0, 0, time.UTC), AvailableSeats: 80, TotalSeats: 80, Price: 35.00},
		{ID: 3, Title: "Classical Symphony", Venue: "City Concert Hall", Date: time.Date(2025, 8, 5, 18, 0, 0, 0, time.UTC), AvailableSeats: 150, TotalSeats: 150, Price: 60.00},
		{ID: 4, Title: "Pop Fest 2025", Venue: "Open Air Park", Date: time.Date(2025, 9, 1, 17, 0, 0, 0, time.UTC), AvailableSeats: 500, TotalSeats: 500, Price: 25.00},
		{ID: 5, Title: "Electronic Beats", Venue: "Warehouse 7", Date: time.Date(2025, 10, 12, 22, 0, 0, 0, time.UTC), AvailableSeats: 300, TotalSeats: 300, Price: 40.00},
	}
	db.Create(&events)
	log.Println("seeded 5 events")
}
