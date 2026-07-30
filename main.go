package main

import (
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	input "github.com/quasilyte/ebitengine-input"
	"github.com/setanarut/kamera/v2"
	"github.com/solarlune/dngn"
	"github.com/yohamta/ganim8/v2"
)

const (
	action_left input.Action = iota
	action_right
	action_up
	action_down
	sample_rate = 48000
	anim_rate   = (1000 / 8) * time.Millisecond // 8fps pixel art animation (looping 3-frame walk cycles)
)

var (
	cam       *kamera.Camera
	game_map  *dngn.Layout
	wall_img  *ebiten.Image
	door_img  *ebiten.Image
	floor_img *ebiten.Image
)

type Game struct {
	player            *Player
	player_anim       [4]*ganim8.Animation // an animation for each of the 4 directions
	player_dir        int                  // indexes the animation array
	screen_w          int
	screen_h          int
	input_system      input.System
	player_input      *input.Handler
	audio_context     *audio.Context
	player_walk_sound *audio.Player
}

type Player struct {
	x  int
	y  int
	dx int
	dy int
}

func (g *Game) Update() error {
	g.input_system.Update()
	was_walking := g.player.dx != 0 || g.player.dy != 0

	if g.player_input.ActionIsPressed(action_left) {
		g.player.dx = -4
	} else if g.player_input.ActionIsPressed(action_right) {
		g.player.dx = 4
	} else {
		g.player.dx = 0
	}

	if g.player_input.ActionIsPressed(action_up) {
		g.player.dy = -4
	} else if g.player_input.ActionIsPressed(action_down) {
		g.player.dy = 4
	} else {
		g.player.dy = 0
	}
	is_walking := g.player.dx != 0 || g.player.dy != 0

	g.player.x += g.player.dx
	g.player.y += g.player.dy

	if g.player_input.ActionIsJustPressed(action_down) {
		g.player_dir = 0
	} else if g.player_input.ActionIsJustPressed(action_right) {
		g.player_dir = 1
	} else if g.player_input.ActionIsJustPressed(action_left) {
		g.player_dir = 2
	} else if g.player_input.ActionIsJustPressed(action_up) {
		g.player_dir = 3
	} else if g.player_input.ActionIsJustReleased(action_down) || g.player_input.ActionIsJustReleased(action_right) || g.player_input.ActionIsJustReleased(action_left) || g.player_input.ActionIsJustReleased(action_up) {
		// if the player just released a key, change direction based on any other key that is still pressed
		if g.player_input.ActionIsPressed(action_down) {
			g.player_dir = 0
		} else if g.player_input.ActionIsPressed(action_right) {
			g.player_dir = 1
		} else if g.player_input.ActionIsPressed(action_left) {
			g.player_dir = 2
		} else if g.player_input.ActionIsPressed(action_up) {
			g.player_dir = 3
		}
	}

	if !was_walking && is_walking {
		g.player_walk_sound.Rewind()
		g.player_walk_sound.Play()
	} else if was_walking && !is_walking {
		g.player_walk_sound.Pause()
		g.player_anim[g.player_dir].GoToFrame(2)
	}

	if is_walking {
		g.player_anim[g.player_dir].Update()
	}

	cam.LookAt(float64(g.player.x), float64(g.player.y))
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Clear()

	// get the camera bounds in world coords for culling purposes
	x1, y1 := cam.ScreenToWorld(0, 0)
	x2, y2 := cam.ScreenToWorld(g.screen_w, g.screen_h)

	// draw the map
	map_select := game_map.Select()
	op := &ebiten.DrawImageOptions{}
	for cell := range map_select.Cells {
		// cull (only draw what's actually on-screen to avoid 100% CPU usage)
		if isRectangleOverlap(x1, y1, x2, y2, float64(cell.X*16), float64(cell.Y*16), float64(cell.X*16+16), float64(cell.Y*16+16)) {
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
	cam.Draw(g.player_anim[g.player_dir].Frame(), op, screen)
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

	g.audio_context = audio.NewContext(sample_rate)
	// wav files shouldn't be closed here b/c audio.Player manages stream state
	f, err := os.Open("audio/walk.wav")
	Check(err)
	d, err := wav.DecodeF32(f)
	Check(err)
	loop_walk := audio.NewInfiniteLoop(d, d.Length())
	Check(err)
	g.player_walk_sound, err = g.audio_context.NewPlayerF32(loop_walk)
	Check(err)

	g32 := ganim8.NewGrid(16, 32, 48, 128)
	g.player_anim[0] = ganim8.New(character_img, g32.Frames("1-3", 1), anim_rate)
	g.player_anim[1] = ganim8.New(character_img, g32.Frames("1-3", 2), anim_rate)
	g.player_anim[2] = ganim8.New(character_img, g32.Frames("1-3", 3), anim_rate)
	g.player_anim[3] = ganim8.New(character_img, g32.Frames("1-3", 4), anim_rate)

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

func isRectangleOverlap(x1 float64, y1 float64, x2 float64, y2 float64, x3 float64, y3 float64, x4 float64, y4 float64) bool {
	// If any of these are true, the rectangles do NOT overlap
	if y3 >= y2 || y4 <= y1 || x3 >= x2 || x4 <= x1 {
		return false
	}
	return true
}
