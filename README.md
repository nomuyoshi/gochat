# gochat
golangで作ったチャットアプリ

# 環境変数

googleログインを使っているので、https://console.developers.google.com/ でclient_idとsecret_keyを生成。

```
export GOOGLE_CLIENT_ID=
export GOOGLE_SECRET_KEY=
export GOOGLE_STATE=[任意の文字列]
export GOOGLE_REDIRECT_URL=http://localhost:3000/auth/callback/google // ローカルの場合
```

