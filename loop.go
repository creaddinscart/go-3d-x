package x3d

import "time"

// GameLoop runs a fixed-timestep update with a render callback, mirroring the
// classic "fix your timestep" pattern.
type GameLoop struct {
	TargetFPS float32
	MaxFrames int           // 0 = run forever
	Accum     time.Duration // internal
	FixedStep time.Duration
}

// NewGameLoop builds a loop at the given frame rate.
func NewGameLoop(targetFPS float32) *GameLoop {
	step := time.Duration(float32(time.Second) / targetFPS)
	return &GameLoop{TargetFPS: targetFPS, FixedStep: step}
}

// Run drives update(dtSeconds) and render() until maxFrames is reached or the
// stop function (if provided) returns true. update may be nil.
func (g *GameLoop) Run(update func(dt float32), render func() bool) {
	tick := time.NewTicker(g.FixedStep)
	defer tick.Stop()
	frames := 0
	for range tick.C {
		if update != nil {
			update(float32(g.FixedStep) / float32(time.Second))
		}
		if render != nil && !render() {
			return
		}
		frames++
		if g.MaxFrames > 0 && frames >= g.MaxFrames {
			return
		}
	}
}
