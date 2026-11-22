package x3d

import "math"

// EaseFunc maps a normalized time t in [0,1] to an eased value.
type EaseFunc func(t float32) float32

// Built-in easing functions.
var (
	EaseLinear    EaseFunc = func(t float32) float32 { return t }
	EaseInQuad    EaseFunc = func(t float32) float32 { return t * t }
	EaseOutQuad   EaseFunc = func(t float32) float32 { return t * (2 - t) }
	EaseInOutQuad EaseFunc = func(t float32) float32 {
		if t < 0.5 {
			return 2 * t * t
		}
		return 1 - float32(math.Pow(float64(-2*t+2), 2))/2
	}
	EaseInCubic    EaseFunc = func(t float32) float32 { return t * t * t }
	EaseOutCubic   EaseFunc = func(t float32) float32 { f := t - 1; return f*f*f + 1 }
	EaseInOutCubic EaseFunc = func(t float32) float32 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		f := -2*t + 2
		return 1 - float32(math.Pow(float64(f), 3))/2
	}
	EaseInSine    EaseFunc = func(t float32) float32 { return 1 - float32(math.Cos(float64(t)*math.Pi/2)) }
	EaseOutSine   EaseFunc = func(t float32) float32 { return float32(math.Sin(float64(t) * math.Pi / 2)) }
	EaseInOutSine EaseFunc = func(t float32) float32 {
		return -(float32(math.Cos(float64(t)*math.Pi)) - 1) / 2
	}
	EaseOutElastic EaseFunc = func(t float32) float32 {
		if t == 0 || t == 1 {
			return t
		}
		c4 := 2 * float32(math.Pi) / 3
		return float32(math.Pow(2, float64(-10*float32(t))))*float32(math.Sin(float64((t*10-0.75)*float32(c4)))) + 1
	}
	EaseOutBounce EaseFunc = func(t float32) float32 {
		n1 := float32(7.5625)
		d1 := float32(2.75)
		switch {
		case t < 1/d1:
			return n1 * t * t
		case t < 2/d1:
			t -= 1.5 / d1
			return n1*t*t + 0.75
		case t < 2.5/d1:
			t -= 2.25 / d1
			return n1*t*t + 0.9375
		default:
			t -= 2.625 / d1
			return n1*t*t + 0.984375
		}
	}
	EaseOutBack EaseFunc = func(t float32) float32 {
		c1 := float32(1.70158)
		c3 := c1 + 1
		return 1 + c3*float32(math.Pow(float64(t-1), 3)) + c1*float32(math.Pow(float64(t-1), 2))
	}
)

// Tween animates a single float from -> to over a duration.
type Tween struct {
	From, To, Duration float32
	Ease               EaseFunc
}

// NewTween builds a tween with the given easing.
func NewTween(from, to, duration float32, ease EaseFunc) Tween {
	if ease == nil {
		ease = EaseLinear
	}
	return Tween{From: from, To: to, Duration: duration, Ease: ease}
}

// Value returns the tweened value at local time t (clamped to the duration).
func (tw Tween) Value(t float32) float32 {
	if tw.Duration <= 0 {
		return tw.To
	}
	u := Clamp01(t / tw.Duration)
	return Lerp(tw.From, tw.To, tw.Ease(u))
}

// Keyframe is a single (time, value) sample for a Vec3 track.
type Keyframe struct {
	Time  float32
	Value Vec3
	Ease  EaseFunc
}

// Vec3Track interpolates a sequence of keyframes.
type Vec3Track struct{ Keys []Keyframe }

// Sample evaluates the track at absolute time t.
func (tr Vec3Track) Sample(t float32) Vec3 {
	if len(tr.Keys) == 0 {
		return Zero3
	}
	if t <= tr.Keys[0].Time {
		return tr.Keys[0].Value
	}
	last := tr.Keys[len(tr.Keys)-1]
	if t >= last.Time {
		return last.Value
	}
	for i := 0; i+1 < len(tr.Keys); i++ {
		a, b := tr.Keys[i], tr.Keys[i+1]
		if t >= a.Time && t <= b.Time {
			u := InvLerp(a.Time, b.Time, t)
			if a.Ease != nil {
				u = a.Ease(u)
			}
			return a.Value.Lerp(b.Value, u)
		}
	}
	return last.Value
}

// Clip is a named set of tracks played together.
type Clip struct {
	Name      string
	Duration  float32
	Positions map[string]Vec3Track
}

// NewClip returns an empty clip.
func NewClip(name string) *Clip {
	return &Clip{Name: name, Positions: map[string]Vec3Track{}}
}

// AddPosition registers a position track for a node name.
func (c *Clip) AddPosition(node string, keys []Keyframe) {
	c.Positions[node] = Vec3Track{Keys: keys}
	if len(keys) > 0 && keys[len(keys)-1].Time > c.Duration {
		c.Duration = keys[len(keys)-1].Time
	}
}

// Animator plays a clip, optionally looping, and applies results to a scene.
type Animator struct {
	Clip   *Clip
	Time   float32
	Speed  float32
	Loop   bool
	Scene  *Scene
	byName map[string]*Node
}

// NewAnimator binds a clip to a scene, indexing nodes by name.
func NewAnimator(clip *Clip, scene *Scene) *Animator {
	an := &Animator{Clip: clip, Speed: 1, Scene: scene, byName: map[string]*Node{}}
	var walk func(n *Node)
	walk = func(n *Node) {
		an.byName[n.Name] = n
		for _, c := range n.children {
			walk(c)
		}
	}
	walk(scene.Root)
	return an
}

// Advance steps the animation clock by dt seconds and applies the pose.
func (an *Animator) Advance(dt float32) {
	an.Time += dt * an.Speed
	if an.Clip == nil {
		return
	}
	t := an.Time
	if an.Clip.Duration > 0 {
		if an.Loop {
			t = float32(math.Mod(float64(t), float64(an.Clip.Duration)))
		} else if t > an.Clip.Duration {
			t = an.Clip.Duration
		}
	}
	for name, tr := range an.Clip.Positions {
		if n, ok := an.byName[name]; ok {
			n.Position = tr.Sample(t)
		}
	}
}
