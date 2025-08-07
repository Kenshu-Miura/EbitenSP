package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 400
	screenHeight = 600
	playerSize   = 40
	obstacleSize = 30
)

type Game struct {
	player        *Player
	obstacles     []*Obstacle
	score         int
	gameOver      bool
	scrollX       float64
	obstacleTimer int
	touchPressed  bool
	gravity       float64
	lives         int // ライフ数
	maxLives      int // 最大ライフ数
}

type Player struct {
	x, y     float64
	velocity float64
	onGround bool
}

type Obstacle struct {
	x, y, width, height float64
}

func NewGame() *Game {
	return &Game{
		player: &Player{
			x:        50,
			y:        100, // 空中でスタート
			velocity: 0,
			onGround: false,
		},
		obstacles:     make([]*Obstacle, 0),
		score:         0,
		gameOver:      false,
		scrollX:       0,
		obstacleTimer: 0,
		touchPressed:  false,
		gravity:       0.3, // ふんわりとした重力
		lives:         4,   // 初期ライフ数（4回まで当たれる）
		maxLives:      4,   // 最大ライフ数
	}
}

func (g *Game) Update() error {
	// タッチ状態の更新
	var touchIDs []ebiten.TouchID
	touchIDs = inpututil.AppendJustPressedTouchIDs(touchIDs)
	if len(touchIDs) > 0 {
		g.touchPressed = true
	}

	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || g.touchPressed || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			*g = *NewGame()
			return nil
		}
		return nil
	}

	// プレイヤーのジャンプ処理（マウスクリックまたはタッチ）
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || g.touchPressed {
		g.player.velocity = -8 // ふんわりとしたジャンプ
		g.touchPressed = false
	}

	// 重力の適用（ふんわりとした落下）
	g.player.velocity += g.gravity
	g.player.y += g.player.velocity

	// 画面下に落下したらゲームオーバー
	if g.player.y > screenHeight {
		g.gameOver = true
		return nil
	}

	// スクロール処理（プレイヤーが右に進む）
	g.scrollX += 1.5

	// 障害物の生成
	g.obstacleTimer++
	if g.obstacleTimer >= 90 { // 1.5秒ごとに障害物を生成
		// ランダムな位置と大きさの障害物を生成
		obstacleWidth := float64(rand.Intn(40) + 20)        // 20-60のランダムな幅
		obstacleHeight := float64(rand.Intn(60) + 30)       // 30-90のランダムな高さ
		obstacleY := float64(rand.Intn(screenHeight - 100)) // 0からscreenHeight-100のランダムなY位置

		g.obstacles = append(g.obstacles, &Obstacle{
			x:      screenWidth + 50,
			y:      obstacleY,
			width:  obstacleWidth,
			height: obstacleHeight,
		})
		g.obstacleTimer = 0
	}

	// 障害物の移動と衝突判定
	for i := len(g.obstacles) - 1; i >= 0; i-- {
		obstacle := g.obstacles[i]
		obstacle.x -= 1.5

		// 画面外に出た障害物を削除
		if obstacle.x < -obstacle.width {
			g.obstacles = append(g.obstacles[:i], g.obstacles[i+1:]...)
			g.score++
			continue
		}

		// プレイヤーとの衝突判定
		if g.checkCollision(g.player, obstacle) {
			g.lives--
			if g.lives <= 0 {
				g.gameOver = true
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
		vector.DrawFilledRect(screen, float32(obstacle.x), float32(obstacle.y), float32(obstacle.width), float32(obstacle.height), color.RGBA{139, 69, 19, 255}, false)
	}

	// スコアを表示
	scoreText := "Score: " + fmt.Sprintf("%d", g.score)
	ebitenutil.DebugPrint(screen, scoreText)

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

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("EbitenSP - 空中アクション")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
