package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/yohamta/ganim8/v2"
	"github.com/solarlune/dngn"
)

var (
	game_map  *dngn.Layout
	wall_img  *ebiten.Image
	door_img  *ebiten.Image
	floor_img *ebiten.Image
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

	// draw the map
	map_select := game_map.Select()
	for cell := range map_select.Cells {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(cell.X * 16), float64(cell.Y * 16))
		op.GeoM.Scale(1, 1)
		// smooth anti-aliasing (and so ebitengine batches calls due to identical Filter param)
		op.Filter = ebiten.FilterLinear

		v := game_map.Get(cell.X, cell.Y)
		if v == 'x' || v == '|' {
  		screen.DrawImage(wall_img, op)
		} else if v == ' ' {
			screen.DrawImage(floor_img, op)
		} else if v == '#' {
			screen.DrawImage(door_img, op)
		}

	}
	g.player_anim.Draw(screen, ganim8.DrawOpts(float64(g.screen_w)/2, float64(g.screen_h)/2, 0, 1, 1, 0.5, 0.5))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.screen_w, g.screen_h
}

func main() {
	ebiten.SetWindowSize(1280, 960)
	ebiten.SetWindowTitle("Ebitengine Template")

	game_map = dngn.NewLayout(100, 100)
	game_map.GenerateBSP(dngn.NewDefaultBSPOptions())

	var character_img, _, err = ebitenutil.NewImageFromFile("images/character_sheet.png")
	Check(err)
	wall_img, _, err = ebitenutil.NewImageFromFile("images/wall.png")
	Check(err)
	door_img, _, err = ebitenutil.NewImageFromFile("images/door.png")
	Check(err)
	floor_img, _, err = ebitenutil.NewImageFromFile("images/floor.png")
	Check(err)

	g := &Game{
		screen_w: 640,
		screen_h: 480,
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
