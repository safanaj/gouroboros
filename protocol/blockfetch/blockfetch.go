// Copyright 2023 Blink Labs, LLC.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blockfetch

import (
	"time"

	"github.com/blinklabs-io/gouroboros/protocol"

	"github.com/blinklabs-io/gouroboros/ledger"
)

const (
	PROTOCOL_NAME        = "block-fetch"
	PROTOCOL_ID   uint16 = 3
)

var (
	STATE_IDLE      = protocol.NewState(1, "Idle")
	STATE_BUSY      = protocol.NewState(2, "Busy")
	STATE_STREAMING = protocol.NewState(3, "Streaming")
	STATE_DONE      = protocol.NewState(4, "Done")
)

var StateMap = protocol.StateMap{
	STATE_IDLE: protocol.StateMapEntry{
		Agency: protocol.AgencyClient,
		Transitions: []protocol.StateTransition{
			{
				MsgType:  MESSAGE_TYPE_REQUEST_RANGE,
				NewState: STATE_BUSY,
			},
			{
				MsgType:  MESSAGE_TYPE_CLIENT_DONE,
				NewState: STATE_DONE,
			},
		},
	},
	STATE_BUSY: protocol.StateMapEntry{
		Agency: protocol.AgencyServer,
		Transitions: []protocol.StateTransition{
			{
				MsgType:  MESSAGE_TYPE_START_BATCH,
				NewState: STATE_STREAMING,
			},
			{
				MsgType:  MESSAGE_TYPE_NO_BLOCKS,
				NewState: STATE_IDLE,
			},
		},
	},
	STATE_STREAMING: protocol.StateMapEntry{
		Agency: protocol.AgencyServer,
		Transitions: []protocol.StateTransition{
			{
				MsgType:  MESSAGE_TYPE_BLOCK,
				NewState: STATE_STREAMING,
			},
			{
				MsgType:  MESSAGE_TYPE_BATCH_DONE,
				NewState: STATE_IDLE,
			},
		},
	},
	STATE_DONE: protocol.StateMapEntry{
		Agency: protocol.AgencyNone,
	},
}

type BlockFetch struct {
	Client *Client
	Server *Server
}

type Config struct {
	BlockFunc         BlockFunc
	BatchStartTimeout time.Duration
	BlockTimeout      time.Duration
}

// Callback function types
type BlockFunc func(ledger.Block) error

func New(protoOptions protocol.ProtocolOptions, cfg *Config) *BlockFetch {
	b := &BlockFetch{
		Client: NewClient(protoOptions, cfg),
		Server: NewServer(protoOptions, cfg),
	}
	return b
}

type BlockFetchOptionFunc func(*Config)

func NewConfig(options ...BlockFetchOptionFunc) Config {
	c := Config{
		BatchStartTimeout: 5 * time.Second,
		BlockTimeout:      60 * time.Second,
	}
	// Apply provided options functions
	for _, option := range options {
		option(&c)
	}
	return c
}

func WithBlockFunc(blockFunc BlockFunc) BlockFetchOptionFunc {
	return func(c *Config) {
		c.BlockFunc = blockFunc
	}
}

func WithBatchStartTimeout(timeout time.Duration) BlockFetchOptionFunc {
	return func(c *Config) {
		c.BatchStartTimeout = timeout
	}
}

func WithBlockTimeout(timeout time.Duration) BlockFetchOptionFunc {
	return func(c *Config) {
		c.BlockTimeout = timeout
	}
}
