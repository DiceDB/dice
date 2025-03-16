#!/bin/bash

LICENSE_NOTICE="// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information."

# Function to add license notice while preserving shebang and package declaration
add_license_notice() {
    local file="$1"
    local temp_file
    local permissions

    # Preserve file permissions
    permissions=$(stat -c "%a" "$file")

    # Create a temporary file
    temp_file=$(mktemp)

    # Read the first two lines to check for shebang and package declaration
    read -r first_line < "$file"
    read -r second_line < <(sed -n '2p' "$file")

    if [[ "$first_line" =~ ^#! ]]; then
        # If the first line is a shebang, preserve it
        {
            echo "$first_line"
            echo ""
            echo "$LICENSE_NOTICE"
            echo ""
            tail -n +2 "$file"
        } > "$temp_file"
    elif [[ "$first_line" =~ ^package ]]; then
        # If the first line is a package declaration, insert the license above it
        {
            echo "$LICENSE_NOTICE"
            echo ""
            cat "$file"
        } > "$temp_file"
    else
        # Just prepend the license notice
        {
            echo "$LICENSE_NOTICE"
            echo ""
            cat "$file"
        } > "$temp_file"
    fi

    # Replace original file with modified one
    mv "$temp_file" "$file"

    # Restore file permissions
    chmod "$permissions" "$file"

    echo "Added license notice to $file"
}

export -f add_license_notice
export LICENSE_NOTICE

# Find all Go files that don't already contain the license
mapfile -t files < <(find . -type f -name "*.go" ! -path "./vendor/*" ! -exec grep -qF "$LICENSE_NOTICE" {} \; -print)

# Process files
for file in "${files[@]}"; do
    add_license_notice "$file"
done

echo "Finished adding license notice to all Go files."
