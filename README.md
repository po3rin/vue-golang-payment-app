![pay-cover.png](https://qiita-image-store.s3.amazonaws.com/0/186028/3d6c3897-c9e2-b119-14e0-6b26bbef9096.png)

これは Qiita にも掲載されています。
https://qiita.com/po3rin/items/9638eab0a6a70faca86e

そろそろカード決済の実装経験しとくかと思い、PAY.JPを眺めたらかなりドキュメントが充実してたので使いやすかった。今後、カード決済するサービスを作るのを見越して決済サービスをgRPCでマイクロサービス化してみた。そのまま Vue.js と Go言語を使い、カード決済できるWEBサービスのサンプルを試しに作ってみた。その実装を簡略化してハンズオン形式で紹介します。

全コードは GitHub にあげてます。
https://github.com/po3rin/vue-golang-payment-app

## 得られるもの

* Vue.js + Go言語で簡易的なSPAをつくる経験
* gRPC で簡単なマイクロサービスをつくる経験
* PAY.JP を使ったカード決済の流れの理解

## 今回使う技術スタック

フロントエンドは Vue.js。サーバーサイドは Go言語で実装します。それ以外で今回使う技術は下記！

### PAY.JP

![pay.png](https://qiita-image-store.s3.amazonaws.com/0/186028/68b5acb8-3eb6-3f1b-31b0-1e7ccc320387.png)

支払い機能をシンプルなAPIで実装できる！分かりやすい料金形態で決済を導入することが可能です。日本の企業が作ったサービスなので日本語の情報が豊富です。Go言語で実装する方法があまりまとまってないので、今回はそこもお話しします。

### gRPC
![grpc-p.png](https://qiita-image-store.s3.amazonaws.com/0/186028/921bbdc2-a113-bf8d-7792-1dd68f82724a.png)

そもそもRPCとは、Remote Procedure Call と呼ばれる、別のアドレス空間にあるサブルーチンや手続きを実行することを可能にする技術です。

そして gRPC はHTTP/2を標準でサポートしたRPCフレームワークです。ProtocolBufferをgRPC用に書いた上で、サポートしている言語(Go Python Ruby Javaなど)にコード書き出しを行うと、異なる言語間でも型保証された通信を行うことができます。出来たのは最近で2015年にGoogleが発表した様子。

今回はgRPCを使って決済機能をマイクロサービス化します。これによってAPIサーバーへの影響を下げれる且つ、例えば今回の目指す形(下記に記載)であれば、APIサーバーをRubyで書き換えたいとなっても、RubyからGo言語の処理を叩けるので影響範囲を抑えれます。

## 今回目指す形

![pay-go-vue.png](https://qiita-image-store.s3.amazonaws.com/0/186028/9df053de-d9e6-0317-12ba-2beade53e587.png)

上記のような形を目指していきます。payment-service と item-service 間は gRPC で通信します。本当は商品情報を扱う処理もマイクロサービス化したかったのですが、ハンズオンとしては複雑になりそうなのでやめました。

データベースはMySQLを使います。ここに商品情報を格納し、フロントエンドに返したりします。

ちなみにPAY.JPでは直でカード情報をサーバーに渡して処理する形も昔はできましたが、現在は推奨されていません。クレジットカード情報をいかに所持せずに決済処理を提供するかが必要になっています。


ディレクトリ構造は下記のようにしました。

```
.(GOPATH)
└── src
    └── vue-golang-payment-app
        ├── backend-api -------(フロントエンドとやりとりするJSON API)
        ├── frontend-spa ------(Vue.jsで作るフロントエンド)
        └── payment-service ---(gRPCでつくるカード決済マイクロサービス)
```

## Go言語 + gRPC でカード決済サービスをつくる

![pay-grpc-pnly.png](https://qiita-image-store.s3.amazonaws.com/0/186028/baa65fa7-b279-0e81-e7aa-1e88c8c0a391.png)


まずは上記の形をめざします。
payment-service というディレクトリにPAY.JPのAPIを叩いて実際に支払いをするマイクロサービスをつくります。手順としては3ステップです。

![grpc3.png](https://qiita-image-store.s3.amazonaws.com/0/186028/ae79f294-323c-bce0-8e0d-714be3539846.png)

protoファイルからRPC通信で使うコードを自動生成し、そのコードを使ってサーバーを実装します。

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
  string description =　5;
}

// カード決済後のレスポンスを定義
message PayResponse {
  bool paid = 1;
  bool captured = 3;
  int64 amount = 2;
}
```

message宣言でリクエストやレスポンス等で使う型を定義します。
service宣言でサービスを定義し、定義したmessageを引数や返り値に定義できます。
これだけでRPCするためのGo言語のコードが自動的に作られます。

### Protocol Buffer から Go言語への書き出し

ここまででGo言語のコードを生成する準備が整いました！早速下記を実行してみましょう

```bash
$ protoc --go_out=plugins=grpc:. proto/pay.proto
```

これでGo言語で書かれたソースコード proto/pay.pd.go が出来ています。中身を確認してみましょう
下記の構造体や interface が確認できるはずです。

```go
// ... 省略

// message宣言で定義された PayRequest の定義から生成
type PayRequest struct {
	Id                   int64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Token                string   `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
	Amount               int64    `protobuf:"varint,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Name                 string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	Description          string   `protobuf:"bytes,5,opt,name=description,proto3" json:"description,omitempty"`
	// ...
}

// ... 省略

// message宣言で定義された PayResponse の定義から生成
type PayResponse struct {
	Paid                 bool     `protobuf:"varint,1,opt,name=paid,proto3" json:"paid,omitempty"`
	Captured             bool     `protobuf:"varint,3,opt,name=captured,proto3" json:"captured,omitempty"`
	Amount               int64    `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	// ...
}

// ... 省略

// 先ほど定義したserviceから生成された interface
type PayManagerServer interface {
	Charge(context.Context, *PayRequest) (*PayResponse, error)
}
```

基本、上のコードはいじりません。変更を加える時は.protoファイルを変更して、また先ほどの生成コマンドを叩けば更新されます。このinterfaceやstructを使って、サーバー側のコードを書いていきます。

### 実際の支払い処理を実装

先ほど生成された interface を満たすようにコード書いていきます。payment-service/server/server.go を作成します。

そしてついにここで PAY.JP がでてきます。PAY.JPの API を叩くクライアントのコードを実装しても良いですが、今回は https://github.com/payjp/payjp-go を使います。これは PAY.JP のAPI とのやりとりを抽象化してくれているパッケージです。詳しいドキュメントはありませんが、パッケージのGo言語のコードに日本語でコメントがついているので、使い方も簡単に理解できます。
下記はgRPCで商品情報とカードのToken情報を受け取って、実際に支払いを行います。

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
		Description: req.Name + ":" + req.Description,
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

環境変数である PAYJP_TEST_SECRET_KEY は PAY.JP にアカウント登録をして手に入れます。下記にアクセスして管理画面 > API > テスト秘密鍵 で手に入ります。
https://pay.jp/

この鍵はテスト用のKeyなので実際に支払いが行われることはありません。この文字列を環境変数に登録しておきます。僕はdirenvを使っているので下記のように登録してます。

```bash
export PAYJP_TEST_SECRET_KEY=sk_test_**************
```

これで支払いの為のマイクロサービスが完成しました。
しかし、このサービスはHTTPでやりとりできません。RPCを話すためです。もちろん curlコマンドも使えません。
ゆえにフロントエンドとやりとりするAPIサーバーを作り、APIサーバーからChargeメソッドをgRPCで叩くようにしましょう。

## Go言語で JSON API サーバー実装

![sever-pay.png](https://qiita-image-store.s3.amazonaws.com/0/186028/dcd85bee-568d-adb4-979b-76723ed70556.png)

上記のような構成をめざします。上には記載していませんが、DBから商品データをフロントエンドに渡すAPIもつくります。
つまり下記の機能があるAPIサーバーを作ります。

```
GET    /api/v1/items             --> 商品を全て返す
GET    /api/v1/items/:id         --> id で指定された商品情報を返す
POST   /api/v1/charge/items/:id  --> id で指定された商品を購入する (Tokenを渡す必要あり)
```

ディレクトリ構成は下記。今回はハンズオンなのでゆるいアーキテクチャにしてます。

```
.
├── Makefile
├── db-----------------(DB接続とDBとのやりとり)
│   ├── driver.go
│   └── repository.go
├── domai--------------(entity層)
│   ├── item.go
│   └── token.go
├── handler-------------(handler)
│   ├── contecxt.go
│   ├── item.go
│   └── payment.go
├── infrastructure------(ルーターの設定)
│   └── router.go
├── init----------------(DBの初期化用)
│   └── init.sql
└── main.go
```

### domain 層

まずは domainパッケージを作りましょう。ここでサーバーで使うデータ型をまとめます。

domain/item.go をつくります。

```go
package domain

// Item - set of item
type Item struct {
	ID          int64
	Name        string
	Description string
	Amount      int64
}

// Items -set of item list
type Items []Item

```

domain/token.go もつくります。

```go 
package domain

// Payment - PAY.JP payment parameter
type Payment struct {
	Token string
}
```

これでアプリケーションで使うデータ構造が定義できました。

### DBとやりとりする層

次にdbとやりとりするパッケージをつくります。

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
ここでのポイントは init() です。パッケージを初期化する際に呼ばれます。参考は下記
[init関数のふしぎ #golang](https://qiita.com/tenntenn/items/7c70e3451ac783999b4f)

ここで 初期化した Conn を通して MySQL とやりとりします。
db/repository.go を作ります。ここでは商品リストを全て返す処理とid指定で商品を一つ返す処理をつくります。

```go
package db

import (
	// ...
	"vue-golang-payment-app/backend-api/domain"
)

// SelectAllItems - select all posts
func SelectAllItems() (items domain.Items, err error) {
	stmt, err := Conn.Query("SELECT * FROM items")
	if err != nil {
		return
	}
	defer stmt.Close()
	for stmt.Next() {
		var id int64
		var name string
		var description string
		var amount int64
		if err := stmt.Scan(&id, &name, &description, &amount); err != nil {
			continue
		}
		item := domain.Item{
			ID:          id,
			Name:        name,
			Description: description,
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
	var description string
	var amount int64
	err = stmt.QueryRow(identifier).Scan(&id, &name, &description, &amount)
	if err != nil {
		return
	}
	item.ID = id
	item.Name = name
	item.Description = description
	item.Amount = amount
	return
}
```

Go言語の面白い点で、戻り値に名前をつけて定義した関数は return だけで終了しても構いません。これでもちゃんと item と err が返ります。
参考は下記
[Goは関数の戻り値に名前を付けられる / deferの驚き](http://imagawa.hatenadiary.jp/entry/2016/12/08/190000)

これでデータベースを操作するパッケージができました。

### router 部分
つづいて API の router 部分を作っていきましょう。
まずは 起点になる main.go を作ります。

```go
package main

import (
	// ...

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

環境変数 CLIENT_CORS_ADDR も忘れずに！僕は http://localhost:8080 に設定してます(あとで Vue.js を localhost:8080 で立ち上げるため)

これでmain.goで呼ばれていたRouterの設定が終わりました。続いて、APIへのリクエストがあった際の実際の処理を handler パッケージに書きましょう。

### handler 部分

今回は ginフレームワークにアプリケーション全てを依存させないめに interface を使って gin.Context を抽象化します
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

そして、商品データをフロントに返す handler を handler/item.go に書きます。

```go
package handler

import (
	// ...

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
また、支払いを行う handler を書きます。ここではレクエストで渡された id を使って DB から商品情報を取得して、最初に作った gRPCサーバーの Charge に引数として cardToken と商品情報と渡して実行します。

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

	// id から item情報取得
	res, err := db.SelectItem(int64(identifer))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	// gRPC サーバーに送る Request を作成
	greq := &gpay.PayRequest{
		Id:          int64(identifer),
		Token:       t.Token,
		Amount:      res.Amount,
		Name:        res.Name,
		Description: res.Description,
	}

	//IPアドレス(ここではlocalhost)とポート番号(ここでは50051)を指定して、サーバーと接続する
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		c.JSON(http.StatusForbidden, err)
	}
	defer conn.Close()
	client := gpay.NewPayManagerClient(conn)

	// gRPCマイクロサービスの支払い処理関数を叩く
	gres, err := client.Charge(context.Background(), greq)
	if err != nil {
		c.JSON(http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gres)
}
```

お疲れ様です。これで下記の機能がある APIサーバーができました。

```
GET    /api/v1/items             --> 商品を全て返す
GET    /api/v1/items/:id         --> id で指定された商品情報を返す
POST   /api/v1/charge/items/:id  --> id で指定された商品を購入する (Tokenを渡す必要あり)
```

### 動作確認

動作確認するためにMySQLにtestデータを入れます。
ローカルのMySQLで "itemsDB"という名前のデータベースを CREATE して
init/init.sql を作成しましょう

```sql
DROP TABLE IF EXISTS items;

CREATE TABLE items (
  id integer AUTO_INCREMENT,
  name varchar(255),
  description varchar(255),
  amount integer,
  primary key(id)
);

INSERT INTO items (name, description, amount)
VALUES
  ('toy', 'test-toy', 2000);

INSERT INTO items (name, description, amount)
VALUES
  ('game', 'test-game', 6000);
```

そしてこれをデータベースに注入します。

```bash
mysql -p -u root itemsDB < init/init.sql
```

ちょっとここらで動くか確認しましょう。

```bash
curl -X GET localhost:8888/api/v1/items/1 
{"ID":1,"Name":"toy","Description":"test-toy","Amount":2000}

curl -X GET localhost:8888/api/v1/items
[{"ID":1,"Name":"toy","Description":"test-toy","Amount":2000},{"ID":2,"Name":"game","Description":"test-game","Amount":6000}]

curl -X POST localhost:8888/api/v1/charge/items/1 
{"code":2,"message":"Charge.Create() parameter error: One of the following parameters is required: CustomerID, CardToken, Card"}
```

決済処理だけ必要なパラメータがないと言われています。最初に構成をお話しした通り、Tokenは Vue で直接 PAY.JP とやりとりして手に入れます。
ではついに最終決戦。Vue.js でフロントを作ります。


## Vue.js でクライアントを実装しよう

くー長い！もう少しで完成です。最終段階です。最初に見せた形までもっていきます。

![pay-go-vue.png](https://qiita-image-store.s3.amazonaws.com/0/186028/e4812cfb-2b49-1ce6-080a-5da0b264534f.png)


今回は vue-cli でプロジェクトのひな形を作ります。下記をプロジェクトのルート(GOPATH/src/vue-golang-payment-app)で実行

```bash
$ npm install -g vue-cli　　　　　　　　　　　　　　　　　　# vue-cli がなければインストール
$ vue init webpack frontend-spa　　　　# 何か色々聞かれるが全部 Enter で可能。vue-router は必ず入れておく。
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

Home.vue は商品リスト
Item.vue は商品の詳細ページ
ItemCard.vue は Home.vue で使う商品を表示するコンポーネントです。

画面はひどく殺風景ですが下のようになります。

HOME画面(商品リスト表示)
<img width="444" alt="スクリーンショット 2018-07-23 02.02.14.png" src="https://qiita-image-store.s3.amazonaws.com/0/186028/91be52e0-118c-6e9b-ef59-9e5fc72bb343.png">

商品詳細画面(商品リスト表示)
<img width="425" alt="スクリーンショット 2018-07-23 02.02.21.png" src="https://qiita-image-store.s3.amazonaws.com/0/186028/7b00f25e-9f50-c71b-6997-2f957d6ed4d4.png">

一旦サーバーとのやりとりに必要な axios モジュールを加えます。axios は Promise ベースの HTTPクライアントです。

```bash
$ npm install axios --save
```

また、今回 PAY.JP で カード情報を Token化するために https://github.com/ngs/vue-payjp-checkout を使います。
これは PAY.JP のカード情報入力コンポーネントを Vue.js で使えるようにしたものです。このような画面がひらくようになります。

<img width="783" alt="スクリーンショット 2018-07-22 16.43.02.png" src="https://qiita-image-store.s3.amazonaws.com/0/186028/c1606836-770a-541b-87e5-b2387df13632.png">

上のPAY.JPのクレジットカード入力フォームを使うと、開発者はクレジットカード番号に触れることなく決済機能が提供できるようになります。クレジットカード番号は盗み取られたりすると大きなリスクになります。そのため、今回はこの入力フォームを使いましょう。

```bash
$ npm install --save vue-payjp-checkout
```

vue-payjp-checkout　を Vue.js で使えるように src/main.js に一行追加します。

```js
// 省略...
import PayjpCheckout from 'vue-payjp-checkout'

Vue.config.productionTip = false

Vue.use(PayjpCheckout)
/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
})

```

これで payjp-checkout のコンポーネントが使えるようになりました。

### router 設定

次にvue-router でルーティングを正しく設定しましょう。
src/router/index.js を修正します。

```js
// ...省略

import Home from '@/components/Home'
import Item from '@/components/Item'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/items/:id',
      name: 'Item',
      component: Item
    }
  ]
})
```

:id は動的ルーティングのパラメータを表します。つまりここに商品番号が入ると、その商品詳細ページがみれるというつくりです。
この時点ではまだ Homeコンポーネントも Itemコンポーネントとも作ってないのでエラーが出ます。

### コンポーネント作成

次に Home.vue をつくります。最初に APIサーバーから商品一覧をとってきて、v-for でデータの数だけ後につくる ItemCard.vue に渡してあげます。

```html
<template>
<div class="hello">
  <ul>
    <li v-for="item in items" :key="item.ID" @click="pageto(item.ID)">
      <item-card :item="item"></item-card>
    </li>
  </ul>
</div>
</template>

<script>
import axios from 'axios'
import ItemCard from './ItemCard'
export default {
  name: 'Home',
  components: {
    ItemCard
  },
  data () {
    return {
      items: []
    }
  },
  methods: {
	// ページ移動
    pageto: function (id) {
      this.$router.push(`/items/${id}`)
    }
  },
  // 商品リストをすべてとってくる
  created () {
    axios.get('http://localhost:8888/api/v1/items').then(res => {
      this.items = res.data
    })
  }
}
</script>

<!-- css は vue init で最初に作られていた Helloworld.vue から転用-->
```

pageto イベントではクリックしたアイテムのidを使って、その商品の詳細ページに飛びます。
この時点ではItemCard.vue がないと言われます。


次に Home.vue で読み込んでいる ItemCard.vue を作りましょう。Home.vue から props で渡ってきた商品データを描写してるだけです。

```html
<template>
  <div class="itemcard">
    <h1>{{ item.Name }}</h1>
    <h2>{{ item.Description }}</h2>
    <h2>{{ item.Amount }}円</h2>
  </div>
</template>

<script>
export default {
  name: 'ItemCard',
  props: [
    'item'
  ]
}
</script>

<style scoped>
/* css は vue init で最初に作られていた Helloworld.vue の css も追加 */

.itemcard {
  border: solid 1px gray;
}
</style>
```

ここまでで商品一覧ページは完成しました。

あとは商品詳細ページをつくりましょう。商品のデータの表示はもちろん、ここではPAY.JP の API と直接やりとりして、カードの情報をトークン化して、そのトークンを使って、さきほどGo言語で作った API を叩きます。payjp-checkoutのコンポーネントは install した vue-payjp-checkout モジュールからもってきてます。

```html
<template>
  <div class="hello">
    <h1>{{ item.Name }}</h1>
    <h2>{{ item.Description }}</h2>
    <h2>{{ item.Amount }}円</h2>

    <payjp-checkout
      api-key="<< PAY.JPの管理画面にある公開テストKey >>"
      text="カードを情報を入力して購入"
      submit-text="購入確定"
      name-placeholder="田中 太郎"
      v-on:created="onTokenCreated"
      v-on:failed="onTokenFailed">
    </payjp-checkout>

    <p>{{ message }}</p>
    <router-link to="/">HOMEへ</router-link>
  </div>
</template>

<script>
import axios from 'axios'
export default {
  name: 'ItemCard',
  data () {
    return {
      item: {},
      message: ''
    }
  },
  created () {
	// urlで指定された動的パラメーターから商品除法をとってくる。
    axios.get(`http://localhost:8888/api/v1/items/${this.$route.params.id}`).then(res => {
      this.item = res.data
    })
  },
  beforeDestroy () {
    window.PayjpCheckout = null
  },
  methods: {
	// カードのToken化に成功したら呼ばれる。そのTokenでそのまま商品購入にうつる。
    onTokenCreated: function (res) {
      console.log(res.id)
	  const data = {Token: res.id}
      axios.post(`http://localhost:8888/api/v1/charge/items/${this.$route.params.id}`, data).then(res => {
        this.message = '商品の購入が完了しました！'
      })
	},
	// Token化に失敗したら呼ばれる。
    onTokenFailed: function (status, err) {
      console.log(status)
      console.log(err)
    }
  }
}
</script>
<!-- css は vue init で最初に作られていた Helloworld.vue から転用-->
```

<< PAY.JPの管理画面にある公開テストKey >> に　PAY.JP　の公開テストキーをいれるのを忘れずに！　管理画面から手に入ります。

vue-payjp-checkout と本家の Checkout でパラメーター名が違いますが、下記の Checkout リファレンスと vue-payjp-chackout の index.ts の中身を見比べれば vue-payjp-checkout のパラメーターがどういう意味なのか確認できます。
[PAY.JP Checkout 公式リファレンス](https://pay.jp/docs/checkout)
[vue-payjp-checkout の index.ts](https://github.com/ngs/vue-payjp-checkout/blob/master/src/index.ts)

また、ここでのポイントは beforeDestroy() で実行される window.PayjpCheckout = null です。これがないとページを移動したりするとカード登録ボタンが消えてしまいます。これは payjp-checkout のコンポネーネントがHTMLドキュメントの読み込みを起点として決済フォームを構築するためです。 そこでインスタンスが破棄される前に呼ばれる beforeDestroy() のライフサイクルで window.PayjpCheckout を一回空にして次のページ移動でもう一度コンポーネントを構築するようにしています。

参考にしたサイトでは Timeout で待ったりしていましたので、対処の仕方は色々あります。
[PAY.JPのチェックアウトのスクリプトをVue.jsのSPAで実装する](https://tackeyy.com/blog/posts/implement-payjp-checkout-with-vue-spa)

上のコードでは Token化した後すぐにその Token を使って支払いに入っていますが、もちろんToken化したあと確認ページへ遷移させるという実装も可能ですね。



## 動作確認

本当にお疲れ様です。実装は全て終わりました。
実際に動くか確認してみましょう。

backend API の立ち上げ

```bash
$ go run backend-api/main.go
```

gRPC サーバーの立ち上げ

```bash
$ go run payment-service/server/server.go
```

Vue で 作った SPA の立ち上げ

```bash
$ npm run dev
```

これで localhost:8080 にアクセスしてください。
(別のポート番号で立ち上がっている場合もあるので注意。その際は API の CORS の設定もそこに合わせます。)

カードデータはPAY.JPが用意しているテスト用の情報をいれます。

<img width="783" alt="スクリーンショット 2018-07-22 16.43.02.png" src="https://qiita-image-store.s3.amazonaws.com/0/186028/c1606836-770a-541b-87e5-b2387df13632.png">

購入確定ボタンを押せば「商品の購入が完了しました！」と画面にでているはずです。
ここまでいけば支払い情報が PAY.JP の管理画面で確認できます。

<img width="1121" alt="スクリーンショット 2018-07-22 17.54.21.png" src="https://qiita-image-store.s3.amazonaws.com/0/186028/49eda8c6-2cf2-ac29-2370-69759e659eca.png">

## まとめ

これで Vue.js + Go言語 + PAY.JP でカード決済できるWEBアプリケーションができました。めちゃくちゃ長くなりました。あとはUI整えたり商品管理画面つくったりでプロダクトに近づけていけば良さそうです。もしミスがあったらご指摘お願いします!!

全コードは GitHub にあげてます。
https://github.com/po3rin/vue-golang-payment-app

