package main

import (
	"encoding/base64"
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type templateHandler struct {
	once     sync.Once // 関数を1度だけ実行するように制限できる。once.Do(func() {...})
	filename string
	templ    *template.Template // templateはhtml/templateで定義されているtype
}

// *templateHandler型にServeHTTPメソッドを定義
// メソッド呼び出し時に常に同じsync.Once を使うために、メソッドのレシーバはポインタである必要がある。
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}

	if authCookie, err := r.Cookie("auth"); err == nil {
		name, decodeErr := base64.StdEncoding.DecodeString(authCookie.Value)
		if decodeErr == nil {
			data["UserName"] = string(name)
		} else {
			http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
			return
		}
	}
	// Execute は *Template型のメソッド。(html/template)
	// パースされたテンプレート(html)を第二引数のデータオブジェクトに適用して
	// 第一引数のio.Writerに出力する。
	if err := t.templ.Execute(w, data); err != nil {
		http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
	}
}

func main() {
	var port = flag.String("port", ":3000", "port")
	flag.Parse()
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/auth/google", googleLoginHandler)
	http.HandleFunc("/auth/callback/google", googleCallbackHandler)
	// http.Handle の第二引数は Handler型。Handler型はServeHTTPメソッドを持つインターフェース
	// *templateHandlerにServeHTTPメソッドを定義したのは、http.Handleにわたすため。
	// MustAuthにより、&templateHandlerをラップしたauthHandlerを生成する
	// まずauthHandlerのServeHTTPが呼ばれ、ログイン判定を行い、ログイン済みの場合はtemplateHandlerの
	// ServeHTTPが呼ばれる。
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	// roomのserveHTTP内でWebSocketとのコネクション確立、room.joinへの追加、継続的なWebSocketのデータ読み込みが行われる
	r := newRoom()
	http.Handle("/room", r)
	// 別のgoroutineでrunメソッドを実行
	// runメソッドのselect節では、defaultがないのでjoin,leave,forwardのどれかのチャネルから
	// データを受信するまで待機する。
	go r.run()
	// webサーバ起動
	log.Print("Application starting. Port", *port)
	if err := http.ListenAndServe(*port, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
