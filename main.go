package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 400
	screenHeight = 600
	playerSize   = 40
	obstacleSize = 30
	laserWidth   = 4
	laserHeight  = 20
	laserSpeed   = 8
	minPlayerY   = -50 // プレイヤーの最小Y座標（上昇制限）
)

var audioContext *audio.Context

type Game struct {
	player               *Player
	obstacles            []*Obstacle
	lasers               []*Laser
	score                int
	gameOver             bool
	scrollX              float64
	obstacleTimer        int
	nextObstacleInterval int // 次の障害物までの間隔
	touchPressed         bool
	gravity              float64
	lives                int     // ライフ数
	maxLives             int     // 最大ライフ数
	gameTime             int     // ゲーム開始からの経過時間（フレーム数）
	scrollSpeed          float64 // 現在のスクロール速度
	baseSpeed            float64 // 基本スクロール速度
	audioContext         *audio.Context
	jumpSound            *audio.Player
	hitSound             *audio.Player
	destroySound         *audio.Player
	powerdownSound       *audio.Player
	gameOverSound        *audio.Player
}

type Player struct {
	x, y     float64
	velocity float64
	onGround bool
}

type Obstacle struct {
	x, y, width, height float64
	hitCount            int // レーザーが当たった回数
}

type Laser struct {
	x, y float64
}

// 音声ファイルを読み込む関数
func loadSound(filename string) (*audio.Player, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// ファイルの内容をメモリに読み込む
	audioBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// バイトデータからデコード
	decoded, err := wav.DecodeWithSampleRate(audioContext.SampleRate(), bytes.NewReader(audioBytes))
	if err != nil {
		return nil, err
	}

	player, err := audioContext.NewPlayer(decoded)
	if err != nil {
		return nil, err
	}

	return player, nil
}

// 簡単な音声データを生成する関数（音声ファイルが見つからない場合の代替）
func createSimpleSound(frequency float64, duration float64) *audio.Player {
	sampleRate := 44100
	numSamples := int(float64(sampleRate) * duration)

	// 簡単な正弦波を生成
	data := make([]byte, numSamples*2) // 16ビット = 2バイト
	for i := 0; i < numSamples; i++ {
		value := math.Sin(2 * math.Pi * frequency * float64(i) / float64(sampleRate))
		sample := int16(value * 0.3 * 32767) // 音量30%
		data[i*2] = byte(sample & 0xFF)
		data[i*2+1] = byte((sample >> 8) & 0xFF)
	}

	// 音声プレイヤーを作成
	player := audioContext.NewPlayerFromBytes(data)
	return player
}

func NewGame() *Game {
	// 音声ファイルを読み込み
	jumpSound, err := loadSound("se_shot_002.wav")
	if err != nil {
		log.Printf("ジャンプ音の読み込みエラー: %v", err)
		// 代替音声を生成
		jumpSound = createSimpleSound(800, 0.2)
	}

	hitSound, err := loadSound("se_hit_004.wav")
	if err != nil {
		log.Printf("ヒット音の読み込みエラー: %v", err)
		// 代替音声を生成
		hitSound = createSimpleSound(400, 0.1)
	}

	destroySound, err := loadSound("se_hit_005.wav")
	if err != nil {
		log.Printf("破壊音の読み込みエラー: %v", err)
		// 代替音声を生成
		destroySound = createSimpleSound(200, 0.3)
	}

	powerdownSound, err := loadSound("se_powerdown_006.wav")
	if err != nil {
		log.Printf("パワーダウン音の読み込みエラー: %v", err)
		// 代替音声を生成
		powerdownSound = createSimpleSound(300, 0.4)
	}

	gameOverSound, err := loadSound("jingle_original_die_003.wav")
	if err != nil {
		log.Printf("ゲームオーバー音の読み込みエラー: %v", err)
		// 代替音声を生成
		gameOverSound = createSimpleSound(150, 1.0)
	}

	return &Game{
		player: &Player{
			x:        50,
			y:        100, // 空中でスタート
			velocity: 0,
			onGround: false,
		},
		obstacles:            make([]*Obstacle, 0),
		lasers:               make([]*Laser, 0),
		score:                0,
		gameOver:             false,
		scrollX:              0,
		obstacleTimer:        0,
		nextObstacleInterval: rand.Intn(91) + 30, // 初期間隔をランダムに設定（30-120フレーム、つまり0.5-2秒）
		touchPressed:         false,
		gravity:              0.3, // ふんわりとした重力
		lives:                4,   // 初期ライフ数（4回まで当たれる）
		maxLives:             4,   // 最大ライフ数
		gameTime:             0,   // ゲーム開始からの経過時間
		scrollSpeed:          1.5, // 初期スクロール速度
		baseSpeed:            1.5, // 基本スクロール速度
		audioContext:         audioContext,
		jumpSound:            jumpSound,
		hitSound:             hitSound,
		destroySound:         destroySound,
		powerdownSound:       powerdownSound,
		gameOverSound:        gameOverSound,
	}
}

