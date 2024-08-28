#!/bin/bash

# Directory containing architecture-specific subdirectories (sibling of stash)
LIB_SOURCE_DIR="../build"
# Directory where the libraries should be copied to for packaging
LIB_DEST_DIR="stash/_libs"

# Check if the source directory exists
if [ ! -d "$LIB_SOURCE_DIR" ]; then
  echo "Error: Directory $LIB_SOURCE_DIR does not exist."
  exit 1
fi

rm -rf dist/

# Iterate over each subdirectory in the source directory
for arch in $(ls -d ${LIB_SOURCE_DIR}/*/ | xargs -n 1 basename); do
  echo "Building package for architecture: $arch"

  # Clean the build directory and the destination directory before copying new files
  rm -rf build/
  rm -rf ${LIB_DEST_DIR}
  mkdir -p ${LIB_DEST_DIR}

  # Copy the architecture-specific files to the destination directory
  cp -r ${LIB_SOURCE_DIR}/${arch}/* ${LIB_DEST_DIR}/

  # Adjust the platform tag based on the architecture
  case "$arch" in
    darwin_amd64)
      platform_tag="macosx_10_9_x86_64"
      ;;
    darwin_arm64)
      platform_tag="macosx_11_0_arm64"
      ;;
    linux_amd64)
      platform_tag="manylinux1_x86_64"
      ;;
    windows_amd64)
      platform_tag="win_amd64"
      ;;
    *)
      echo "Unsupported architecture: $arch. Skipping."
      continue
      ;;
  esac

  # Set the correct environment variables for platform-specific builds
  export ARCHFLAGS="-arch ${arch##*_}"

  # Call setup.py to build the package
  python setup.py bdist_wheel --plat-name $platform_tag

  # Check if the build was successful
  if [ $? -ne 0 ]; then
    echo "Error: Failed to build package for architecture $arch. Skipping."
    continue
  fi
done

# Final cleanup
rm -rf ${LIB_DEST_DIR}

echo "All packages built successfully."