Qiita記事用のリポジトリ。更新中

# Vue.js + Go言語 + Pay.jp でカード決済できるECサイトを実装するハンズオン

そろそろカード決済の実装経験しとくかと思い、Pay.jpを眺めたらかなりドキュメントが充実してたので使いやすかった。今後、カード決済するサービスを作るのを見越して決済サービスをgRPCでマイクロサービス化し、Vue.jsから決済できるWEBサービスの実装をハンズオン形式で紹介します。

## 今回使う技術スタック

フロントエンドは Vue.js。サーバーサイドは Go言語で実装します。その他の今回使う技術は下記！

### RPC

RPCとは、RPC (Remote Procedure Call 別のアドレス空間にあるサブルーチンや手続きを実行することを可能にする技術)を実現するためにGoogleが開発したプロトコルです。Protocol Buffers を使ってデータをシリアライズし、高速な通信を実現できる点が特長です。さらっと出てきたが Protocol Buffer は構造化データをバイト列に変換(シリアライズ)する技術で、RPC でデータをやり取りする際などに用いられる。Protocol Buffer自体は新しい技術ではなく、2008年からオープンソース化している。

## gRPC

HTTP/2を標準でサポートしたRPCフレームワークで、。 デフォルトで対応しているProtocolBufferをgRPC用に書いた上で、サポートしている言語(Go Python Ruby Javaなど)にコード書き出しを行うと、異なる言語間でも型保証された通信を行うことができます。出来たのは最近で2015年にGoogleが発表した様子。

## Pay.jp
支払い機能をシンプルなAPIで実装できる！分かりやすい料金形態で決済を導入することが可能です。日本の企業が作ったサービスなので日本語の情報が豊富です。Go言語で実装する方法があまりまとまってないのでそこを今回は中心にお話しします。

## 今回目指す形

下記のような形を目指していきます。

ディレクトリ構造は
下記のようにしました。

```
.(GOPATH)
└── src
    └── vue-golang-payment-app
        ├── backend-api (フロントとやりとりするJSON API)
        ├── frontend-spa (Vue.jsで作るフロントエンド)
        └── payment-service (gRPCでつくる支払いマイクロサービス)
```

## まずはGo言語で gRPC に触れる

まずは上記の形をめざします。
payment-service というディレクトリにPay.jpのAPIを叩いて実際に支払いをするマイクロサービスをつくります。

### gRPC開発環境を作る

まずはgRPCを使えるようにするのと、protoファイルからGo言語のコードを自動生成するツールのインストール

```bash
$ go get -u google.golang.org/grpc
$ go get -u github.com/golang/protobuf/protoc-gen-go
```
ちなみにbinにパスが通っているか確認。これがないとコード自動生成時にエラーが出ます。

```bash
export PATH=$PATH:$GOPATH/bin
```

そして RPC するコードを生成する protoc コンパイラーをインストールします。下記で自分のOS等に合うものをダウンロードして展開します
https://github.com/google/protobuf/releases

そしてそれをパスの通っている場所におきます。僕は /usr/local/bin/ に起きました

```bash
$ cp ~/Download/protoc-3.6.0-osx-x86_64/bin/protoc /usr/local/bin/
```

ここでprotocコマンドが使えるか確認しておきましょう

```bash
$ protoc --version
```

### protoファイル作成

payment-service/proto/pay.proto を作ります。
そこで Protocol Buffers で使う gRPC service と request と response それぞれの型を定義します。

```proto
syntax = "proto3";

package paymentservice;

service PayManager {
  // 支払いを行うサービスを定義
  rpc Charge (PayRequest) returns (PayResponse) {}
}

// カード決済に使うパラメーターをリクエストに定義
message PayRequest {
  int64 id = 1;
  string token = 2;
  int64 amount = 3;
  string name = 4;
  string discription =5;
}

// カード決済後のレスポンスを定義
message PayResponse {
  bool paid = 1;
  bool captured = 3;
  int64 amount = 2;
}
```

これだけでRPCするためのGo言語のコードが自動的に作られます。

message宣言でリクエストやレスポンス等で使う型を定義します。
service宣言でサービスを定義し、定義したmessageを引数や返り値に定義できます。

ここまででGo言語のコードを生成する準備が整いました！早速下記を実行してみましょう

