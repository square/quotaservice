package config

type Notifier struct {
	Watcher chan struct{}
}

func (n *Notifier) Notify() {
	select {
	case n.Watcher <- struct{}{}:
		// Done.
	default:
		// Already a message on the channel.
	}
}

func NewNotifier() *Notifier {
	return &Notifier{make(chan struct{}, 1)}
}
