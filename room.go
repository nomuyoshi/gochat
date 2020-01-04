package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	// ルームに参加しているクライアントに転送するためのメッセージを保持するチャネル
	forward chan []byte
	// ルームへの入退室するクライアントを管理するチャネル
	join  chan *client
	leave chan *client
	// ルームに参加中のクライアントを管理するmap
	clients map[*client]bool
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// joinチャネルからデータを受信した（クライアントが入室してきた）場合
			// clients mapに追加
			r.clients[client] = true
		case client := <-r.leave:
			// leaveチャネルからデータを受信した（クライアントが退室した）場合
			// clients mapからキーを削除し、sendチャネルをcloseする（今後メッセージが送信されないようにするため）
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// forwardチャネルからデータを受信した（clientのreadメソッドを通じてWebSocketからデータが読み込まれた）場合
			// ルームに参加しているクライアント全員のsendチャネルにメッセージを送信
			// sendチャネルにメッセージを送信すると、clientのwriteメソッドを呼び出してWebSocketに書き込まれる。
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージ送信
				default:
					// メッセージ送信失敗
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// ServeHTTP メソッドを定義することで、*room をHTTP Handlerとして扱えるようにする。
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	// WebSocketコネクションを取得失敗したら、終了
	// 成功したら、clientを生成し、roomに参加させる
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}
	client := &client{
		conn: conn,
		send: make(chan []byte, messageBufferSize),
		room: r,
	}

	r.join <- client
	// client.readメソッドの無限ループが終了したとき、後処理としてルームから退室させる
	defer func() { r.leave <- client }()
	// WebSocketの読み込みと書き込みは別々のgoroutineで実行する
	go client.write()
	client.read()
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}
