// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package istanbul

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/params"
	"github.com/naoina/toml"
)

type ProposerPolicyId uint64

const (
	RoundRobin ProposerPolicyId = iota
	Sticky
)

// ProposerPolicy represents the Validator Proposer Policy
type ProposerPolicy struct {
	Id         ProposerPolicyId    // Could be RoundRobin or Sticky
	By         ValidatorSortByFunc // func that defines how the ValidatorSet should be sorted
	registry   []ValidatorSet      // Holds the ValidatorSet for a given block height
	registryMU *sync.Mutex         // Mutex to lock access to changes to Registry
}

// NewRoundRobinProposerPolicy returns a RoundRobin ProposerPolicy with ValidatorSortByString as default sort function
func NewRoundRobinProposerPolicy() *ProposerPolicy {
	return NewProposerPolicy(RoundRobin)
}

// NewStickyProposerPolicy return a Sticky ProposerPolicy with ValidatorSortByString as default sort function
func NewStickyProposerPolicy() *ProposerPolicy {
	return NewProposerPolicy(Sticky)
}

func NewProposerPolicy(id ProposerPolicyId) *ProposerPolicy {
	return NewProposerPolicyByIdAndSortFunc(id, ValidatorSortByString())
}

func NewProposerPolicyByIdAndSortFunc(id ProposerPolicyId, by ValidatorSortByFunc) *ProposerPolicy {
	return &ProposerPolicy{Id: id, By: by, registryMU: new(sync.Mutex)}
}

type proposerPolicyToml struct {
	Id ProposerPolicyId
}

