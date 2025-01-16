#!/bin/bash

# Exit on error
set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format v1.0.0"
    exit 1
fi

# Ensure we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Error: Must be on main branch to create a release"
    exit 1
fi

# Ensure working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory is not clean. Commit or stash changes first."
    exit 1
fi

# Pull latest changes
git pull origin main

# Create and push tag
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION

# Wait for CI to complete
echo "Pushed tag $VERSION to origin. GitHub Actions will create the release automatically."
echo "You can monitor the progress at: https://github.com/$(gh repo view --json nameWithOwner -q .nameWithOwner)/actions" 