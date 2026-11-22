package x3d

// Node is a transform in the scene graph. It carries a local transform and any
// number of children. Renderable nodes attach a Mesh + Material via Entity.
type Node struct {
	Name     string
	Position Vec3
	Rotation Quaternion
	Scale    Vec3
	Visible  bool

	children []*Node
	entity   *Entity
	parent   *Node
}

// NewNode creates a visible node with an identity transform.
func NewNode(name string) *Node {
	return &Node{Name: name, Scale: One3, Rotation: QuatIdentity(), Visible: true}
}

// AddChild attaches a child node and returns it for chaining.
func (n *Node) AddChild(c *Node) *Node {
	c.parent = n
	n.children = append(n.children, c)
	return c
}

// Attach binds a mesh/material to this node.
func (n *Node) Attach(e *Entity) { n.entity = e }

// Children returns the node's children.
func (n *Node) Children() []*Node { return n.children }

// Local returns the node's local transform matrix.
func (n *Node) Local() Mat4 {
	return Translate(n.Position).Mul(n.Rotation.Mat4()).Mul(Scale(n.Scale))
}

// World composes the transform up the parent chain.
func (n *Node) World() Mat4 {
	if n.parent == nil {
		return n.Local()
	}
	return n.parent.World().Mul(n.Local())
}

// Entity binds geometry to a material.
type Entity struct {
	Mesh     *Mesh
	Material Material
}

// NewEntity builds an entity from a mesh and material.
func NewEntity(m *Mesh, mat Material) *Entity { return &Entity{Mesh: m, Material: mat} }

// Scene is the root of a scene graph plus its lights and background.
type Scene struct {
	Root       *Node
	Lights     []Light
	Background Color
}

// NewScene returns an empty scene with a root node and default lights.
func NewScene() *Scene {
	return &Scene{Root: NewNode("root"), Lights: DefaultLights(), Background: Hex("#0b0e14")}
}

// Add attaches a node to the scene root and returns it.
func (s *Scene) Add(n *Node) *Node { return s.Root.AddChild(n) }

// AddMesh is a convenience that wraps a mesh in an entity on a new node.
func (s *Scene) AddMesh(name string, mesh *Mesh, mat Material) *Node {
	n := NewNode(name)
	n.Attach(NewEntity(mesh, mat))
	s.Add(n)
	return n
}

// visibleEntities walks the graph collecting (worldMatrix, entity) pairs.
func (s *Scene) visibleEntities() []drawItem {
	var out []drawItem
	var walk func(n *Node)
	walk = func(n *Node) {
		if !n.Visible {
			return
		}
		if n.entity != nil {
			out = append(out, drawItem{world: n.World(), ent: n.entity})
		}
		for _, c := range n.children {
			walk(c)
		}
	}
	walk(s.Root)
	return out
}
