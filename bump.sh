#!/bin/bash

# First check for -y flag
SKIP_CONFIRM=false
if [[ " $* " =~ " -y " ]]; then
    SKIP_CONFIRM=true
    # Remove -y from arguments to process version parameter
    set -- ${@/-y/}
fi

# Check if version type argument is provided
if [ $# -ne 1 ]; then
    echo "Usage: $0 [major|minor|patch|alpha]"
    exit 1
fi

VERSION_TYPE=$1

# Validate version type
if [ "$VERSION_TYPE" != "major" ] && [ "$VERSION_TYPE" != "minor" ] && [ "$VERSION_TYPE" != "patch" ] && [ "$VERSION_TYPE" != "alpha" ]; then
    echo "Error: Version type must be 'major', 'minor', 'patch', or 'alpha'"
    exit 1
fi

# Read current version
CURRENT_VERSION=$(jq -r '.version' ./info.json)
if [ -z "$CURRENT_VERSION" ]; then
    echo "Error: Unable to read version from info.json"
    exit 1
fi

# Split version into components
IFS='.' read -r -a VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR="${VERSION_PARTS[0]}"
MINOR="${VERSION_PARTS[1]}"
PATCH="${VERSION_PARTS[2]}"
ALPHA=0

# Check if current version has alpha component
if [[ "$CURRENT_VERSION" =~ -alpha\.[0-9]+$ ]]; then
    ALPHA=$(echo "$CURRENT_VERSION" | grep -o '[0-9]*$')
fi

# Increment version based on type
case "$VERSION_TYPE" in
    "major")
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ALPHA=0
        ;;
    "minor")
        MINOR=$((MINOR + 1))
        PATCH=0
        ALPHA=0
        ;;
    "patch")
        PATCH=$((PATCH + 1))
        ALPHA=0
        ;;
    "alpha")
        ALPHA=$((ALPHA + 1))
        ;;
esac

# Create new version string
if [ "$VERSION_TYPE" = "alpha" ]; then
    NEW_VERSION="$MAJOR.$MINOR.$PATCH-alpha.$ALPHA"
else
    NEW_VERSION="$MAJOR.$MINOR.$PATCH"
fi

# Update info.json with new version
tmp=$(mktemp)
jq --arg version "$NEW_VERSION" '.version = $version' ./info.json > "$tmp" && mv "$tmp" ./info.json

echo "Version bumped from $CURRENT_VERSION to $NEW_VERSION"

VERSION=$(jq -r '.version' ./info.json)

if [ -z "$VERSION" ]; then
    echo "Error: Unable to read version from info.json"
    exit 1
fi

VERSION="v$VERSION"

if [ "$SKIP_CONFIRM" != "true" ]; then
    echo -n "Are you sure you want to release version $VERSION? (y/N): "
    read confirmation

    if [ "$confirmation" != "y" ] && [ "$confirmation" != "Y" ]; then
        echo "Release cancelled."
        exit 0
    fi
fi

git add info.json
git commit -m "build: 🔖 bump version to $VERSION"
git push

git tag -a -f $VERSION -m "$VERSION"
git push origin -f $VERSION