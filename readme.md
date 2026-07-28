## Commands


`cmd/check_emails/main.go` - gets emails from stored tokens. Iterates over them and uses Bedrock, with internal and external tools, to draft responses to emails that match criteria

`cmd/gmail_login/main.go` - runs an http server on port 8080 with endpoints:
 - `/login` - for human users to login to gmail
 - `/callback` - for Google's OAuth flow

 `cmd/check_gmail/main.go` - for testing only - a single gmail run against the hardcoded email address.

 ## Project

 - agentfiles - this is the local directory for AGENT.md files per email. Only used for local testing.
 - agentfilestorage - package for interface and implementations for retrieval of AGENT.md files.
 - cmd - directory for commands, see above.
 - gmail - package gmail management, i.e. read, draft, login, etc.
 - llm - package for LLMs
 - mcp - package for internal MCP server and tools. The tools are all email-related.
 - tokens - this is the local directory for email OAuth tokens per email. Only used for local testing.
 - tokenstorage - package for interface and implementations for retrieval of OAuth tokens.
 - user - package for user, essentially just an email & email type.

 ## TODO

 - finish Hexagon MCP tools - in Hexagon service
 - infrastructure, managed by Terraform
    - Lambdas+API Gateway -or- ECS and EventBridge
    - SSM Params for config values
    - S3 for per-user AGENT.md files
    - 2 tasks: gmail_login (http server), and check_gmail (cron task)
 - If not using Eventbridge, write command for cron-type task to run check_gmail-type call for all registered gmail accounts
 - Determine how to run 2 services: Procfile, multi-stage docker build, add all to main.go, etc.
 - swap hardcoded values for configurable values for, at least the following:
    - GOOGLE_CLIENT_ID
    - GOOGLE_CLIENT_SECRET
    - HEXAGON_URL
    - Port for http server, maybe
 - add customizable agentfilestorage and tokenstorage using AWS resources. Currently, these are hardcoded local file sources.
   - implement `s3.go` in agentfilestorage
   - implement `ssm.go` in tokenstorage

 - 2nd LLM for verification
 - Hex API security
 - Logging via 
 