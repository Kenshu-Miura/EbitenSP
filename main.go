package main

import (
	"fmt"
	"image/color"
	"log"

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
	groundY      = screenHeight - 100
)

type Game struct {
	player        *Player
	obstacles     []*Obstacle
	score         int
	gameOver      bool
	scrollX       float64
	obstacleTimer int
	touchPressed  bool
}

type Player struct {
	x, y     float64
	velocity float64
	onGround bool
}

type Obstacle struct {
	x, y float64
}

func NewGame() *Game {
	return &Game{
		player: &Player{
			x:        50,
			y:        groundY - playerSize,
			velocity: 0,
			onGround: true,
		},
		obstacles:     make([]*Obstacle, 0),
		score:         0,
		gameOver:      false,
		scrollX:       0,
		obstacleTimer: 0,
		touchPressed:  false,
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
		// デバッグ情報を出力
		if g.touchPressed {
			println("Touch detected in game over state!")
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			println("Space key pressed in game over state!")
		}

		// 直接タッチ検出も試す
		var directTouchIDs []ebiten.TouchID
		directTouchIDs = inpututil.AppendJustPressedTouchIDs(directTouchIDs)
		if len(directTouchIDs) > 0 {
			println("Direct touch detected in game over state!")
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			println("Mouse click detected in game over state!")
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || g.touchPressed || len(directTouchIDs) > 0 || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			println("Restarting game...")
			*g = *NewGame()
			return nil
		}
		return nil
	}

	// プレイヤーのジャンプ処理（マウスクリックまたはタッチ）
	if (inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || g.touchPressed) && g.player.onGround {
		g.player.velocity = -15
		g.player.onGround = false
		g.touchPressed = false // ジャンプ後にタッチ状態をリセット
	}

	// 重力の適用
	g.player.velocity += 0.8
	g.player.y += g.player.velocity

	// 地面との衝突判定
	if g.player.y >= groundY-playerSize {
		g.player.y = groundY - playerSize
		g.player.velocity = 0
		g.player.onGround = true
	}

	// スクロール処理
	g.scrollX += 2

	// 障害物の生成
	g.obstacleTimer++
	if g.obstacleTimer >= 120 { // 2秒ごとに障害物を生成
		g.obstacles = append(g.obstacles, &Obstacle{
			x: screenWidth + 50,
			y: groundY - obstacleSize,
		})
		g.obstacleTimer = 0
	}

	// 障害物の移動と衝突判定
	for i := len(g.obstacles) - 1; i >= 0; i-- {
		obstacle := g.obstacles[i]
		obstacle.x -= 2

		// 画面外に出た障害物を削除
		if obstacle.x < -obstacleSize {
			g.obstacles = append(g.obstacles[:i], g.obstacles[i+1:]...)
			g.score++
			continue
		}

		// プレイヤーとの衝突判定
		if g.checkCollision(g.player, obstacle) {
			g.gameOver = true
		}
	}

	return nil
}

func (g *Game) checkCollision(player *Player, obstacle *Obstacle) bool {
	return player.x < obstacle.x+obstacleSize &&
		player.x+playerSize > obstacle.x &&
		player.y < obstacle.y+obstacleSize &&
		player.y+playerSize > obstacle.y
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 背景を描画
	screen.Fill(color.RGBA{135, 206, 235, 255}) // 空色

	// 地面を描画
	vector.DrawFilledRect(screen, 0, groundY, screenWidth, screenHeight-groundY, color.RGBA{34, 139, 34, 255}, false)

	// プレイヤーを描画
	vector.DrawFilledRect(screen, float32(g.player.x), float32(g.player.y), float32(playerSize), float32(playerSize), color.RGBA{255, 0, 0, 255}, false)

	// 障害物を描画
	for _, obstacle := range g.obstacles {
		vector.DrawFilledRect(screen, float32(obstacle.x), float32(obstacle.y), float32(obstacleSize), float32(obstacleSize), color.RGBA{139, 69, 19, 255}, false)
	}

	// スコアを表示
	scoreText := "Score: " + fmt.Sprintf("%d", g.score)
	ebitenutil.DebugPrint(screen, scoreText)

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
	ebiten.SetWindowTitle("EbitenSP - 横スクロールアクション")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