```bash
$ protoc --go_out=plugins=grpc:. proto/task_list.proto
```

これでGo言語で書かれたソースコード proto/pay.pd.go が出来ています。中身を確認してみましょう
下記の構造体やメソッドが確認できるはずです。

```go
// ... 省略

// message宣言で定義された PayRequest の定義から生成
type PayRequest struct {
	Id                   int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Token                string   `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
	Amount               int64    `protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Discription          string   `protobuf:"bytes,5,opt,name=discription,proto3" json:"discription,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// ... 省略

// message宣言で定義された PayResponse の定義から生成
type PayResponse struct {
	Paid                 bool     `protobuf:"varint,1,opt,name=paid,proto3" json:"paid,omitempty"`
	Captured             bool     `protobuf:"varint,3,opt,name=captured,proto3" json:"captured,omitempty"`
	Amount               int64    `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// ... 省略

// 先ほど定義したserviceから生成されたメソッド
func (c *payManagerClient) Charge(ctx context.Context, in *PayRequest, opts ...grpc.CallOption) (*PayResponse, error) {
	out := new(PayResponse)
	err := c.cc.Invoke(ctx, "/paymentservice.PayManager/Charge", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
// ...
```

基本、上のコードはいじりません。変更を加える時は.protoファイルを変更して、また先ほどの生成コマンドを叩けば更新されます。このメソッドや構造体を使って、サーバー側のコードを書いていきます。下記はgRPCで商品情報とカードのToken情報を受け取って、実際に支払いを行います。payment-service/server/server.go を作成します。

```go
package main

import (
	// ...

	gpay "vue-golang-payment-app/payment-service/proto"

	payjp "github.com/payjp/payjp-go/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement sa
type server struct{}

func (s *server) Charge(ctx context.Context, req *gpay.PayRequest) (*gpay.PayResponse, error) {
	// PAI の初期化
	pay := payjp.New(os.Getenv("PAYJP_TEST_SECRET_KEY"), nil)

	// 支払いをします。第一引数に支払い金額、第二引数に支払いの方法や設定を入れます。
	charge, err := pay.Charge.Create(int(req.Amount), payjp.Charge{
		// 現在はjpyのみサポート
		Currency: "jpy",
		// カード情報、顧客ID、カードトークンのいずれかを指定。今回はToken使います。
		CardToken: req.Token,
		Capture:   true,
		// 概要のテキストを設定できます
		Description: req.Name + ":" + req.Discription,
	})
	if err != nil {
		return nil, err
	}

	// 支払った結果から、Response生成
	res := &gpay.PayResponse{
		Paid:     charge.Paid,
		Captured: charge.Captured,
		Amount:   int64(charge.Amount),
	}
	return res, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	gpay.RegisterPayManagerServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Printf("gRPC Server started: localhost%s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

環境変数である PAYJP_TEST_SECRET_KEY は Pay.jp にアカウント登録をして手に入れます。下記にアクセスして管理画面 > API > テスト秘密鍵 で手に入ります。
https://pay.jp/

この鍵はテスト用のKeyなので実際に支払いが行われることはありません。この文字列を環境変数に登録しておきます。僕はdirenvを使っているので下記のように登録してます。

```bash
export PAYJP_TEST_SECRET_KEY=sk_test_**************
```

これで支払いの為のマイクロサービスが完成しました。
しかし、このサービスはHTTPでやりとりできません。RPCを話すためです。
なのでフロントとやりとりするAPIサーバーを作り、APIサーバーからChargeメソッドをgRPCで叩くようにしましょう。

### Go言語で API サーバー実装

上記のような構成をめざします。上には記載していませんが、フロントエンドに商品データを渡すAPIもつくります。
つまり下記の機能があるAPIサーバーを作ります。

```
GET    /api/v1/items             --> 商品を全て返す
GET    /api/v1/items/:id         --> id で指定された商品情報を返す
POST   /api/v1/charge/items/:id  --> id で指定された商品を購入する (Tokenを渡す必要あり)
```

ディレクトリ構成は下記。

```
.
├── Gopkg.lock
├── Gopkg.toml
├── Makefile
├── db 　　　　　　　　　(DB接続とDBとのやりとり)
│   ├── driver.go
│   └── repository.go
├── domai　　　　　　　　(entity層)
│   ├── item.go
│   └── token.go
├── handler　　　　　　　(いわゆるコントローラ)
│   ├── contecxt.go
│   ├── item.go
│   └── payment.go
├── infrastructure　　　(ルーターの設定)
│   └── router.go
├── init 　　　　　　　　 (DBの初期化用)
│   └── init.sql
├── main.go
└── vendor
```

```go
package main

import (
	"os"
	"vue-golang-payment-app/backend-api/infrastructure"
)

func main() {
	infrastructure.Router.Run(os.Getenv("API_SERVER_PORT"))
}
```

環境変数 API_SERVER_PORT はAPIを走らせるPORTを渡します。
下記のように僕は8888番ポートで走らせます。

```bash
export API_SERVER_PORT=:8888
```

さて、ここで読み込んでいる infrastructure パッケージを作りましょう
infrastructure/router.go ですね。JSON API をなのでJSONを扱うのが少し面倒なのでここでフレームワークの gin を使いましょう。
GitHubにあるドキュメントが参考になります。https://github.com/gin-gonic/gin

ちなみに"github.com/gin-contrib/cors"は gin 用のCORS設定パッケージです。今回は Vue から叩くのでこちらも使います。

```go
package infrastructure

import (
	"os"
	"vue-golang-payment-app/backend-api/handler"

	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
)

// Router - router api server
var Router *gin.Engine

func init() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{os.Getenv("CLIENT_CORS_ADDR")},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	router.GET("/api/v1/items", func(c *gin.Context) { handler.GetLists(c) })
	router.GET("/api/v1/items/:id", func(c *gin.Context) { handler.GetItem(c) })
	router.POST("/api/v1/charge/items/:id", func(c *gin.Context) { handler.Charge(c) })

	Router = router
}
```

init関数はこのパッケージを初期化する際に呼ばれます。参考は下記
[init関数のふしぎ #golang](https://qiita.com/tenntenn/items/7c70e3451ac783999b4f)

環境変数 CLIENT_CORS_ADDR も忘れずに！僕は http://localhost:8080 に設定してます(あとでVue.jsをlocalhost:8080で立ち上げるため)
これでmain.goで呼ばれていたRouterの設定が終わりました。続いて、APIへのリクエストがあった際の実際の処理を handler パッケージに書きましょう。
まずは 商品データをフロントに返す handler を handler/item.go に書きます。

```go
package handler

