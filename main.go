package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	input "github.com/quasilyte/ebitengine-input"
	"github.com/setanarut/kamera/v2"
	"github.com/solarlune/dngn"
	"github.com/solarlune/resolv"
	"github.com/yohamta/ganim8/v2"
)

const (
	action_left input.Action = iota
	action_right
	action_up
	action_down
)

var (
	cam       *kamera.Camera
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

	cam.LookAt(float64(g.player.x), float64(g.player.y))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()

	// get the camera bounds in world coords for culling purposes
	x1, y1 := cam.ScreenToWorld(0, 0)
	x2, y2 := cam.ScreenToWorld(g.screen_w, g.screen_h)
	cam_rect := resolv.NewRectangleFromCorners(x1, y1, x2, y2)

	// draw the map
	map_select := game_map.Select()
	op := &ebiten.DrawImageOptions{}
	for cell := range map_select.Cells {
		// TODO: create all these rects ONCE on map generation instead of on every frame
		cell_rect := resolv.NewRectangle(float64(cell.X*16), float64(cell.Y*16), 16, 16)

		// cull (only draw what's actually on-screen to avoid 100% CPU usage)
		// apparently Intersection() only returns whether the *borders* or the rects intersect w/ each-other
		// if one is entirely contained by the other, you have to also check IsContainedBy()
		if cell_rect.IsContainedBy(cam_rect) || !cam_rect.Intersection(cell_rect).IsEmpty() {
			op.GeoM.Reset()
			op.GeoM.Translate(float64(cell.X*16), float64(cell.Y*16))
			// smooth anti-aliasing (and so ebitengine batches calls due to identical Filter param)
			// op.Filter = ebiten.FilterLinear

			v := game_map.Get(cell.X, cell.Y)
			if v == 'x' || v == '|' {
				cam.Draw(wall_img, op, screen)
			} else if v == ' ' {
				cam.Draw(floor_img, op, screen)
			} else if v == '#' {
				cam.Draw(door_img, op, screen)
			}
		}
	}
	op.GeoM.Reset()
	op.GeoM.Translate(float64(g.player.x), float64(g.player.y))
	cam.Draw(g.player_anim.Frame(), op, screen)
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
	g.player = &Player{
		x: g.screen_w / 2,
		y: g.screen_h / 2,
	}

	g32 := ganim8.NewGrid(16, 32, 48, 128)
	g.player_anim = ganim8.New(character_img, g32.Frames("1-3", 3), 150*time.Millisecond)

	cam = kamera.NewCamera(float64(g.player.x), float64(g.screen_h/2), float64(g.screen_w), float64(g.screen_h))
	cam.ShakeEnabled = true
	cam.SmoothType = kamera.SmoothDamp
	cam.SmoothOptions.SmoothDampTimeX = 0.15

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
