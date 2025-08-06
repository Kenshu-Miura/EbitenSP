package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	}
}

func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			*g = *NewGame()
		}
		return nil
	}

	// プレイヤーのジャンプ処理（マウスクリックまたはタッチ）
	if (inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || len(inpututil.JustPressedTouchIDs()) > 0) && g.player.onGround {
		g.player.velocity = -15
		g.player.onGround = false
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
	ebitenutil.DrawRect(screen, 0, groundY, screenWidth, screenHeight-groundY, color.RGBA{34, 139, 34, 255})

	// プレイヤーを描画
	ebitenutil.DrawRect(screen, g.player.x, g.player.y, playerSize, playerSize, color.RGBA{255, 0, 0, 255})

	// 障害物を描画
	for _, obstacle := range g.obstacles {
		ebitenutil.DrawRect(screen, obstacle.x, obstacle.y, obstacleSize, obstacleSize, color.RGBA{139, 69, 19, 255})
	}

	// スコアを表示
	scoreText := "Score: " + fmt.Sprintf("%d", g.score)
	ebitenutil.DebugPrint(screen, scoreText)

	if g.gameOver {
		ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-50, screenHeight/2-20)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to restart", screenWidth/2-80, screenHeight/2+20)
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
