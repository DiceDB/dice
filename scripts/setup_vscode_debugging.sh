#!/bin/bash

# Define the directory and file path
VSCODE_DIR=".vscode"
LAUNCH_JSON="$VSCODE_DIR/launch.json"

# Create the .vscode directory if it doesn't exist
if [ ! -d "$VSCODE_DIR" ]; then
  mkdir "$VSCODE_DIR"
  echo "Created directory: $VSCODE_DIR"
fi

# Define the JSON content
read -r -d '' JSON_CONTENT << EOM
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "\${workspaceFolder}",
      "env": {},
      "args": []
    }
  ]
}
EOM

# Write the JSON content to the launch.json file
echo "$JSON_CONTENT" > "$LAUNCH_JSON"

echo "Created $LAUNCH_JSON with Go debug configuration."
