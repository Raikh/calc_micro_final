package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/raikh/calc_micro_final/internal/config"
	pb "github.com/raikh/calc_micro_final/proto"
)

func getTask(ctx context.Context, client pb.TaskServiceClient) *pb.TaskResponse {
	resp, err := client.Task(ctx, &pb.Empty{})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			// Ошибка не является gRPC ошибкой
			log.Fatalf("Unknown error: %v", err)
		}
		switch st.Code() {
		case codes.NotFound:
			return nil
		default:
			log.Printf("Error getting task: %v", err)
			return nil
		}
	}

	return resp
}

func computeTask(task *pb.TaskResponse) float64 {
	time.Sleep(time.Duration(task.OperationTime) * time.Millisecond)

	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}

func worker(ctx context.Context, client pb.TaskServiceClient) {
	for {
		task := getTask(ctx, client)
		if task == nil {
			time.Sleep(1 * time.Second)
			continue
		}
		result := computeTask(task)
		out := &pb.TaskResult{
			Id:     task.Id,
			Result: result,
		}
		_, err := client.CalculatedTask(ctx, out)
		if err != nil {
			log.Println("Error sending result:", err)
		}
	}
}

func main() {
	ctx := context.Background()
	cfg := config.InitConfig()

	grpcAddr := fmt.Sprintf("%s:%s", cfg.GetKey("CLIENT_GRPC_ARRT"), cfg.GetKey("CLIENT_GRPC_PORT"))
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("could not connect to grpc server: ", err)
		os.Exit(1)
	}
	// закроем соединение, когда выйдем из функции
	defer conn.Close()

	grpcClient := pb.NewTaskServiceClient(conn)
	computingPower, _ := strconv.Atoi(cfg.GetKey("CLIENT_COMPUTING_POWER"))
	if computingPower == 0 {
		computingPower = 2
	}

	for i := 0; i < computingPower; i++ {
		go worker(ctx, grpcClient)
	}

	select {}
}
