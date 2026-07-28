package llm

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Bedrock struct {
	Region      string
	MCPSessions []*mcpsdk.ClientSession
}

func NewBedrock(region string, mcpSessions ...*mcpsdk.ClientSession) *Bedrock {
	return &Bedrock{
		Region:      region,
		MCPSessions: mcpSessions,
	}
}
