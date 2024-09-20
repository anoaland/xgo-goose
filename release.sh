VERSION=$(jq -r '.version' ./info.json)

if [ -z "$VERSION" ]; then
    echo "Error: Unable to read version from info.json"
    exit 1
fi

VERSION="v$VERSION"

echo -n "Are you sure you want to release version $VERSION? (y/N): "
read confirmation

if [ "$confirmation" != "y" ] && [ "$confirmation" != "Y" ]; then
    echo "Release cancelled."
    exit 0
fi

git tag -a -f $VERSION -m "$VERSION"

git push origin -f $VERSION
