#!/bin/bash
# Script to download fonts for ActaLog
# Run this script from the web/public/fonts directory

# OpenDyslexic - download from opendyslexic.org
echo "Downloading OpenDyslexic..."
mkdir -p opendyslexic
curl -sL -o /tmp/opendyslexic.zip "https://github.com/antijingoist/opendyslexic/releases/download/v0.91.12/opendyslexic-0.91.12-webfont.tar.gz"
tar -xzf /tmp/opendyslexic.zip -C /tmp/
cp /tmp/opendyslexic-0.91.12-webfont/OpenDyslexic-Regular.woff2 opendyslexic/ 2>/dev/null || echo "  Regular not found"
cp /tmp/opendyslexic-0.91.12-webfont/OpenDyslexic-Bold.woff2 opendyslexic/ 2>/dev/null || echo "  Bold not found"

# Atkinson Hyperlegible - from Braille Institute
echo "Downloading Atkinson Hyperlegible..."
mkdir -p atkinson-hyperlegible
curl -sL -o /tmp/atkinson.zip "https://www.brailleinstitute.org/wp-content/uploads/2020/11/Atkinson-Hyperlegible-Font-Print-and-Web-2020-0514.zip"
unzip -q /tmp/atkinson.zip -d /tmp/atkinson/
find /tmp/atkinson -name "*.woff2" -exec cp {} atkinson-hyperlegible/ \;

# Source Serif Pro - use google-webfonts-helper manually
echo ""
echo "For Source Serif Pro, visit: https://gwfh.mranftl.com/fonts/source-serif-pro?subsets=latin"
echo "Download Regular and SemiBold weights in woff2 format"
echo ""

echo "Font download complete. Check each directory for files."
ls -la */
