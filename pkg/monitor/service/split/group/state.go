package group

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type State struct {
	log     logrus.FieldLogger
	Sources map[string]*Split

	mu sync.Mutex
}

type Split struct {
	Hash       string
	Controller string
}

func NewState(log logrus.FieldLogger) *State {
	return &State{
		log:     log,
		Sources: make(map[string]*Split),
	}
}

func (s *State) UpdateSplit(source, hash, controller string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if split, exists := s.Sources[source]; exists {
		split.Hash = hash
		split.Controller = controller
	} else {
		s.Sources[source] = &Split{
			Hash:       hash,
			Controller: controller,
		}
	}
}

func (s *State) Merge(other *State) (changedSources []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	other.mu.Lock()
	defer other.mu.Unlock()

	for source, split := range other.Sources {
		if _, exists := s.Sources[source]; !exists {
			s.Sources[source] = &Split{
				Hash:       split.Hash,
				Controller: split.Controller,
			}
		}

		if currentSplit, exists := s.Sources[source]; !exists {
			s.Sources[source] = split

			changedSources = append(changedSources, source)
		} else if currentSplit.Hash != split.Hash || currentSplit.Controller != split.Controller {
			changedSources = append(changedSources, source)
		}
	}

	return changedSources
}
