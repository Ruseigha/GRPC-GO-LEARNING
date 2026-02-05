#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Generating Protocol Buffer code...${NC}"

# Generate Go code
protoc \
  --go_out=. \
  --go-grpc_out=. \
  proto/user/v1/user.proto

echo -e "${GREEN}✓ Protocol Buffer code generated${NC}"