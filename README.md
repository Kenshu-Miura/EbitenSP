# EbitenSP - 横スクロールアクションゲーム

Ebitengineを使用したAndroid向けの横スクロールアクションゲームです。

## ゲームの特徴

- キャラクターが自動的に右に進む
- 画面をタップ（またはクリック）でジャンプ
- 障害物に当たるとゲームオーバー
- スコアシステム

## 必要な環境

- Go 1.21以上
- Android SDK
- gomobile

## ビルド手順

### 1. 依存関係のインストール

```bash
go mod tidy
```

### 2. gomobileのインストール

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
gomobile init
```

### 3. Android向けビルド

```bash
gomobile build -target=android -o ebitensp.apk .
```

### 4. デスクトップ向けビルド（テスト用）

```bash
go run .
```

## 操作方法

- **タップ/クリック**: ジャンプ
- **スペースキー**: ゲームオーバー時にリスタート

## ファイル構成

- `main.go`: メインゲームロジック
- `go.mod`: Goモジュール設定
- `AndroidManifest.xml`: Androidアプリ設定
- `build.gradle`: Androidビルド設定

## ライセンス

MIT License 