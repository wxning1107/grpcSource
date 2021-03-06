package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"grpcClient/certificattedRpc"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var (
	port = ":1107"
)

type myGrpcServer struct {
}

func (p *myGrpcServer) Hello(ctx context.Context, args *certificattedRpc.String) (*certificattedRpc.String, error) {
	reply := &certificattedRpc.String{Value: "hello" + args.GetValue()}

	return reply, nil
}

func (p *myGrpcServer) Channel(stream certificattedRpc.HelloService_ChannelServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		reply := &certificattedRpc.String{Value: "hello " + args.GetValue()}

		err = stream.Send(reply)
		if err != nil {
			return err
		}
	}
}

func startServer() {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(filter))
	certificattedRpc.RegisterHelloServiceServer(grpcServer, new(myGrpcServer))

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Panicf("could not listen on %s: %s", port, err)
	}

	if err := grpcServer.Serve(listener); err != nil {
		log.Panicf("grpc serve error: %s", err)
	}
}

func doClientWork() {
	conn, err := grpc.Dial("localhost:1107", grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := certificattedRpc.NewHelloServiceClient(conn)
	reply, err := client.Hello(context.Background(), &certificattedRpc.String{Value: " grpc"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)

	stream, err := client.Channel(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			num := rand.Intn(100)
			if err := stream.Send(&certificattedRpc.String{Value: "grpc" + strconv.Itoa(num)}); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		fmt.Println(reply.GetValue())

	}
}

// 拦截器
func filter(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	log.Println("filter:", info)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return handler(ctx, req)
}

func main() {
	go startServer()
	log.Println("Server is running")
	time.Sleep(time.Second)

	doClientWork()
}
