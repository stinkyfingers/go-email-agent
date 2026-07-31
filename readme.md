## Commands

  
  

`cmd/check_emails/main.go` - gets emails from stored tokens. Iterates over them and uses Bedrock, with internal and external tools, to draft responses to emails that match criteria

  

`cmd/gmail_login/main.go` - runs an http server on port 8080 with endpoints:

-  `/login` - for human users to login to gmail

-  `/callback` - for Google's OAuth flow

  
  

## Run Locally

  

- assure you have an .env file with SLACK_WEBHOOK, GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET

- you'll need a slack app with a webhook

- you'll need a Google Cloud account - John has a personal one established

- Assure you have an agent file at agentfiles/<your_email>/AGENT.md, detailing to the LLM how to handle emails.

  

`go run cmd/gmail_login/main.go`

- visit localhost:8080/login and authorize a gmail address

- note: the address must be added to "allowed emails," which is currently in my personal Google Cloud account

`go run cmd/check_emails/main.go`

  
  

## Deploy

  

#### Bootstrap

  

`cd terraform/bootstrap`

`terraform init`

`terraform apply`

- Do once only, ever. Creates the S3 bucket for terraform state.

*NOTE*  `state-bucket-name` go-email-agent-terraform-state needs to be globally unique. If this was created in a test AWS account, it may need to be cleaned up before

deploying to a live AWS account.

  

#### Set Vars

  

Set values in tfvars.tf (copy template from tfvars.tf.example)

  

#### Deploy to AWS

  

`cd terraform` # back in the main dir

`terraform init` # picks up the S3 backend bootstrap just created, above

`terraform plan`

`terraform apply`

  

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

  

- in Hexagon codebase

  - [x] finish Hexagon MCP tools - in Hexagon service, branch HEX-310-mcp-server - copy remaining DB tools from Connor's original project.

- in this codebase

  - [x] infrastructure, managed by Terraform

  - [x] Lambdas+API Gateway -or- ECS and EventBridge (or other cron-style task)

  - [x] SSM Params for config values

    - [x] S3 for per-user AGENT.md files

  - [x] 2 tasks: gmail_login (http server), and check_gmail (cron task)

  - [x] Determine best way to run 2 services (http server and cronjob): Procfile, multi-stage docker build, etc.

  - [x] replace hardcoded envvar values with configurable values for, at least the following:

    - [x] GOOGLE_CLIENT_ID

    - [x] GOOGLE_CLIENT_SECRET

    - [x] HEXAGON_URL

  - [x] add customizable agentfilestorage and tokenstorage using AWS resources. Currently, these are hardcoded local file sources in the various cmd/.../main.go packages.

    - [x] implement `s3.go` in agentfilestorage (or somethingsimilar)

    - [x] implement `ssm.go` in tokenstorage (or something similar)

  - [ ] add 2nd LLM step for verification

  - [ ] Sentry integration

  - [ ]  <b>Confirm Terraform works</B> and deploy

- General items

  - [ ] Hex API authentication for /mcp endpoint used by this rep

  - [ ] Set up "company" Google Cloud account and give team access.

    - APIs & Services

    - enable Gmail API

    - OAuth Consent Screen -> Clients

  - add a Web Client (save ClientID and Secret)

    - add redirect URI (localhost:8080/callback and whatever host the go-email-agent winds up being)

    - OAuth Consent Screen -> Audience

    - add Test Users

  - [ ] Move project to repo in ticketco-web github account