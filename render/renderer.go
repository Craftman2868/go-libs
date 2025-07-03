package render

import "github.com/Craftman2868/go-libs/event"

type Renderer struct {
	elements []Element
	handler  event.Eventable
}

func (rndr *Renderer) Init(handler event.Eventable) {
	rndr.handler = handler
}

func (rndr *Renderer) Render() {
	renderHasBegun := false

	for _, elem := range rndr.elements {
		if !elem.NeedRender() {
			continue
		}
		if !renderHasBegun {
			rndr.handler.HandleEvent(event.BaseEvent("renderBegin"))
			renderHasBegun = true
		}
		elem.Render()
	}

	if renderHasBegun {
		rndr.handler.HandleEvent(event.BaseEvent("renderEnd"))
	}
}

// The elements are rendered in the same order they are added
func (rndr *Renderer) AddElement(elem Element) {
	rndr.elements = append(rndr.elements, elem)
}
