package main

import (
	"log"
	"net"
	"os"
	"time"

	pb "github.com/necutya/decentrilized_apps/lab2/api-service/gen/devicepb"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/grpcserver"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/model"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/publisher"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/repo"
	"github.com/necutya/decentrilized_apps/lab2/api-service/internal/service"
	"google.golang.org/grpc"
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
	dbPath := getenv("DB_PATH", "api.db")
	grpcAddr := getenv("GRPC_ADDR", ":50052")
	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Device{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	var pub *publisher.Publisher
	for i := 0; i < 5; i++ {
		pub, err = publisher.New(rabbitURL)
		if err == nil {
			break
		}
		log.Printf("rabbitmq connect attempt %d/5 failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if pub == nil {
		log.Fatalf("could not connect to rabbitmq: %v", err)
	}
	defer pub.Close()

	deviceRepo := repo.NewDeviceRepo(db)
	deviceSvc := service.NewDeviceService(deviceRepo, pub)

	seed(deviceSvc)

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	srv := grpc.NewServer()
	pb.RegisterDeviceServiceServer(srv, grpcserver.New(deviceSvc))
	reflection.Register(srv)

	log.Printf("gRPC server listening on %s", grpcAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func seed(svc *service.DeviceService) {
	existing, _ := svc.List(nil)
	if len(existing) > 0 {
		return
	}
	samples := []model.Device{
		{
			Name: "Logitech MX Keys", Origin: "USA", Price: 109.99, Critical: false,
			Peripheral: true, PowerWatts: 5, HasCooler: false, Group: "io",
			Ports: model.StringSlice{"USB"},
		},
		{
			Name: "Corsair RM850x", Origin: "USA", Price: 139.99, Critical: true,
			Peripheral: false, PowerWatts: 850, HasCooler: false, Group: "io",
			Ports: model.StringSlice{"COM"},
		},
		{
			Name: "Razer Kraken", Origin: "China", Price: 79.99, Critical: false,
			Peripheral: true, PowerWatts: 3, HasCooler: false, Group: "multimedia",
			Ports: model.StringSlice{"USB", "COM"},
		},
	}
	for i := range samples {
		if _, err := svc.Create(nil, &samples[i]); err != nil {
			log.Printf("seed error: %v", err)
		}
	}
	log.Printf("seeded %d devices", len(samples))
}
