package handshake

import (
	"fmt"
	"github.com/cloudstruct/go-ouroboros-network/protocol"
)

type Server struct {
	*protocol.Protocol
	config *Config
}

func NewServer(protoOptions protocol.ProtocolOptions, cfg *Config) *Server {
	s := &Server{
		config: cfg,
	}
	protoConfig := protocol.ProtocolConfig{
		Name:                PROTOCOL_NAME,
		ProtocolId:          PROTOCOL_ID,
		Muxer:               protoOptions.Muxer,
		ErrorChan:           protoOptions.ErrorChan,
		Mode:                protoOptions.Mode,
		Role:                protocol.ProtocolRoleServer,
		MessageHandlerFunc:  s.handleMessage,
		MessageFromCborFunc: NewMsgFromCbor,
		StateMap:            StateMap,
		InitialState:        STATE_PROPOSE,
	}
	s.Protocol = protocol.New(protoConfig)
	return s
}

func (s *Server) handleMessage(msg protocol.Message, isResponse bool) error {
	var err error
	switch msg.Type() {
	case MESSAGE_TYPE_PROPOSE_VERSIONS:
		err = s.handleProposeVersions(msg)
	default:
		err = fmt.Errorf("%s: received unexpected message type %d", PROTOCOL_NAME, msg.Type())
	}
	return err
}

func (s *Server) handleProposeVersions(msgGeneric protocol.Message) error {
	if s.config.FinishedFunc == nil {
		return fmt.Errorf("received handshake ProposeVersions message but no callback function is defined")
	}
	msg := msgGeneric.(*MsgProposeVersions)
	var highestVersion uint16
	var fullDuplex bool
	var versionData []interface{}
	for proposedVersion := range msg.VersionMap {
		if proposedVersion > highestVersion {
			for _, allowedVersion := range s.config.ProtocolVersions {
				if allowedVersion == proposedVersion {
					highestVersion = proposedVersion
					versionData = msg.VersionMap[proposedVersion].([]interface{})
					//nolint:gosimple
					if versionData[1].(bool) == DIFFUSION_MODE_INITIATOR_AND_RESPONDER {
						fullDuplex = true
					} else {
						fullDuplex = false
					}
					break
				}
			}
		}
	}
	if highestVersion > 0 {
		resp := NewMsgAcceptVersion(highestVersion, versionData)
		if err := s.SendMessage(resp); err != nil {
			return err
		}
		return s.config.FinishedFunc(highestVersion, fullDuplex)
	} else {
		// TODO: handle failures
		// https://github.com/cloudstruct/go-ouroboros-network/issues/32
		return fmt.Errorf("handshake failed, but we don't yet support this")
	}
}
