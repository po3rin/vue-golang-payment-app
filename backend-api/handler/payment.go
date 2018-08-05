package handler

import (
	"context"
	"net/http"
	"strconv"

	"vue-golang-payment-app/backend-api/db"
	"vue-golang-payment-app/backend-api/domain"
	gpay "vue-golang-payment-app/payment-service/proto"

	"google.golang.org/grpc"
)

var addr = "localhost:50051"

// Charge exec payment-service charge
func Charge(c Context) {
	//パラメータや body をうけとる
	t := domain.Payment{}
	c.Bind(&t)
	identifier, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	// id から item情報所得
	res, err := db.SelectItem(int64(identifier))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	greq := &gpay.PayRequest{
		Id:          int64(identifier),
		Token:       t.Token,
		Amount:      res.Amount,
		Name:        res.Name,
		Description: res.Description,
	}

	//IPアドレス(ここではlocalhost)とポート番号(ここでは5000)を指定して、サーバーと接続する
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusForbidden, err)
	}
	defer conn.Close()
	client := gpay.NewPayManagerClient(conn)
	gres, err := client.Charge(context.Background(), greq)
	if err != nil {
		c.JSON(http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gres)
}
