// Copyright 2016 The go-ethereum Authors
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

package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// IstanbulConfig is the consensus engine configs for Istanbul based sealing.
type IstanbulConfig struct {
	Epoch          uint64   `json:"epoch"`                    // Epoch length to reset votes and checkpoint
	ProposerPolicy uint64   `json:"policy"`                   // The policy for proposer selection
	Ceil2Nby3Block *big.Int `json:"ceil2Nby3Block,omitempty"` // Number of confirmations required to move from one state to next [2F + 1 to Ceil(2N/3)]
}

// String implements the stringer interface, returning the consensus engine details.
func (c *IstanbulConfig) String() string {
	return "istanbul"
}

type QBFTConfig struct {
	EpochLength              uint64                `json:"epochlength"`                       // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds       uint64                `json:"blockperiodseconds"`                // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	EmptyBlockPeriodSeconds  *uint64               `json:"emptyblockperiodseconds,omitempty"` // Minimum time between two consecutive QBFT a block and empty block’ timestamps in seconds
	RequestTimeoutSeconds    uint64                `json:"requesttimeoutseconds"`             // Minimum request timeout for each QBFT round in milliseconds
	ProposerPolicy           uint64                `json:"policy"`                            // The policy for proposer selection
	Ceil2Nby3Block           *big.Int              `json:"ceil2Nby3Block,omitempty"`          // Number of confirmations required to move from one state to next [2F + 1 to Ceil(2N/3)]
	BlockReward              *math.HexOrDecimal256 `json:"blockReward,omitempty"`             // Reward from start, works only on QBFT consensus protocol
	BeneficiaryMode          *string               `json:"beneficiaryMode,omitempty"`         // Mode for setting the beneficiary, either: list, besu, validators (beneficiary list is the list of validators)
	MiningBeneficiary        *common.Address       `json:"miningBeneficiary,omitempty"`       // Wallet address that benefits at every new block (besu mode)
	ValidatorSelectionMode   *string               `json:"validatorselectionmode,omitempty"`  // Select model for validators
	Validators               []common.Address      `json:"validators"`                        // Validators list
	MaxRequestTimeoutSeconds *uint64               `json:"maxRequestTimeoutSeconds"`          // The max round time
}

func (c QBFTConfig) String() string {
	return "qbft"
}

const (
	ContractMode    = "contract"
	BlockHeaderMode = "blockheader"
)

type Transition struct {
	Block                        *big.Int              `json:"block"`
	EpochLength                  uint64                `json:"epochlength,omitempty"`                  // Number of blocks that should pass before pending validator votes are reset
	BlockPeriodSeconds           uint64                `json:"blockperiodseconds,omitempty"`           // Minimum time between two consecutive QBFT blocks’ timestamps in seconds
	EmptyBlockPeriodSeconds      *uint64               `json:"emptyblockperiodseconds,omitempty"`      // Minimum time between two consecutive QBFT a block and empty block’ timestamps in seconds
	RequestTimeoutSeconds        uint64                `json:"requesttimeoutseconds,omitempty"`        // Minimum request timeout for each QBFT round in milliseconds
	ContractSizeLimit            uint64                `json:"contractsizelimit,omitempty"`            // Maximum smart contract code size
	Validators                   []common.Address      `json:"validators"`                             // List of validators
	ValidatorSelectionMode       string                `json:"validatorselectionmode,omitempty"`       // Validator selection mode to switch to
	EnhancedPermissioningEnabled *bool                 `json:"enhancedPermissioningEnabled,omitempty"` // aka QIP714Block
	PrivacyEnhancementsEnabled   *bool                 `json:"privacyEnhancementsEnabled,omitempty"`   // privacy enhancements (mandatory party, private state validation)
	PrivacyPrecompileEnabled     *bool                 `json:"privacyPrecompileEnabled,omitempty"`     // enable marker transactions support
	GasPriceEnabled              *bool                 `json:"gasPriceEnabled,omitempty"`              // enable gas price
	MinerGasLimit                uint64                `json:"miner.gaslimit,omitempty"`               // Gas Limit
	TwoFPlusOneEnabled           *bool                 `json:"2FPlus1Enabled,omitempty"`               // Ceil(2N/3) is the default you need to explicitly use 2F + 1
	TransactionSizeLimit         uint64                `json:"transactionSizeLimit,omitempty"`         // Modify TransactionSizeLimit
	BlockReward                  *math.HexOrDecimal256 `json:"blockReward,omitempty"`                  // validation rewards
	BeneficiaryMode              *string               `json:"beneficiaryMode,omitempty"`              // Mode for setting the beneficiary, either: list, besu, validators (beneficiary list is the list of validators)
	MiningBeneficiary            *common.Address       `json:"miningBeneficiary,omitempty"`            // Wallet address that benefits at every new block (besu mode)
	MaxRequestTimeoutSeconds     *uint64               `json:"maxRequestTimeoutSeconds,omitempty"`     // The max a timeout should be for a round change
}
