package render

type Element interface {
	NeedRender() bool
	Render()
}

// Doesn't implement Element, to be embedded in another struct with a Render() method
type UpdatableElement struct {
	updated bool
}

// Should not be called in any element's Render() method
func (elem *UpdatableElement) Updated() {
	elem.updated = true
}

// To be called at the end of the Render() method
func (elem *UpdatableElement) RenderEnd() {
	elem.updated = false
}

func (elem *UpdatableElement) NeedRender() bool {
	return elem.updated
}

type BaseElement struct {
	UpdatableElement

	renderFunc func()
}

func NewBaseElement(renderFunc func()) *BaseElement {
	return &BaseElement{
		UpdatableElement{true},
		renderFunc,
	}
}

func (elem *BaseElement) Render() {
	elem.renderFunc()
	elem.RenderEnd()
}
