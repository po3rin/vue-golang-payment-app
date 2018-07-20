Qiita記事用のリポジトリ。更新中

# Vue.js + Go言語 + gRPC + Pay.jp でカード決済マイクロサービスを実装するハンズオン

そろそろカード決済の実装経験しとくかと思い、Pay.jpを眺めたらかなりドキュメントが充実してたので使いやすかった。今後、カード決済するサービスを作るのを見越して決済サービスをgRPCでマイクロサービス化したので、ハンズオン形式で紹介します。

## そもそもRPCとは

RPCとは、RPC (Remote Procedure Call 別のアドレス空間にあるサブルーチンや手続きを実行することを可能にする技術)を実現するためにGoogleが開発したプロトコルです。Protocol Buffers を使ってデータをシリアライズし、高速な通信を実現できる点が特長です。さらっと出てきたが Protocol Buffer は構造化データをバイト列に変換(シリアライズ)する技術で、RPC でデータをやり取りする際などに用いられる。Protocol Buffer自体は新しい技術ではなく、2008年からオープンソース化している。

## それを踏まえてgRPCとは

HTTP/2を標準でサポートしたRPCフレームワークで、。 デフォルトで対応しているProtocolBufferをgRPC用に書いた上で、サポートしている言語(Go Python Ruby Javaなど)にコード書き出しを行うと、異なる言語間でも型保証された通信を行うことができます。出来たのは最近で2015年にGoogleが発表した様子。

## 今回目指す形

下記のような形を目指していきます。

## まずはGo言語で gRPC に触れる

### gRPC開発環境を作る

まずはgRPCを使えるようにするのと、protoファイルからGo言語のコードを自動生成するツールのインストール

```
$ go get -u google.golang.org/grpc
$ go get -u github.com/golang/protobuf/protoc-gen-go
```
ちなみにbinにパスが通っているか確認。これがないとコード自動生成時にエラーが出ます。

```
export PATH=$PATH:$GOPATH/bin
```

そして RPC するコードを生成する protoc コンパイラーをインストールします。下記で自分のOS等に合うものをダウンロードして展開します
https://github.com/google/protobuf/releases

そしてそれをパスの通っている場所におきます。僕は /usr/local/bin/ に起きました
```
$ cp ~/Download/protoc-3.6.0-osx-x86_64/bin/protoc /usr/local/bin/
```

ここでprotocコマンドが使えるか確認しておきましょう

```
$ protoc --version
```

### protoファイル作成

まずは Protocol Buffers で使う gRPC service と request と response それぞれの型を定義します。

```proto
syntax = "proto3";

package paymentservice;

service PayManager {
  // 支払いを行うサービスを定義
  rpc Charge (PayRequest) returns (PayResponse) {}
}

// カード決済に使うパラメーターをリクエストに定義
message PayRequest {
  int64 amount = 1;
  string num = 2;
  string cvc = 3;
  string expm = 4;
  string expy = 5;
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
	Amount               int64    `protobuf:"varint,1,opt,name=amount,proto3" json:"amount,omitempty"`
	Num                  string   `protobuf:"bytes,1,opt,name=num,proto3" json:"num,omitempty"`
	Cvc                  string   `protobuf:"bytes,2,opt,name=cvc,proto3" json:"cvc,omitempty"`
	Expm                 string   `protobuf:"bytes,3,opt,name=expm,proto3" json:"expm,omitempty"`
	Expy                 string   `protobuf:"bytes,4,opt,name=expy,proto3" json:"expy,omitempty"`
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
	err := c.cc.Invoke(ctx, "/paymentgateway.PayManager/Charge", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, ni
}
// ...
```

基本、上のコードはいじりません。変更を加える時は.protoファイルを変更して、また先ほどの生成コマンドを叩けば更新されます。このメソッドや構造体を使って、サーバー側のコードを書いていきます。決済処理はまだ加えてません。単純に固定のレスポンスを返す用になっています。

