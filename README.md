Qiita記事用のリポジトリ。更新中

# Vue.js + Go言語 + gRPC + Pay.jp でカード決済マイクロサービスを実装するハンズオン

そろそろカード決済の実装経験しとくかと思い、Pay.jpを眺めたらかなりドキュメントが充実してたので使いやすかった。今後、カード決済するサービスを作るのを見越して決済サービスをgRPCでマイクロサービス化したので、ハンズオン形式で紹介します。

## そもそもRPCとは

RPCとは、RPC (Remote Procedure Call 別のアドレス空間にあるサブルーチンや手続きを実行することを可能にする技術)を実現するためにGoogleが開発したプロトコルです。Protocol Buffers を使ってデータをシリアライズし、高速な通信を実現できる点が特長です。さらっと出てきたが Protocol Buffer は構造化データをバイト列に変換(シリアライズ)する技術で、RPC でデータをやり取りする際などに用いられる。Protocol Buffer自体は新しい技術ではなく、2008年からオープンソース化している。

## それを踏まえてgRPCとは

HTTP/2を標準でサポートしたRPCフレームワークで、。 デフォルトで対応しているProtocolBufferをgRPC用に書いた上で、サポートしている言語(Go Python Ruby Javaなど)にコード書き出しを行うと、異なる言語間でも型保証された通信を行うことができます。出来たのは最近で2015年にGoogleが発表した様子。

## 今回目指す形

下記のような形を目指していきます。

## まずはGo言語でに触れる

### gRPC開発環境を作る

まずはgRPCを使えるようにするのと、protoファイルからGo言語のコードを自動生成するツールのインストール

```
$ go get -u google.golang.org/grpc
$ go get -u github.com/golang/protobuf/protoc-gen-go
```
ちなみにbinにパスが通っているか確認。これがないとコード自動生成時にエラーが出ます。

```
export PATH=$PATH:$GOPATH/bn
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

package paymentgateway;

// For grpc gateway
// import "google/api/annotations.proto";


service PayManager {
  rpc Charge (PayRequest) returns (PayResponse) {}
}

message PayRequest {
  string num = 1;
  string cvc = 2;
  string expm = 3;
  string expy = 4;
}

message PayResponse {
  bool paid = 1;
  bool captured = 3;
  int64 amount = 2;
}
```

ここまででGo言語のコードを生成する準備が整いました！早速下記を実行してみましょう

```
$ protoc --go_out=plugins=grpc:. proto/task_list.proto
```

これでGo言語で書かれたソースコードが proto /に出来ています。中身を確認してみましょう
下記のメソッドが確認できるはずです。

```go
// ...
```

クライアント側は下記のような実装になります。

```go
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
