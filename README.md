# sasakiy84.net/relive-multi-aggregator

YouTube のプレイリストからライブの動画情報を取得し、データベースに保存する。
保存したデータから、イベントごとに JSON ファイルを生成する。

## setup

```bash
> docker-compose up -d
> atlas migrate apply --url "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
```

## run

以下の環境変数を設定

```
YOUTUBE_API_KEY=xxxxx
POSTGRESQL_URL=postgresql://postgres:postgres@localhost:5432/postgres
```

そして、以下のコマンドを実行

```bash
> go run ./... retrieve <event-name> <playlist ID>
> go run ./... dump
```

## development

```bash
> sqlc generate
```