func (g *Game) Update() error {
	// タッチ状態の更新
	var touchIDs []ebiten.TouchID
	touchIDs = inpututil.AppendJustPressedTouchIDs(touchIDs)
	g.touchPressed = len(touchIDs) > 0

	// デバッグ用：タッチ状態をログ出力
	if g.touchPressed {
		fmt.Printf("タップ検出: touchIDs=%d\n", len(touchIDs))
	}

	// マウスクリックも検出
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		fmt.Printf("マウスクリック検出\n")
	}

	if g.gameOver {
		if g.touchPressed || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			*g = *NewGame()
			return nil
		}
		return nil
	}

	// プレイヤーのジャンプ処理（タップまたはマウスクリック）
	if g.touchPressed || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.player.velocity = -8 // ふんわりとしたジャンプ
		fmt.Printf("ジャンプ実行: velocity=%.1f\n", g.player.velocity)

		// ジャンプ音を再生
		if g.jumpSound != nil {
			g.jumpSound.Rewind()
			g.jumpSound.Play()
		}
	}

	// 重力の適用（ふんわりとした落下）
	g.player.velocity += g.gravity
	g.player.y += g.player.velocity

	// 上昇制限の適用
	if g.player.y < minPlayerY {
		g.player.y = minPlayerY
		g.player.velocity = 0 // 上昇制限に達したら速度をリセット

		// 画面外に到達した場合、ライフを1減らす
		g.lives--

		// パワーダウン音を再生
		if g.powerdownSound != nil {
			g.powerdownSound.Rewind()
			g.powerdownSound.Play()
		}

		if g.lives <= 0 {
			g.gameOver = true
			// ゲームオーバー音を再生
			if g.gameOverSound != nil {
				g.gameOverSound.Rewind()
				g.gameOverSound.Play()
			}
		}
	}

	// 画面下に落下したらゲームオーバー
	if g.player.y > screenHeight {
		g.gameOver = true
		// ゲームオーバー音を再生
		if g.gameOverSound != nil {
			g.gameOverSound.Rewind()
			g.gameOverSound.Play()
		}
		return nil
	}

	// スクロール処理（プレイヤーが右に進む）
	g.scrollX += g.scrollSpeed

	// 時間経過に応じてスクロール速度を調整
	g.gameTime++
	// より滑らかな速度上昇のため、フレームごとに少しずつ調整
	timeFactor := float64(g.gameTime) / 3600.0       // 60秒で最大速度に達するように調整
	g.scrollSpeed = g.baseSpeed + (timeFactor * 2.0) // 最大で3.5倍の速度まで上昇

	// 障害物の生成
	g.obstacleTimer++
	if g.obstacleTimer >= g.nextObstacleInterval { // ランダムな間隔で障害物を生成
		// 一度に1〜2個の障害物をランダムに生成
		obstacleCount := rand.Intn(2) + 1 // 1〜2個

		for i := 0; i < obstacleCount; i++ {
			// 重複しない位置を見つけるまで試行
			var obstacle *Obstacle
			maxAttempts := 50 // 最大試行回数

			for attempt := 0; attempt < maxAttempts; attempt++ {
				// ランダムな位置と大きさの障害物を生成
				obstacleWidth := float64(rand.Intn(40) + 20)        // 20-60のランダムな幅
				obstacleHeight := float64(rand.Intn(60) + 30)       // 30-90のランダムな高さ
				obstacleY := float64(rand.Intn(screenHeight - 100)) // 0からscreenHeight-100のランダムなY位置

				obstacle = &Obstacle{
					x:        screenWidth + 50,
					y:        obstacleY,
					width:    obstacleWidth,
					height:   obstacleHeight,
					hitCount: 0, // 障害物生成時はヒットカウントを0に
				}

				// 既存の障害物との重複チェック
				if !g.isObstacleOverlapping(obstacle) {
					break // 重複しなければループを抜ける
				}
			}

			// 障害物を追加（重複チェックを通過したもの、または最大試行回数に達したもの）
			if obstacle != nil {
				g.obstacles = append(g.obstacles, obstacle)
			}
		}
		g.obstacleTimer = 0
		// 次の障害物までの間隔をランダムに設定（30-120フレーム、つまり0.5-2秒）
		g.nextObstacleInterval = rand.Intn(91) + 30
	}

	// レーザーの発射（タップまたはマウスクリック時）
	if g.touchPressed || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.lasers = append(g.lasers, &Laser{
			x: g.player.x + playerSize,                  // プレイヤーの右側から発射
			y: g.player.y + playerSize/2 - laserWidth/2, // プレイヤーの中央から発射（横棒なのでlaserWidthを使用）
		})
		fmt.Printf("レーザー発射: 位置(%.1f, %.1f), レーザー数=%d\n", g.player.x+playerSize, g.player.y+playerSize/2-laserWidth/2, len(g.lasers))
	}

	// レーザーの移動
	for i := len(g.lasers) - 1; i >= 0; i-- {
		laser := g.lasers[i]
		laser.x += laserSpeed // 右に移動

		// 画面外に出たレーザーを削除
		if laser.x > screenWidth {
			g.lasers = append(g.lasers[:i], g.lasers[i+1:]...)
			continue
		}

		// 障害物との衝突判定
		for j := len(g.obstacles) - 1; j >= 0; j-- {
			obstacle := g.obstacles[j]
			if g.checkLaserCollision(laser, obstacle) {
				obstacle.hitCount++
				if obstacle.hitCount >= 2 { // 2回ヒットしたら削除
					g.obstacles = append(g.obstacles[:j], g.obstacles[j+1:]...)
					g.score += 3 // 障害物破壊で3点追加

					// 破壊音を再生
					if g.destroySound != nil {
						g.destroySound.Rewind()
						g.destroySound.Play()
					}
				} else {
					g.score++ // 通常のヒットで1点追加

					// ヒット音を再生
					if g.hitSound != nil {
						g.hitSound.Rewind()
						g.hitSound.Play()
					}
				}
				g.lasers = append(g.lasers[:i], g.lasers[i+1:]...)
				break // 衝突したら次のレーザーに進む
			}
		}
	}

	// 障害物の移動と衝突判定
	for i := len(g.obstacles) - 1; i >= 0; i-- {
		obstacle := g.obstacles[i]
		obstacle.x -= g.scrollSpeed // スクロール速度に合わせて移動

		// 画面外に出た障害物を削除
		if obstacle.x < -obstacle.width {
			g.obstacles = append(g.obstacles[:i], g.obstacles[i+1:]...)
			g.score++
			continue
		}

		// プレイヤーとの衝突判定
		if g.checkCollision(g.player, obstacle) {
			g.lives--

			// パワーダウン音を再生
			if g.powerdownSound != nil {
				g.powerdownSound.Rewind()
				g.powerdownSound.Play()
			}

			if g.lives <= 0 {
				g.gameOver = true
				// ゲームオーバー音を再生
				if g.gameOverSound != nil {
					g.gameOverSound.Rewind()
					g.gameOverSound.Play()
				}
			}
			// 衝突した障害物を削除
			g.obstacles = append(g.obstacles[:i], g.obstacles[i+1:]...)
		}
	}

	return nil
}

