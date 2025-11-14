#!/bin/bash

# Test API call for the /explain endpoint
# Make sure the server is running on port 8080 first!

curl -X POST http://localhost:8080/explain \
  -H "Content-Type: application/json" \
  -d '{
    "text_content": "resource \"aws_instance\" \"web\" {\n  ami           = \"ami-0c55b159cbfafe1f0\"\n  instance_type = \"t2.micro\"\n  \n  tags = {\n    Name = \"WebServer\"\n  }\n}\n\nvariable \"region\" {\n  description = \"AWS region\"\n  type        = string\n  default     = \"us-east-1\"\n}\n\noutput \"instance_id\" {\n  value = aws_instance.web.id\n}"
  }'

echo ""

