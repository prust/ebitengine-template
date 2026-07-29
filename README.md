# ebitengine-template
General template repository that integrates the main libs as a starting point for game jams.

View online at https://prust.github.io/ebitengine-template/.

**Note:** By default, Go thinks that the /vendor/ folder is for its own vendored modules instead of vendored JS (`wasm_exec.js`), so to run locally use `go run -mod=mod .` instead of simply `go run .`

Libraries:
* [x] [ebitengine](https://ebitengine.org/) (game engine for Go)
* [x] [ganim8](https://github.com/yohamta0/ganim8-lib) (animation lib)
* [ ] [resolv](https://github.com/SolarLune/resolv) (collision lib)
* [ ] [dngn](https://github.com/SolarLune/dngn) (random map generation lib)
* [ ] [grid](https://github.com/s0rg/grid) (pathfinding and line-of-sight testing)
