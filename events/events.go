package eventbus

type Event struct {
	Name      string
	Listeners map[string][]chan interface{}
}

func (b *Event) AddListener(e string, ch chan interface{}) {
	if b.Listeners == nil {
		b.Listeners = make(map[string][]chan interface{})
	}
	if _, ok := b.Listeners[e]; ok {
		b.Listeners[e] = append(b.Listeners[e], ch)
	} else {
		b.Listeners[e] = []chan interface{}{ch}
	}
}

func (b *Event) RemoveListener(e string, ch chan interface{}) {
	if _, ok := b.Listeners[e]; ok {
		for i := range b.Listeners[e] {
			if b.Listeners[e][i] == ch {
				b.Listeners[e] = append(b.Listeners[e][:i], b.Listeners[e][i+1:]...)
				break
			}
		}
	}
}

func (b *Event) Emit(e string, response string) {
	if _, ok := b.Listeners[e]; ok {
		for _, handler := range b.Listeners[e] {
			go func(handler chan interface{}) {
				handler <- response
			}(handler)
		}
	}
}
