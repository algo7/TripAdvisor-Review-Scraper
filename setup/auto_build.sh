#!/bin/bash

# Args supplied by the user
package=$1

# Input validation
if [[ -z "$package" ]]; then
  echo "usage: $0 <package-name>"
  exit 1
fi

# Extract the pkg name
package_split=(${package//\// })
package_name=${package_split[-1]}

# Supported platforms
platforms=("windows/amd64" "linux/amd64" "darwin/amd64")

for platform in "${platforms[@]}"
do

# Assign the correct OS and ARCH
	platform_split=(${platform//\// })
	GOOS=${platform_split[0]}
	GOARCH=${platform_split[1]}
	# Final pkg name
	output_name=$package_name'-'$GOOS'-'$GOARCH

	# Assign the appropriate extension based on the platform
	if [ $GOOS = "windows" ]; then
		output_name+='.exe'
	fi
	if [ $GOOS = "linux" ]; then
		output_name+='.bin'
	fi
	if [ $GOOS = "darwin" ]; then
		output_name+='.bin'
	fi	

# Set the env vars and build the binary
	env GOOS=$GOOS GOARCH=$GOARCH go build -o ../builds/$output_name $package

	# Exit if the build failed
	if [ $? -ne 0 ]; then
   		echo 'An error has occurred! Aborting the script execution...'
		exit 1
	fi
done