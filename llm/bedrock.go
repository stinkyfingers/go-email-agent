package llm

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Bedrock struct {
	Region string
}

func NewBedrock(region string, mcpSession *mcpsdk.ClientSession) *Bedrock {
	return &Bedrock{
		Region: region,
	}
}
