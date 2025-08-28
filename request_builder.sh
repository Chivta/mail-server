#!/bin/bash

# Read JSON body
body=$(cat ./request_body.json)

# Read HTTP headers
headers=$(cat ./request_header.txt)

# Calculate Content-Length
content_length=${#body}

# Add/replace Content-Length header
# Remove existing Content-Length if present
headers=$(echo "$headers" | grep -vi '^Content-Length:')

# Combine headers and body
echo "$headers"
echo "Content-Length: $content_length"
echo
echo "$body"