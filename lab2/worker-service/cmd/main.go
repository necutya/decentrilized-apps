package main

import (
	"log"
	"net"
	"os"
	"time"

	pb "github.com/necutya/decentrilized_apps/lab2/worker-service/gen/statspb"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/consumer"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/grpcserver"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/model"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/repo"
	"github.com/necutya/decentrilized_apps/lab2/worker-service/internal/service"
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
	dbPath    := getenv("DB_PATH", "worker.db")
	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	grpcAddr  := getenv("GRPC_ADDR", ":50053")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Device{}, &model.EventStat{}, &model.LastProcessed{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	deviceRepo := repo.NewDeviceRepo(db)
	statsRepo  := repo.NewStatsRepo(db)
	workerSvc  := service.New(deviceRepo, statsRepo)

	// ── RabbitMQ consumer ────────────────────────────────────────────────────
	var cons *consumer.Consumer
	for i := 0; i < 5; i++ {
		cons, err = consumer.New(rabbitURL, workerSvc)
		if err == nil {
			break
		}
		log.Printf("rabbitmq connect attempt %d/5 failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if cons == nil {
		log.Fatalf("could not connect to rabbitmq: %v", err)
	}
	defer cons.Close()

	// ── gRPC server ──────────────────────────────────────────────────────────
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterStatsServiceServer(grpcSrv, grpcserver.NewStatsServer(statsRepo, deviceRepo))
	reflection.Register(grpcSrv)

	go func() {
		if err := cons.Consume(); err != nil {
			log.Fatalf("consume: %v", err)
		}
	}()

	log.Printf("gRPC listening on %s", grpcAddr)
	if err := grpcSrv.Serve(lis); err != nil {
		log.Fatalf("grpc serve: %v", err)
	}
}
