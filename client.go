package main

import (
	"github.com/gorilla/websocket"
	"time"
)

// client はチャットを行っているユーザを表す。
type client struct {
	// websocket connection
	conn *websocket.Conn
	// send はメッセージが送られるバッファ付きのチャネル
	// ここに受信したメッセージが入り、WebSocketを通じてブラウザに送信される
	send chan *message
	// room はクライアントが参加しているチャットルーム
	room     *room
	userName string
}

func (c *client) read() {
	for {
		var msg *message
		if err := c.conn.ReadJSON(&msg); err == nil {
			// websocket経由でメッセージデータを読み込み、メッセージをroom.forwardチャネルに送信
			msg.SentAt = time.Now()
			msg.Name = c.userName
			c.room.forward <- msg
		} else {
			// errorがあればbreakして無限ループを抜ける
			break
		}
	}
	// 無限ループを抜けたら（エラーが発生）コネクションを閉じる
	c.conn.Close()
}

func (c *client) write() {
	// 継続的にsendチャネルに溜まっているメッセージを取り出し、WriteMessageメソッドをつかって送信する
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			// 送信に失敗したら、breakして無限ループを抜ける
			break
		}
	}
	// 無限ループを抜けたら（エラーが発生）コネクションを閉じる
	c.conn.Close()
}
