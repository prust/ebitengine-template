package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	input "github.com/quasilyte/ebitengine-input"
	"github.com/solarlune/dngn"
	"github.com/yohamta/ganim8/v2"
)

const (
	action_left input.Action = iota
	action_right
	action_up
	action_down
)

var (
	game_map  *dngn.Layout
	wall_img  *ebiten.Image
	door_img  *ebiten.Image
	floor_img *ebiten.Image
)

type Game struct {
	player       *Player
	player_anim  *ganim8.Animation
	screen_w     int
	screen_h     int
	input_system input.System
	player_input *input.Handler
}

type Player struct {
	x int
	y int
}

func (g *Game) Update() error {
	g.input_system.Update()
	if g.player_input.ActionIsPressed(action_left) {
		g.player.x -= 4
	} else if g.player_input.ActionIsPressed(action_right) {
		g.player.x += 4
	}
	if g.player_input.ActionIsPressed(action_up) {
		g.player.y -= 4
	} else if g.player_input.ActionIsPressed(action_down) {
		g.player.y += 4
	}
	g.player_anim.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()

	// draw the map
	map_select := game_map.Select()
	op := &ebiten.DrawImageOptions{}
	for cell := range map_select.Cells {
		// culling to only draw what's actually on-screen avoids cranking the player's fan
		if (cell.X+1)*16 <= g.screen_w && (cell.Y+1)*16 <= g.screen_h {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(cell.X*16), float64(cell.Y*16))
			op.GeoM.Scale(1, 1)
			// smooth anti-aliasing (and so ebitengine batches calls due to identical Filter param)
			// op.Filter = ebiten.FilterLinear

			v := game_map.Get(cell.X, cell.Y)
			if v == 'x' || v == '|' {
				screen.DrawImage(wall_img, op)
			} else if v == ' ' {
				screen.DrawImage(floor_img, op)
			} else if v == '#' {
				screen.DrawImage(door_img, op)
			}
		}
	}
	g.player_anim.Draw(screen, ganim8.DrawOpts(float64(g.screen_w)/2, float64(g.screen_h)/2, 0, 1, 1, 0.5, 0.5))
	ebitenutil.DebugPrintAt(screen, "player", g.player.x, g.player.y)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.screen_w, g.screen_h
}

func main() {
	ebiten.SetWindowSize(1280, 960)
	ebiten.SetWindowTitle("Ebitengine Template")

	// generate map
	game_map = dngn.NewLayout(100, 100)
	game_map.GenerateBSP(dngn.NewDefaultBSPOptions())

	// load images/spritesheets
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

	// initialize input system
	g.input_system.Init(input.SystemConfig{DevicesEnabled: input.AnyDevice})
	keymap := input.Keymap{
		action_left:  {input.KeyLeft, input.KeyA},
		action_right: {input.KeyRight, input.KeyD},
		action_up:    {input.KeyUp, input.KeyW},
		action_down:  {input.KeyDown, input.KeyS},
	}
	g.player_input = g.input_system.NewHandler(0, keymap)
	g.player = &Player{}

	g32 := ganim8.NewGrid(16, 32, 48, 128)
	g.player_anim = ganim8.New(character_img, g32.Frames("1-3", 3), 150*time.Millisecond)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
