# EbitenSP - 横スクロールアクションゲーム

Ebitengineを使用したAndroid向けの横スクロールアクションゲームです。

## ゲームの特徴

- キャラクターが自動的に右に進む
- 画面をタップ（またはクリック）でジャンプ
- レーザーで障害物を破壊
- 障害物に当たるとライフが減少
- スコアシステム
- **サウンド効果**:
  - ジャンプ時の効果音
  - レーザーが障害物に当たった時の効果音
  - 障害物が壊れた時の効果音

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

- **タップ/クリック**: ジャンプ + レーザー発射
- **タップ**: ゲームオーバー時にリスタート

## サウンド機能

ゲームには以下の効果音が含まれています：

- `jump.wav`: ジャンプ時の効果音（高音、短い）
- `hit.wav`: レーザーが障害物に当たった時の効果音（中音、短い）
- `destroy.wav`: 障害物が壊れた時の効果音（低音、長め）

音声ファイルが見つからない場合は、プログラムが自動的に代替音声を生成します。

## ファイル構成

- `main.go`: メインゲームロジック
- `go.mod`: Goモジュール設定
- `AndroidManifest.xml`: Androidアプリ設定
- `build.gradle`: Androidビルド設定
- `generate_sounds.py`: 音声ファイル生成スクリプト
- `jump.wav`: ジャンプ効果音
- `hit.wav`: ヒット効果音
- `destroy.wav`: 破壊効果音

## ライセンス

MIT License 