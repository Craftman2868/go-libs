package render

type Element interface {
	NeedRender() bool
	Render()
}

type UpdatableElement interface {
	Updated() bool
}

// Doesn't implement Element, to be embedded in another struct with a Render() method
type Updatable struct {
	updated bool
}

// Should not be called in any element's Render() method
func (elem *Updatable) Updated() {
	elem.updated = true
}

// To be called at the end of the Render() method
func (elem *Updatable) RenderEnd() {
	elem.updated = false
}

func (elem *Updatable) NeedRender() bool {
	return elem.updated
}

type BaseElement struct {
	Updatable

	renderFunc func()
}

func NewBaseElement(renderFunc func()) *BaseElement {
	return &BaseElement{
		Updatable{true},
		renderFunc,
	}
}

func (elem *BaseElement) Render() {
	elem.renderFunc()
	elem.RenderEnd()
}

type AlwaysRender struct{}

func (*AlwaysRender) NeedRender() bool {
	return true
}

type Visibility struct {
	Visible bool
}

func (elem *Visibility) NeedRender() bool {
	return elem.Visible
}

func (elem *Visibility) Show() {
	elem.Visible = true
}

func (elem *Visibility) Hide() {
	elem.Visible = false
}
