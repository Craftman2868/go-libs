package render

import "github.com/Craftman2868/go-libs/event"

type RenderEndEvent struct {
	Elements []Element
}

func (RenderEndEvent) Name() string {
	return "renderEnd"
}

type Renderer struct {
	elements []Element
	handler  event.Eventable
}

func (rndr *Renderer) Init(handler event.Eventable) {
	rndr.handler = handler
}

func (rndr *Renderer) Render() {
	var rendered []Element

	for _, elem := range rndr.elements {
		if !elem.NeedRender() {
			continue
		}
		if len(rendered) == 0 {
			rndr.handler.HandleEvent(event.BaseEvent("renderBegin"))
		}
		elem.Render()
		rendered = append(rendered, elem)
	}

	if len(rendered) > 0 {
		rndr.handler.HandleEvent(RenderEndEvent{rendered})
	}
}

// The elements are rendered in the same order they are added
func (rndr *Renderer) AddElement(elem Element) {
	rndr.elements = append(rndr.elements, elem)
}