```go
package main

import (
	// ...

	gpay "grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement sa
type server struct{}

func (s *server) Charge(ctx context.Context, req *gpay.PayRequest) (*gpay.PayResponse, error) {
	res := &gpay.PayResponse{
		Paid:     true,
		Captured: true,
		Amount:   30000,
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

これでサーバー側のコードが一旦動くようになりました。しかし、ここにアクセスするためにはclient側のコードも書く必要があります。というのも作ったサーバーはHTTPプロトコルでは動かないためです。故にcurlで動作確認もできません。先ほど作ったサーバーにリクエストを投げるクライアントを作成します。

```go
package main

import (
	//...

	gpay "grpc/proto"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var addr = "localhost:50051"

func main() {
	//サーバーと接続する
	conn, err := grpc.Dial(addr, grpc.WithInsecure())

	if err != nil {
		fmt.Println(err)
	}

	//接続は最後に必ず閉じる
	defer conn.Close()

	c := gpay.NewPayManagerClient(conn)

	//サーバーに対してリクエストを送信する
	req := &gpay.PayRequest{
		Amount: 3800,
		Num:  "4242424242424242",
		Cvc:  "123",
		Expm: "2",
		Expy: "2020",
	}
	resp, err := c.Charge(context.Background(), req)
	if err != nil {
		log.Fatalf("RPC error: %v", err)
	}
	log.Println(resp.Captured)
}
```

ここまでで gRPC の実装が完了しました。サーバーを起動しましょう

```bash
$ go run server/server.go
2018/07/19 20:45:58 gRPC Server started: localhost:50051
```

そしてクライアントを実行。レスポンスが返ってきます。

```bash
$ go run client/client.go
2018/07/19 20:46:06 true
```

これでgRPCでやり取りができました。

## Pay.jp で決済機能をサクッと実装

早速Pay.jpを使いましょう。まずは下記からアカウントを登録します。
https://pay.jp/
管理画面に入れたら >API をクリックすると APIキーの情報が手に入ります。
最初はテストモードなので実際の支払いが行われることなく、支払いのテストができます。

これで実装の準備ができました。server.go の Charge の中を書き換えます。

```go
func (s *server) Charge(ctx context.Context, req *gpay.PayRequest) (*gpay.PayResponse, error) {
	pay := payjp.New("<<テスト秘密鍵>>", nil)
	// 支払いをします。
	charge, _ := pay.Charge.Create(int(req.Amount), payjp.Charge{
		// 現在はjpyのみサポート。
		Currency: "jpy",
		// カード情報、顧客ID、カードトークンのいずれかを指定。今回はカードを選択。
		Card: payjp.Card{
			Number:   req.Num,
			CVC:      req.Cvc,
			ExpMonth: req.Expm,
			ExpYear:  req.Expy,
		},
		Capture: true,
		// 概要のテキストを設定できます。Pay.jpの管理画面で確認できます。
		Description: "Book: 'The Art of Community'",
		// 追加のメタデータを20件まで設定できます
		Metadata: map[string]string{
			"ISBN": "1449312063",
		},
	})
	if err != nil {
		return nil, err
	}
	res := &gpay.PayResponse{
		Paid:     charge.Paid,
		Captured: charge.Captured,
		Amount:   int64(charge.Amount),
	}
	return res, nil
}
```

Description や Metadata は今回、シンプル化のためにハードコーディングしてますが、もちろんRequestで渡せます。
では早速、支払いができるか試しましょう。

サーバー起動

```bash
$ go run server/server.go
2018/07/19 21:10:53 gRPC Server started: localhost:50051
```

支払いクライアント実行

```bash
$ go run client/client.go
2018/07/19 21:10:57 true
```

true が返ってきたら成功です。
Pay.jpの管理画面に行って支払いが本当に行われているか確認しましょう！

え、こんだけ？とビビるくらい簡単に支払いができました。

### Vue.js でクライアントを実装しよう