func (g *Game) checkCollision(player *Player, obstacle *Obstacle) bool {
	return player.x < obstacle.x+obstacle.width &&
		player.x+playerSize > obstacle.x &&
		player.y < obstacle.y+obstacle.height &&
		player.y+playerSize > obstacle.y
}

func (g *Game) checkLaserCollision(laser *Laser, obstacle *Obstacle) bool {
	return laser.x < obstacle.x+obstacle.width &&
		laser.x+laserHeight > obstacle.x &&
		laser.y < obstacle.y+obstacle.height &&
		laser.y+laserWidth > obstacle.y
}

// 障害物の重複チェック関数
func (g *Game) isObstacleOverlapping(newObstacle *Obstacle) bool {
	minSpacing := 10.0 // 障害物間の最小間隔

	for _, existingObstacle := range g.obstacles {
		// 四角形同士の重複チェック（境界ボックス）
		// 新しい障害物の境界
		newLeft := newObstacle.x
		newRight := newObstacle.x + newObstacle.width
		newTop := newObstacle.y
		newBottom := newObstacle.y + newObstacle.height

		// 既存の障害物の境界
		existingLeft := existingObstacle.x
		existingRight := existingObstacle.x + existingObstacle.width
		existingTop := existingObstacle.y
		existingBottom := existingObstacle.y + existingObstacle.height

		// 最小間隔を考慮した重複チェック
		// 四角形が重なっているか、または最小間隔以内にあるかをチェック
		if newRight+minSpacing > existingLeft &&
			newLeft < existingRight+minSpacing &&
			newBottom+minSpacing > existingTop &&
			newTop < existingBottom+minSpacing {
			return true
		}
	}

	return false
}

