package handler

import (
	"context"
	"fmt"
	"net/http"

	gpay "vue-golang-payment-app/payment-service/proto"

	"google.golang.org/grpc"
)

var addr = "localhost:50051"

// Charge exec payment-service charge
func Charge(c Context) {
	//IPアドレス(ここではlocalhost)とポート番号(ここでは5000)を指定して、サーバーと接続する
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	client := gpay.NewPayManagerClient(conn)

	//サーバーに対してリクエストを送信する
	req := &gpay.PayRequest{
		Amount: 5000,
		Token:  "dsdsdsdsd",
	}
	resp, err := client.Charge(context.Background(), req)
	if err != nil {
		c.JSON(http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