import (
	"net/http"
	"strconv"
	"vue-golang-payment-app/backend-api/db"
)

// GetLists - get all items
func GetLists(c Context) {
	res, err := db.SelectAllItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetItem - get item by id
func GetItem(c Context) {
	identifer, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	res, err := db.SelectItem(int64(identifer))
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, res)
}
```

そしてエラー処理は一旦全て簡易化のため 500エラーで返してます。
また、支払いを行う handler を書きます。ここではレクエストで渡された id から商品情報を所得して、最初に作った gRPCサーバーの Charge に引数として渡して実行します。

```go
package handler

import (
	// ...

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
	identifer, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}

	// id から item情報所得
	res, err := db.SelectItem(int64(identifer))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	greq := &gpay.PayRequest{
		Id:          int64(identifer),
		Token:       t.Token,
		Amount:      res.Amount,
		Name:        res.Name,
		Discription: res.Discription,
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
```

また、今回は ginフレームワークにアプリケーション全てを依存させないめに interface を使って gin.Context を抽象化します
handler/contecxt.go を作ります。

```go
package handler

// Context - context interface
type Context interface {
	Param(string) string
	Bind(interface{}) error
	Status(int)
	JSON(int, interface{})
}
```

これで handler は完成です。あとはハンドラーが扱うデータ型とMySQLとのやりとりを書きます。
つまりあとは dbパッケージと domainパッケージを作って終わりです。
まずは domainパッケージを作りましょう。ここでサーバーで使うデータ型をまとめます。

domain/item.go をつくります。

```go
package domain

// Item - set of item
type Item struct {
	ID          int64
	Name        string
	Discription string
	Amount      int64
}

// Items -set of item list
type Items []Item

```

domain/token.go もつくります。

```go 
package domain

//Payment - pay.jp payment parametor
type Payment struct {
	Token string
}
```

最後に dbパッケージをつくります。

db/driver.go

```go
package db

import (
	"database/sql"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Conn - sql connection handler
var Conn *sql.DB

// NewSQLHandler - init sql handler
func init() {
	user := os.Getenv("MYSQL_USER")
	pass := os.Getenv("MYSQL_PASSWORD")
	name := os.Getenv("MYSQL_DATABASE")

	dbconf := user + ":" + pass + "@/" + name
	conn, err := sql.Open("mysql", dbconf)
	if err != nil {
		panic(err.Error)
	}
	Conn = conn
}

```

コードの中にある環境変数は各自設定をお願いします。僕はローカル確認用に MYSQL_USER は root。MYSQL_DATABASE は itemsDB としました。
また init() でパッケージを初期化しています。この Conn を通して MySQL とやりとりします。db/repository.go を作ります。ここでは商品リストを全て返す処理とid指定で商品を一つ返す処理をつくります。

```go
package db

import (
	// ...
	"vue-golang-payment-app/backend-api/domain"
)

// SelectAllItems - select all
func SelectAllItems() (items domain.Items, err error) {
	stmt, err := Conn.Query("SELECT * FROM items")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	for stmt.Next() {
		var id int64
		var name string
		var discription string
		var amount int64
		if err := stmt.Scan(&id, &name, &discription, &amount); err != nil {
			continue
		}
		item := domain.Item{
			ID:          id,
			Name:        name,
			Discription: discription,
			Amount:      amount,
		}
		items = append(items, item)
	}
	return
}

// SelectItem - select post
func SelectItem(identifier int64) (item domain.Item, err error) {
	stmt, err := Conn.Prepare(fmt.Sprintf("SELECT * FROM items WHERE id = ? LIMIT 1"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer stmt.Close()
	var id int64
	var name string
	var discription string
	var amount int64
	err = stmt.QueryRow(identifier).Scan(&id, &name, &discription, &amount)
	if err != nil {
		return
	}
	item.ID = id
	item.Name = name
	item.Discription = discription
	item.Amount = amount
	return
}
```

お疲れ様です。APIサーバーができました。
ローカルのMySQLで "itemsDB"という名前のデータベースを CREATE して
init/init.sql を作成しましょう

```sql
DROP TABLE IF EXISTS items;

CREATE TABLE items (
  id integer AUTO_INCREMENT,
  name varchar(255),
  discription varchar(255),
  amount integer,
  primary key(id)
);

INSERT INTO items (name, discription, amount)
VALUES
  ('toy', 'test-toy', 2000);

INSERT INTO items (name, discription, amount)
VALUES
  ('game', 'test-game', 6000);
```

```bash
mysql -p -u root itemsDB < init/init.sql
```

これで下記の機能がある APIサーバーができました。

```
GET    /api/v1/items             --> 商品を全て返す
GET    /api/v1/items/:id         --> id で指定された商品情報を返す
POST   /api/v1/charge/items/:id  --> id で指定された商品を購入する (Tokenを渡す必要あり)
```

ちょっとここらで動くか確認しましょう。

```bash
curl -X GET localhost:8888/api/v1/items/1 
{"ID":1,"Name":"toy","Discription":"test-toy","Amount":2000}

curl -X GET localhost:8888/api/v1/items
[{"ID":1,"Name":"toy","Discription":"test-toy","Amount":2000},{"ID":2,"Name":"game","Discription":"test-game","Amount":6000}]

curl -X POST localhost:8888/api/v1/charge/items/1 
{"code":2,"message":"Charge.Create() parameter error: One of the following parameters is required: CustomerID, CardToken, Card"}
```

決済処理だけ必要なパラメータがないと言われています。Tokenは Vue で直接 Pay.jp とやりとりして手に入れます。
ではついに最終決戦。Vue.js でフロントを作ります。


### Vue.js でクライアントを実装しよう

くー長い！もう少しで完成です。

今回は vue-cli でプロジェクトのひな形を作ります。下記をプロジェクトのルート(GOPATH/src/vue-golang-payment-app)で実行

```bash
# vue-cli がなければインストーリ
$ npm install -g vue-cli
$ vue init webpack frontend-spa
```

frontend-spa/src で下記のように.vueファイルを作ります。

```bash
src
├── App.vue
├── components
│   ├── Home.vue
│   ├── Item.vue
│   └── ItemCard.vue
├── main.js
└── router
    └── index.js
```



