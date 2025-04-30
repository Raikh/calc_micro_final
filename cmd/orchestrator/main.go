package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/raikh/calc_micro_final/internal/app"
	"github.com/raikh/calc_micro_final/internal/config"
	"github.com/raikh/calc_micro_final/internal/database"
	"github.com/raikh/calc_micro_final/internal/router"
	"github.com/raikh/calc_micro_final/model"
	pb "github.com/raikh/calc_micro_final/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskServer struct {
	pb.UnimplementedTaskServiceServer
	Config *config.Config
}

func NewServer(cfg *config.Config) *TaskServer {
	return &TaskServer{Config: cfg}
}

func main() {
	done := make(chan error, 2)

	ctx := context.Background()
	app := SetUp(ctx)
	go startGRPCServer(app.Cfg, done)
	go startHTTPServer(app.Cfg, done)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-done:
		log.Fatalf("server error: %v", err)
	case sig := <-quit:
		log.Printf("Shutting down servers due to signal: %v", sig)
		// For graceful shutdown, you can add additional cleanup logic here if needed
		// For example, you can close database connections, release resources, etc.
		// Close the database connection
	}
}

func SetUp(ctx context.Context) *app.App {
	application := new(app.App)
	application.Cfg = config.InitConfig()
	application.DB = database.InitDB(ctx, application.Cfg)
	return application
}

func areDependenciesCompleted(task *model.Task) bool {
	if task.Dependencies == nil {
		return true
	}

	dependentTasks, _ := model.GetTasksByIds(task.Dependencies)
	for idx, depTask := range dependentTasks {
		if !depTask.Completed {
			return false
		}
		updateTaskByDependency(task, idx, depTask.Result)
	}

	return true
}

func updateTaskByDependency(task *model.Task, index int, value *float64) {
	depsCount := len(task.Dependencies)
	if depsCount == 1 {
		if task.Arg1 == nil && *task.Arg2 != 0 {
			task.Arg1 = value
		} else {
			task.Arg2 = value
		}
	} else if depsCount == 2 {
		if index == 0 {
			task.Arg1 = value
		} else {
			if task.Arg1 == nil && *task.Arg2 != 0 {
				task.Arg1 = value
			} else {
				task.Arg2 = value
			}
		}
	}
}

func startGRPCServer(cfg *config.Config, done chan<- error) {
	grpcAddr := fmt.Sprintf("%s:%s", cfg.GetKey("APP_LISTENING_ADDRESS"), cfg.GetKey("APP_GRPC_LISTEN_PORT"))
	grpcLis, err := net.Listen("tcp", grpcAddr)

	if err != nil {
		log.Println("error starting tcp listener: ", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	taskServer := NewServer(cfg)
	pb.RegisterTaskServiceServer(grpcServer, taskServer)

	log.Printf("gRPC server listening on %s", grpcAddr)

	done <- grpcServer.Serve(grpcLis)
}

func startHTTPServer(cfg *config.Config, done chan<- error) {
	done <- router.InitRouter(cfg)
}

func (ts *TaskServer) Task(ctx context.Context, req *pb.Empty) (*pb.TaskResponse, error) {
	redistributionDelay := ts.Config.GetKey("TIME_TASK_IN_PROGRESS_REDISTRIBUTE")
	delay, err := strconv.Atoi(redistributionDelay)
	if err != nil {
		delay = 60
	}
	tasks, _ := model.GetTasksForProcessing(delay)
	for _, task := range tasks {
		if !task.Completed && areDependenciesCompleted(&task) {
			task.IsProcessing = true
			task.Update()

			w := &pb.TaskResponse{
				Id:            task.Id,
				Arg1:          *task.Arg1,
				Arg2:          *task.Arg2,
				Operation:     task.Operation,
				OperationTime: task.OperationTime,
			}
			return w, nil
		}
	}
	return nil, status.Error(codes.NotFound, "object not found")
}

func (ts *TaskServer) CalculatedTask(ctx context.Context, calculatedTask *pb.TaskResult) (*pb.Empty, error) {

	task, err := model.GetTaskById(calculatedTask.Id)
	if err != nil {
		return &pb.Empty{}, status.Error(codes.NotFound, "Task not found")
	}

	if task.Completed {
		return &pb.Empty{}, status.Error(codes.AlreadyExists, "Task already completed")
	}

	task.Result = &calculatedTask.Result
	task.Completed = true
	task.Update()

	if model.IsAllTasksCompleted(task.ExpressionId) {
		expression, _ := model.GetExpressionById(task.ExpressionId)
		expression.Result = task.Result
		expression.Status = "completed"
		expression.Update()
	}

	return &pb.Empty{}, nil
}
