package main

import (
	"context"
	"fmt"

	"log"
	gpay "vue-golang-payment-app/payment-service/proto"

	"google.golang.org/grpc"
)

var addr = "localhost:50051"

func main() {
	//IPアドレス(ここではlocalhost)とポート番号(ここでは5000)を指定して、サーバーと接続する
	conn, err := grpc.Dial(addr, grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
	}

	//接続は最後に必ず閉じる
	defer conn.Close()

	c := gpay.NewPayManagerClient(conn)

	//サーバーに対してリクエストを送信する
	req := &gpay.PayRequest{
		Id:          1,
		Token:       "token",
		Amount:      3000,
		Name:        "toy",
		Description: "this is test toy",
	}
	resp, err := c.Charge(context.Background(), req)
	if err != nil {
		log.Fatalf("RPC error: %v", err)
	}
	log.Println(resp.Captured)
}
