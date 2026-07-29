package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/yohamta/ganim8/v2"
)

type Game struct{
	player_anim *ganim8.Animation
  screen_w int
  screen_h int
}

func (g *Game) Update() error {
	g.player_anim.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()
	ebitenutil.DebugPrint(screen, "Hello, World!")
	g.player_anim.Draw(screen, ganim8.DrawOpts(float64(g.screen_w)/2, float64(g.screen_h)/2, 0, 1, 1, 0.5, 0.5))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.screen_w, g.screen_h
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")

	var character_img, _, err = ebitenutil.NewImageFromFile("images/character_sheet.png")
	Check(err)

	g := &Game{
		screen_w: 320,
		screen_h: 240,
	}

  g32 := ganim8.NewGrid(16, 32, 48, 128)
  g.player_anim = ganim8.New(character_img, g32.Frames("1-3", 3), 150 * time.Millisecond)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
