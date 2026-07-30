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
	"github.com/solarlune/resolv"
	"github.com/yohamta/ganim8/v2"
)

const (
	action_left input.Action = iota
	action_right
	action_up
	action_down
	sample_rate = 48000
	anim_rate   = time.Second / 8 // 8fps pixel art animation (looping 3-frame walk cycles)
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
	space             *resolv.Space
}

type Player struct {
	x    float64
	y    float64
	dx   float64
	dy   float64
	rect *resolv.ConvexPolygon // DRY violation w/ x,y -- should we solely use the collision lib rect?
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
	g.player.rect.Move(g.player.dx, g.player.dy)

	// filter to shapes near the player
	near_shapes := g.player.rect.SelectTouchingCells(4).FilterShapes()
	g.player.rect.IntersectionTest(resolv.IntersectionTestSettings{
		TestAgainst: near_shapes,
		OnIntersect: func(set resolv.IntersectionSet) bool {
			// back off from what we collided/intersected with
			g.player.rect.MoveVec(set.MTV)
			g.player.x += set.MTV.X
			g.player.y += set.MTV.Y
			// keep iterating (in case we're touching something else)
			return true
		},
	})

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
			if v == 'x' {
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

	g := &Game{
		screen_w: 640,
		screen_h: 480,
	}

	// generate map
	game_map = dngn.NewLayout(100, 100)
	game_map.GenerateBSP(dngn.NewDefaultBSPOptions())

	// create resolv (collision detection) rectangles for walls in the grid
	// trying a 32x32 "cell" size (for now) for performant collision checks
	g.space = resolv.NewSpace(100*16, 100*16, 32, 32)
	wall_select := game_map.Select().FilterByRune('x')
	for cell := range wall_select.Cells {
		wall_rect := resolv.NewRectangle(float64(cell.X)*16, float64(cell.Y)*16, 16, 16)
		g.space.Add(wall_rect)
	}

	// load images/spritesheets
	var character_img = loadImg("character_sheet.png")
	wall_img = loadImg("wall.png")
	door_img = loadImg("door.png")
	floor_img = loadImg("floor.png")

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
		x: float64(g.screen_w) / 2,
		y: float64(g.screen_h) / 2,
	}
	g.player.rect = resolv.NewRectangle(g.player.x, g.player.y, 16, 32)
	g.space.Add(g.player.rect)

	g.audio_context = audio.NewContext(sample_rate)

	walk_wav := loadWav("walk.wav")
	loop_walk := audio.NewInfiniteLoop(walk_wav, walk_wav.Length())
	var err error
	g.player_walk_sound, err = g.audio_context.NewPlayerF32(loop_walk)
	Check(err)

	// 16x32 frames, 3 frame columns and 4 frame rows
	g32 := ganim8.NewGrid(16, 32, 16*3, 32*4)
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

// wav files shouldn't be closed here b/c audio.Player manages stream state
func loadWav(filename string) *wav.Stream {
	f, err := os.Open("audio/" + filename)
	Check(err)
	wav_stream, err := wav.DecodeF32(f)
	Check(err)
	return wav_stream
}

func loadImg(filename string) *ebiten.Image {
	wall_img, _, err := ebitenutil.NewImageFromFile("images/" + filename)
	Check(err)
	return wall_img
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
