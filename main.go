package main

import (
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
	// Execute は *Template型のメソッド。(html/template)
	// パースされたテンプレート(html)を第二引数のデータオブジェクトに適用して
	// 第一引数のio.Writerに出力する。
	if err := t.templ.Execute(w, nil); err != nil {
		http.Error(w, "500 Internal Server Error.", http.StatusInternalServerError)
	}
}

func main() {
	// http.Handle の第二引数は Handler型。Handler型はServeHTTPメソッドを持つインターフェース
	// *templateHandlerにServeHTTPメソッドを定義したのは、http.Handleにわたすため。
	http.Handle("/", &templateHandler{filename: "chat.html"})
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
