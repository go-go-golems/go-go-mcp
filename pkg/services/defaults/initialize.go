package defaults

import (
	"context"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

type DefaultInitializeService struct {
	serverName    string
	serverVersion string
}

func NewInitializeService(serverName string, serverVersion string) *DefaultInitializeService {
	return &DefaultInitializeService{
		serverName:    serverName,
		serverVersion: serverVersion,
	}
}

func (s *DefaultInitializeService) Initialize(ctx context.Context, params protocol.InitializeParams) (protocol.InitializeResult, error) {
	// Validate protocol version
	supportedVersions := []string{"2024-11-05"}
	isSupported := false
	for _, version := range supportedVersions {
		if params.ProtocolVersion == version {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return protocol.InitializeResult{}, fmt.Errorf("unsupported protocol version %s, supported versions: %v", params.ProtocolVersion, supportedVersions)
	}

	// Return server capabilities
	return protocol.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		Capabilities: protocol.ServerCapabilities{
			Logging: &protocol.LoggingCapability{},
			Prompts: &protocol.PromptsCapability{
				ListChanged: true,
			},
			Resources: &protocol.ResourcesCapability{
				Subscribe:   true,
				ListChanged: true,
			},
			Tools: &protocol.ToolsCapability{
				ListChanged: true,
			},
		},
		ServerInfo: protocol.ServerInfo{
			Name:    s.serverName,
			Version: s.serverVersion,
		},
	}, nil
}
