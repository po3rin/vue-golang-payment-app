# Go言語 + Python + gRPC で自然言語処理API 実装ハンズオン

APIサーバーはGo言語で実装したいけど、統計処理などの役割は過去にPythonで作ったコードに処理させたいという想定でgRPCのハンズオン記事です。自然言語処理してくれるAPIを例にgRPCに触れていきます。

## そもそもRPCとは

RPCとは、RPC (Remote Procedure Call 別のアドレス空間にあるサブルーチンや手続きを実行することを可能にする技術)を実現するためにGoogleが開発したプロトコルです。Protocol Buffers を使ってデータをシリアライズし、高速な通信を実現できる点が特長です。さらっと出てきたが Protocol Buffer は構造化データをバイト列に変換(シリアライズ)する技術で、RPC でデータをやり取りする際などに用いられる。Protocol Buffer自体は新しい技術ではなく、2008年からオープンソース化している。

## それを踏まえてgRPCとは

HTTP/2を標準でサポートしたRPCフレームワークで、。 デフォルトで対応しているProtocolBufferをgRPC用に書いた上で、サポートしている言語(Go Python Ruby Javaなど)にコード書き出しを行うと、異なる言語間でも型保証された通信を行うことができます。出来たのは最近で2015年にGoogleが発表した様子。

## 今回目指す形

下記のような形を目指していきます。

## まずはGo言語でgRPCに触れる

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

package tasklistgateway;

// Service - ここで定義したメソッドがGo言語で使える関数に変換されます
service TaskManager {
  rpc GetTask (GetTaskRequest) returns (Task) {}
}

// response の型を定義しています
message Task {
  int32 id = 1;
  string title = 2;
}

// request の型を定義しています
message GetTaskRequest {
  int32 id = 1;
}
```

簡単なrpcサービスを定義しました。これは Taskを返すメソッドで、idを使ってTaskを受け取れます
ここまででGo言語のコードを生成する準備が整いました！早速下記を実行してみましょう

```
$ protoc --go_out=plugins=grpc:. proto/task_list.proto
```

これでGo言語で書かれたソースコードが proto /に出来ています。中身を確認してみましょう
下記のメソッドが確認できるはずです。

```go
// ...

func (c *taskManagerClient) GetTask(ctx context.Context, in *GetTaskRequest, opts ...grpc.CallOption) (*Task, error) {
	out := new(Task)
	err := c.cc.Invoke(ctx, "/tasklistgateway.TaskManager/GetTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ...
```

こいつを自分で作るコードの中に組み込んでいきます。このメソッドを使う際には
下記のような実装になります。

```go
package main

import (
	"context"
	"errors"
	"log"
	"net"

	tlpb "grpc/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

type server struct{}

func (s *server) GetTask(ctx context.Context, req *tlpb.GetTaskRequest) (*tlpb.Task, error) {
	log.Println("GetTask in gPRC server")
	var task = &tlpb.Task{
		Id:    1,
		Title: "Hello gRPC server",
	}
	if req.Id == task.Id {
		return task, nil
	}
	return nil, errors.New("Not find Task")
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	tlpb.RegisterTaskManagerServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	log.Printf("gRPC Server started: localhost%s\n", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

```
