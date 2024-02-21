package manager

import "github.com/Erikqwerty/KronosKeeper/internal/pkg/config"

type Kkmanager struct {
}

func New(conf config.Config) *Kkmanager {
	return &Kkmanager{}
}
