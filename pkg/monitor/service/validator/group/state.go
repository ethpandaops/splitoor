package group

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type State struct {
	log        logrus.FieldLogger
	Validators map[string]*Validators

	mu sync.Mutex
}

type Validators struct {
	Sources map[string]*Validator
}

type Validator struct {
	Balance                   uint64
	Status                    MetricsStatus
	WithdrawalCredentialsCode int64
}

func NewState(log logrus.FieldLogger) *State {
	return &State{
		log:        log,
		Validators: make(map[string]*Validators),
	}
}

func (s *State) UpdateValidator(source, pubkey string, balance uint64, status MetricsStatus, withdrawalCredentialsCode int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Validators[pubkey]; !exists {
		s.Validators[pubkey] = &Validators{
			Sources: make(map[string]*Validator),
		}
	}

	if validator, exists := s.Validators[pubkey].Sources[source]; exists {
		validator.Balance = balance
		validator.Status = status
		validator.WithdrawalCredentialsCode = withdrawalCredentialsCode
	} else {
		s.Validators[pubkey].Sources[source] = &Validator{
			Balance:                   balance,
			Status:                    status,
			WithdrawalCredentialsCode: withdrawalCredentialsCode,
		}
	}
}

func (s *State) Merge(other *State) (changedPubkeys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	other.mu.Lock()
	defer other.mu.Unlock()

	for pubkey, validators := range other.Validators {
		if _, exists := s.Validators[pubkey]; !exists {
			s.Validators[pubkey] = &Validators{
				Sources: make(map[string]*Validator),
			}
		}

		for source, validator := range validators.Sources {
			if currentValidator, exists := s.Validators[pubkey].Sources[source]; !exists {
				s.Validators[pubkey].Sources[source] = validator

				changedPubkeys = append(changedPubkeys, pubkey)
			} else {
				if currentValidator.Balance != validator.Balance || currentValidator.Status != validator.Status || currentValidator.WithdrawalCredentialsCode != validator.WithdrawalCredentialsCode {
					changedPubkeys = append(changedPubkeys, pubkey)
				}

				currentValidator.Balance = validator.Balance
				currentValidator.Status = validator.Status
				currentValidator.WithdrawalCredentialsCode = validator.WithdrawalCredentialsCode
			}
		}
	}

	return changedPubkeys
}
