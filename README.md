# ebitengine-template
General template repository that integrates the main libs as a starting point for game jams.

View online at https://prust.github.io/ebitengine-template/.

Libraries:
* [x] [ebitengine](https://ebitengine.org/) (game engine for Go)
* [x] [ganim8](https://github.com/yohamta0/ganim8-lib) (animation lib)
* [ ] [resolv](https://github.com/SolarLune/resolv) (collision lib)
* [x] [dngn](https://github.com/SolarLune/dngn) (random map generation lib)
* [ ] [grid](https://github.com/s0rg/grid) (pathfinding and line-of-sight testing)
* [ ] [ebitengine-input](https://github.com/quasilyte/ebitengine-input) (input lib)
* [ ] [kamera](https://github.com/setanarut/kamera) (camera lib)

# How to run the game from the terminal

```
$ go run -mod=mod .
```

# How to build the game for the browser

```
$ env GOOS=js GOARCH=wasm go build -mod=mod -o ebitengine-template.wasm github.com/prust/ebitengine-template
```

The `-mod=mod` is necessary to force go to not use the `/vendor/` folder for modules (we're using it for vendored web assets, i.e. `wasm_exec.js`).