func (p *ProposerPolicy) MarshalTOML() (interface{}, error) {
	if p == nil {
		return nil, nil
	}
	pp := &proposerPolicyToml{Id: p.Id}
	data, err := toml.Marshal(pp)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (p *ProposerPolicy) UnmarshalTOML(decode func(interface{}) error) error {
	var innerToml string
	err := decode(&innerToml)
	if err != nil {
		return err
	}
	var pp proposerPolicyToml
	err = toml.Unmarshal([]byte(innerToml), &pp)
	if err != nil {
		return err
	}
	p.Id = pp.Id
	p.By = ValidatorSortByString()
	return nil
}

// Use sets the ValidatorSortByFunc for the given ProposerPolicy and sorts the validatorSets according to it
func (p *ProposerPolicy) Use(v ValidatorSortByFunc) {
	p.By = v

	for _, validatorSet := range p.registry {
		validatorSet.SortValidators()
	}
}

// RegisterValidatorSet stores the given ValidatorSet in the policy registry
func (p *ProposerPolicy) RegisterValidatorSet(valSet ValidatorSet) {
	p.registryMU.Lock()
	defer p.registryMU.Unlock()

	if len(p.registry) == 0 {
		p.registry = []ValidatorSet{valSet}
	} else {
		p.registry = append(p.registry, valSet)
	}
}

// ClearRegistry removes any ValidatorSet from the ProposerPolicy registry
func (p *ProposerPolicy) ClearRegistry() {
	p.registryMU.Lock()
	defer p.registryMU.Unlock()

	p.registry = nil
}

type Config struct {
	RequestTimeout           uint64                `toml:",omitempty"` // The timeout for each Istanbul round in milliseconds.
	BlockPeriod              uint64                `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second
	EmptyBlockPeriod         uint64                `toml:",omitempty"` // Default minimum difference between a block and empty block's timestamps in second
	ProposerPolicy           *ProposerPolicy       `toml:",omitempty"` // The policy for proposer selection
	Epoch                    uint64                `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
	Ceil2Nby3Block           *big.Int              `toml:",omitempty"` // Number of confirmations required to move from one state to next [2F + 1 to Ceil(2N/3)]
	AllowedFutureBlockTime   uint64                `toml:",omitempty"` // Max time (in seconds) from current time allowed for blocks, before they're considered future blocks
	BeneficiaryMode          *string               `toml:",omitempty"` // Mode for setting the beneficiary, either: list, besu, validators (beneficiary list is the list of validators)
	BlockReward              *math.HexOrDecimal256 `toml:",omitempty"` // Reward
	MiningBeneficiary        *common.Address       `toml:",omitempty"` // Wallet address that benefits at every new block (besu mode)
	Validators               []common.Address      `toml:",omitempty"`
	ValidatorSelectionMode   *string               `toml:",omitempty"`
	Client                   bind.ContractCaller   `toml:",omitempty"`
	MaxRequestTimeoutSeconds uint64                `toml:",omitempty"`
	Transitions              []params.Transition
}

var DefaultConfig = &Config{
	RequestTimeout:         10000,
	BlockPeriod:            5,
	EmptyBlockPeriod:       0,
	ProposerPolicy:         NewRoundRobinProposerPolicy(),
	Epoch:                  30000,
	Ceil2Nby3Block:         big.NewInt(0),
	AllowedFutureBlockTime: 0,
}

func (c Config) GetConfig(blockNumber *big.Int) Config {
	newConfig := c

	c.getTransitionValue(blockNumber, func(transition params.Transition) {
		if transition.RequestTimeoutSeconds != 0 {
			// RequestTimeout is on milliseconds
			newConfig.RequestTimeout = transition.RequestTimeoutSeconds * 1000
		}
		if transition.EpochLength != 0 {
			newConfig.Epoch = transition.EpochLength
		}
		if transition.BlockPeriodSeconds != 0 {
			newConfig.BlockPeriod = transition.BlockPeriodSeconds
		}
		if transition.EmptyBlockPeriodSeconds != nil {
			newConfig.EmptyBlockPeriod = *transition.EmptyBlockPeriodSeconds
		}
		if transition.BeneficiaryMode != nil {
			newConfig.BeneficiaryMode = transition.BeneficiaryMode
		}
		if transition.BlockReward != nil {
			newConfig.BlockReward = transition.BlockReward
		}
		if transition.MiningBeneficiary != nil {
			newConfig.MiningBeneficiary = transition.MiningBeneficiary
		}
		if transition.ValidatorSelectionMode != "" {
			newConfig.ValidatorSelectionMode = &transition.ValidatorSelectionMode
		}
		if len(transition.Validators) > 0 {
			newConfig.Validators = transition.Validators
		}
		if transition.MaxRequestTimeoutSeconds != nil {
			newConfig.MaxRequestTimeoutSeconds = *transition.MaxRequestTimeoutSeconds
		}
	})

	return newConfig
}

func (c Config) GetValidatorSelectionMode(blockNumber *big.Int) string {
	mode := params.BlockHeaderMode
	if c.ValidatorSelectionMode != nil {
		mode = *c.ValidatorSelectionMode
	}
	c.getTransitionValue(blockNumber, func(transition params.Transition) {
		if transition.ValidatorSelectionMode != "" {
			mode = transition.ValidatorSelectionMode
		}
	})
	return mode
}

func (c Config) GetValidatorsAt(blockNumber *big.Int) []common.Address {
	if blockNumber.Cmp(big.NewInt(0)) == 0 && len(c.Validators) > 0 {
		return c.Validators
	}

	if blockNumber != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(blockNumber) == 0; i++ {
			return c.Transitions[i].Validators
		}
	}

	//Note! empty means we will get the valset from previous block header which contains votes, validators etc
	return []common.Address{}
}

func (c Config) Get2FPlus1Enabled(blockNumber *big.Int) bool {
	twoFPlusOneEnabled := false
	c.getTransitionValue(blockNumber, func(transition params.Transition) {
		if transition.TwoFPlusOneEnabled != nil {
			twoFPlusOneEnabled = *transition.TwoFPlusOneEnabled
		}
	})
	return twoFPlusOneEnabled
}

func (c *Config) getTransitionValue(num *big.Int, callback func(transition params.Transition)) {
	if c != nil && num != nil && c.Transitions != nil {
		for i := 0; i < len(c.Transitions) && c.Transitions[i].Block.Cmp(num) <= 0; i++ {
			callback(c.Transitions[i])
		}
	}
}