func (g *Game) drawLives(screen *ebiten.Image) {
	// 右上にライフを表示
	heartSize := 20.0
	heartSpacing := 25.0
	startX := screenWidth - 30 - (heartSpacing * float64(g.maxLives))
	startY := 10.0

	for i := 0; i < g.maxLives; i++ {
		x := startX + float64(i)*heartSpacing
		y := startY

		// ハートの色を決定（残りライフなら赤、失ったライフならグレー）
		var heartColor color.RGBA
		if i < g.lives {
			heartColor = color.RGBA{255, 0, 0, 255} // 赤色
		} else {
			heartColor = color.RGBA{128, 128, 128, 255} // グレー
		}

		// 簡単なハートマークを描画（四角形で代用）
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(heartSize), float32(heartSize), heartColor, false)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 背景を描画
	screen.Fill(color.RGBA{135, 206, 235, 255}) // 空色

	// プレイヤーを描画
	vector.DrawFilledRect(screen, float32(g.player.x), float32(g.player.y), float32(playerSize), float32(playerSize), color.RGBA{255, 0, 0, 255}, false)

	// 障害物を描画
	for _, obstacle := range g.obstacles {
		// ヒット回数に応じて色を変更
		var obstacleColor color.RGBA
		switch obstacle.hitCount {
		case 0:
			obstacleColor = color.RGBA{139, 69, 19, 255} // 茶色（初期）
		case 1:
			obstacleColor = color.RGBA{255, 165, 0, 255} // オレンジ（1回ヒット）
		default:
			obstacleColor = color.RGBA{255, 0, 0, 255} // 赤（2回ヒット、破壊直前）
		}
		vector.DrawFilledRect(screen, float32(obstacle.x), float32(obstacle.y), float32(obstacle.width), float32(obstacle.height), obstacleColor, false)
	}

	// レーザーを描画（横棒として）
	for _, laser := range g.lasers {
		vector.DrawFilledRect(screen, float32(laser.x), float32(laser.y), float32(laserHeight), float32(laserWidth), color.RGBA{255, 255, 0, 255}, false)
	}

	// スコアを表示
	scoreText := "Score: " + fmt.Sprintf("%d", g.score)
	ebitenutil.DebugPrint(screen, scoreText)

	// 現在のスクロール速度を表示
	speedText := "Speed: " + fmt.Sprintf("%.1f", g.scrollSpeed)
	ebitenutil.DebugPrintAt(screen, speedText, 0, 20)

	// 操作説明を表示
	ebitenutil.DebugPrintAt(screen, "Tap: Jump + Laser", 0, 40)

	// ライフを表示（右上に赤いハートマーク）
	g.drawLives(screen)

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-50, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Tap to restart", screenWidth/2-50, screenHeight/2+20)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// 音声プレイヤーをクリーンアップする関数
func (g *Game) Close() {
	if g.jumpSound != nil {
		g.jumpSound.Close()
	}
	if g.hitSound != nil {
		g.hitSound.Close()
	}
	if g.destroySound != nil {
		g.destroySound.Close()
	}
	if g.powerdownSound != nil {
		g.powerdownSound.Close()
	}
	if g.gameOverSound != nil {
		g.gameOverSound.Close()
	}
}

func main() {
	// 音声コンテキストを初期化（一度だけ）
	audioContext = audio.NewContext(44100)

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("EbitenSP - 空中アクション")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
