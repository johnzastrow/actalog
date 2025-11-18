#!/bin/bash

################################################################################
# PWA Health Check Script
################################################################################
# This script checks whether a website meets all the requirements to run as
# a Progressive Web App (PWA).
#
# What this script checks:
# 1. HTTPS connection (required for PWA)
# 2. Valid manifest.json file
# 3. Service worker registration
# 4. Manifest properties (name, icons, start_url, display)
# 5. Icon sizes and formats
# 6. Offline capability
# 7. Viewport meta tag
# 8. Theme color
# 9. Apple touch icons
# 10. Performance and accessibility scores (via Lighthouse if available)
#
# Usage:
#   chmod +x pwa-health-check.sh
#   ./pwa-health-check.sh
#   ./pwa-health-check.sh https://example.com
#
# Requirements:
#   - curl
#   - jq (for JSON parsing)
#   - openssl (for HTTPS checking)
#   - Optional: lighthouse (for detailed PWA scoring)
################################################################################

set -e  # Exit on error (we'll handle errors manually)
set -u  # Treat unset variables as an error

################################################################################
# Configuration
################################################################################

# Default timeout for curl requests (in seconds)
TIMEOUT=10

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Unicode symbols
CHECK_MARK="✓"
CROSS_MARK="✗"
WARNING="⚠"
INFO="ℹ"

# Test results counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_WARNING=0
TESTS_TOTAL=0

################################################################################
# Helper Functions
################################################################################

# Print colored status messages
print_header() {
    echo -e "\n${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}\n"
}

print_section() {
    echo -e "\n${BLUE}━━━ $1 ━━━${NC}\n"
}

print_pass() {
    echo -e "${GREEN}${CHECK_MARK}${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_fail() {
    echo -e "${RED}${CROSS_MARK}${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_warning() {
    echo -e "${YELLOW}${WARNING}${NC} $1"
    TESTS_WARNING=$((TESTS_WARNING + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

print_info() {
    echo -e "${BLUE}${INFO}${NC} $1"
}

print_detail() {
    echo -e "   ${NC}$1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check dependencies
check_dependencies() {
    local missing_deps=()

    if ! command_exists curl; then
        missing_deps+=("curl")
    fi

    if ! command_exists jq; then
        missing_deps+=("jq")
    fi

    if ! command_exists openssl; then
        missing_deps+=("openssl")
    fi

    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "${RED}Error: Missing required dependencies:${NC}"
        for dep in "${missing_deps[@]}"; do
            echo -e "  - $dep"
        done
        echo ""
        echo "Install with:"
        echo "  sudo apt-get install ${missing_deps[*]}"
        exit 1
    fi

    # Check for optional lighthouse
    if ! command_exists lighthouse; then
        print_info "Lighthouse not found (optional). Install with: npm install -g lighthouse"
    fi
}

# Normalize URL (add https:// if missing, remove trailing slash)
normalize_url() {
    local url="$1"

    # Remove trailing slash
    url="${url%/}"

    # Add https:// if no protocol specified
    if [[ ! "$url" =~ ^https?:// ]]; then
        url="https://$url"
    fi

    echo "$url"
}

################################################################################
# Test 1: HTTPS Check
################################################################################

check_https() {
    print_section "1. HTTPS Connection"

    if [[ "$TARGET_URL" =~ ^https:// ]]; then
        # Try to connect with HTTPS
        if curl -s -o /dev/null -w "%{http_code}" --max-time "$TIMEOUT" "$TARGET_URL" | grep -q "^[23]"; then
            print_pass "Site is served over HTTPS"

            # Check SSL certificate validity
            local domain=$(echo "$TARGET_URL" | sed -e 's|^https://||' -e 's|/.*||' -e 's|:.*||')
            if echo | openssl s_client -servername "$domain" -connect "$domain:443" 2>/dev/null | openssl x509 -noout -dates 2>/dev/null >/dev/null; then
                print_pass "SSL certificate is valid"
            else
                print_warning "Could not verify SSL certificate"
            fi
        else
            print_fail "Cannot connect to site via HTTPS"
            return 1
        fi
    else
        print_fail "Site is NOT served over HTTPS (PWA requires HTTPS)"
        print_detail "HTTP is only allowed for localhost during development"
        return 1
    fi
}

################################################################################
# Test 2: Fetch HTML and Parse
################################################################################

fetch_html() {
    print_section "2. Fetching Page Content"

    # Fetch HTML content
    HTML_CONTENT=$(curl -sL --max-time "$TIMEOUT" -H "User-Agent: Mozilla/5.0 (compatible; PWA-Health-Check/1.0)" "$TARGET_URL" 2>&1)

    if [ $? -eq 0 ] && [ -n "$HTML_CONTENT" ]; then
        print_pass "Successfully fetched page content"
        print_detail "Content size: $(echo "$HTML_CONTENT" | wc -c) bytes"
    else
        print_fail "Failed to fetch page content"
        return 1
    fi
}

################################################################################
# Test 3: Web App Manifest Check
################################################################################

check_manifest() {
    print_section "3. Web App Manifest"

    # Look for manifest link in HTML
    local manifest_path=$(echo "$HTML_CONTENT" | grep -oP '<link[^>]*rel=["\']manifest["\'][^>]*href=["\']?\K[^"'\'' >]+' | head -1)

    if [ -z "$manifest_path" ]; then
        print_fail "No manifest.json link found in HTML"
        print_detail "Add: <link rel=\"manifest\" href=\"/manifest.json\">"
        return 1
    fi

    print_pass "Manifest link found in HTML: $manifest_path"

    # Construct full manifest URL
    if [[ "$manifest_path" =~ ^https?:// ]]; then
        MANIFEST_URL="$manifest_path"
    elif [[ "$manifest_path" =~ ^/ ]]; then
        # Absolute path
        local base_url=$(echo "$TARGET_URL" | sed -E 's|(https?://[^/]+).*|\1|')
        MANIFEST_URL="$base_url$manifest_path"
    else
        # Relative path
        MANIFEST_URL="$TARGET_URL/$manifest_path"
    fi

    print_detail "Manifest URL: $MANIFEST_URL"

    # Fetch manifest
    MANIFEST_CONTENT=$(curl -sL --max-time "$TIMEOUT" "$MANIFEST_URL" 2>&1)

    if [ $? -ne 0 ] || [ -z "$MANIFEST_CONTENT" ]; then
        print_fail "Failed to fetch manifest.json"
        return 1
    fi

    # Validate JSON
    if ! echo "$MANIFEST_CONTENT" | jq empty 2>/dev/null; then
        print_fail "Manifest is not valid JSON"
        return 1
    fi

    print_pass "Manifest fetched and is valid JSON"

    # Check required fields
    check_manifest_fields
}

check_manifest_fields() {
    print_section "4. Manifest Properties"

    local manifest="$MANIFEST_CONTENT"

    # Check name or short_name
    local name=$(echo "$manifest" | jq -r '.name // empty' 2>/dev/null)
    local short_name=$(echo "$manifest" | jq -r '.short_name // empty' 2>/dev/null)

    if [ -n "$name" ]; then
        print_pass "Name: \"$name\""
    elif [ -n "$short_name" ]; then
        print_pass "Short name: \"$short_name\""
    else
        print_fail "Missing 'name' or 'short_name' property"
    fi

    # Check icons
    local icons_count=$(echo "$manifest" | jq '.icons | length' 2>/dev/null)
    if [ "$icons_count" -gt 0 ]; then
        print_pass "Icons defined: $icons_count icon(s)"

        # Check for required sizes
        local has_192=$(echo "$manifest" | jq '.icons[] | select(.sizes | contains("192x192"))' 2>/dev/null)
        local has_512=$(echo "$manifest" | jq '.icons[] | select(.sizes | contains("512x512"))' 2>/dev/null)

        if [ -n "$has_192" ]; then
            print_pass "Has 192x192 icon (required for installability)"
        else
            print_fail "Missing 192x192 icon (required for PWA)"
        fi

        if [ -n "$has_512" ]; then
            print_pass "Has 512x512 icon (required for splash screen)"
        else
            print_warning "Missing 512x512 icon (recommended for splash screen)"
        fi

        # List all icons
        print_detail "Icon details:"
        echo "$manifest" | jq -r '.icons[] | "   - \(.sizes) \(.type // "unknown type") \(.src)"' 2>/dev/null
    else
        print_fail "No icons defined in manifest"
    fi

    # Check start_url
    local start_url=$(echo "$manifest" | jq -r '.start_url // empty' 2>/dev/null)
    if [ -n "$start_url" ]; then
        print_pass "Start URL: \"$start_url\""
    else
        print_warning "Missing 'start_url' (recommended)"
    fi

    # Check display mode
    local display=$(echo "$manifest" | jq -r '.display // empty' 2>/dev/null)
    if [ -n "$display" ]; then
        print_pass "Display mode: \"$display\""
        if [[ "$display" =~ ^(standalone|fullscreen|minimal-ui)$ ]]; then
            print_detail "Good! Display mode is PWA-friendly"
        else
            print_warning "Display mode '$display' may not provide app-like experience"
        fi
    else
        print_warning "Missing 'display' property (recommended: standalone)"
    fi

    # Check theme_color
    local theme_color=$(echo "$manifest" | jq -r '.theme_color // empty' 2>/dev/null)
    if [ -n "$theme_color" ]; then
        print_pass "Theme color: $theme_color"
    else
        print_warning "Missing 'theme_color' (recommended)"
    fi

    # Check background_color
    local background_color=$(echo "$manifest" | jq -r '.background_color // empty' 2>/dev/null)
    if [ -n "$background_color" ]; then
        print_pass "Background color: $background_color"
    else
        print_warning "Missing 'background_color' (recommended for splash screen)"
    fi

    # Check description
    local description=$(echo "$manifest" | jq -r '.description // empty' 2>/dev/null)
    if [ -n "$description" ]; then
        print_pass "Description: \"${description:0:60}...\""
    else
        print_warning "Missing 'description' (recommended)"
    fi

    # Check scope
    local scope=$(echo "$manifest" | jq -r '.scope // empty' 2>/dev/null)
    if [ -n "$scope" ]; then
        print_info "Scope: \"$scope\""
    fi

    # Check orientation
    local orientation=$(echo "$manifest" | jq -r '.orientation // empty' 2>/dev/null)
    if [ -n "$orientation" ]; then
        print_info "Orientation: \"$orientation\""
    fi
}

################################################################################
# Test 5: Service Worker Check
################################################################################

check_service_worker() {
    print_section "5. Service Worker"

    # Look for service worker registration in HTML or JS
    # Check for common patterns: navigator.serviceWorker.register

    # First, check inline scripts and linked scripts
    local has_sw_registration=false

    # Check for inline service worker registration
    if echo "$HTML_CONTENT" | grep -q "serviceWorker.register\|navigator.serviceWorker"; then
        has_sw_registration=true
    fi

    # Check linked JavaScript files for service worker registration
    local js_files=$(echo "$HTML_CONTENT" | grep -oP '<script[^>]*src=["\']?\K[^"'\'' >]+\.js' | head -5)

    for js_file in $js_files; do
        # Construct full JS URL
        if [[ "$js_file" =~ ^https?:// ]]; then
            local js_url="$js_file"
        elif [[ "$js_file" =~ ^/ ]]; then
            local base_url=$(echo "$TARGET_URL" | sed -E 's|(https?://[^/]+).*|\1|')
            local js_url="$base_url$js_file"
        else
            local js_url="$TARGET_URL/$js_file"
        fi

        # Fetch and check JS file
        local js_content=$(curl -sL --max-time "$TIMEOUT" "$js_url" 2>/dev/null || true)
        if echo "$js_content" | grep -q "serviceWorker.register\|navigator.serviceWorker"; then
            has_sw_registration=true
            break
        fi
    done

    if [ "$has_sw_registration" = true ]; then
        print_pass "Service worker registration code found"
    else
        print_fail "No service worker registration found"
        print_detail "PWA requires a service worker for offline functionality"
        return 1
    fi

    # Try to find service worker file
    # Common names: sw.js, service-worker.js, firebase-messaging-sw.js
    local sw_files=("sw.js" "service-worker.js" "serviceworker.js" "firebase-messaging-sw.js")
    local found_sw=false

    local base_url=$(echo "$TARGET_URL" | sed -E 's|(https?://[^/]+).*|\1|')

    for sw_file in "${sw_files[@]}"; do
        local sw_url="$base_url/$sw_file"
        local status_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "$TIMEOUT" "$sw_url" 2>/dev/null || echo "000")

        if [[ "$status_code" =~ ^2 ]]; then
            print_pass "Service worker file found: $sw_file (HTTP $status_code)"
            found_sw=true

            # Check service worker content
            local sw_content=$(curl -sL --max-time "$TIMEOUT" "$sw_url" 2>/dev/null || true)

            # Check for fetch event handler (required for offline support)
            if echo "$sw_content" | grep -q "addEventListener.*fetch\|self.addEventListener('fetch'"; then
                print_pass "Service worker has fetch event handler (offline support)"
            else
                print_warning "Service worker may not handle fetch events (limited offline support)"
            fi

            # Check for install event
            if echo "$sw_content" | grep -q "addEventListener.*install\|self.addEventListener('install'"; then
                print_pass "Service worker has install event handler"
            fi

            # Check for activate event
            if echo "$sw_content" | grep -q "addEventListener.*activate\|self.addEventListener('activate'"; then
                print_pass "Service worker has activate event handler"
            fi

            break
        fi
    done

    if [ "$found_sw" = false ]; then
        print_warning "Could not locate service worker file (checked: ${sw_files[*]})"
        print_detail "Service worker might use a different filename"
    fi
}

################################################################################
# Test 6: HTML Meta Tags
################################################################################

check_meta_tags() {
    print_section "6. HTML Meta Tags"

    # Viewport meta tag
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name=["\']viewport["\']'; then
        print_pass "Viewport meta tag present"
        local viewport=$(echo "$HTML_CONTENT" | grep -oP '<meta[^>]*name=["\']viewport["\'][^>]*content=["\']?\K[^"'\'' >]+' | head -1)
        print_detail "Content: $viewport"
    else
        print_fail "Missing viewport meta tag (required for responsive PWA)"
        print_detail "Add: <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">"
    fi

    # Theme color meta tag
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name=["\']theme-color["\']'; then
        local theme_color=$(echo "$HTML_CONTENT" | grep -oP '<meta[^>]*name=["\']theme-color["\'][^>]*content=["\']?\K[^"'\'' >]+' | head -1)
        print_pass "Theme color meta tag: $theme_color"
    else
        print_warning "Missing theme-color meta tag (recommended)"
    fi

    # Description meta tag
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name=["\']description["\']'; then
        local description=$(echo "$HTML_CONTENT" | grep -oP '<meta[^>]*name=["\']description["\'][^>]*content=["\']?\K[^"'\'' >]+' | head -1)
        print_pass "Description meta tag present"
        print_detail "${description:0:80}..."
    else
        print_warning "Missing description meta tag (recommended for SEO)"
    fi

    # Apple mobile web app capable
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name=["\']apple-mobile-web-app-capable["\']'; then
        print_pass "Apple mobile web app capable tag present"
    else
        print_info "Missing apple-mobile-web-app-capable (optional for iOS)"
    fi

    # Apple touch icon
    if echo "$HTML_CONTENT" | grep -q '<link[^>]*rel=["\']apple-touch-icon["\']'; then
        print_pass "Apple touch icon present (good for iOS)"
    else
        print_info "Missing apple-touch-icon (optional for iOS home screen)"
    fi
}

################################################################################
# Test 7: Lighthouse PWA Check (Optional)
################################################################################

run_lighthouse() {
    print_section "7. Lighthouse PWA Audit (Optional)"

    if ! command_exists lighthouse; then
        print_info "Lighthouse not installed, skipping detailed PWA audit"
        print_detail "Install: npm install -g lighthouse"
        return 0
    fi

    print_info "Running Lighthouse PWA audit (this may take 30-60 seconds)..."

    # Create temp directory for report
    local temp_dir=$(mktemp -d)
    local report_file="$temp_dir/lighthouse-report.json"

    # Run Lighthouse in headless mode
    if lighthouse "$TARGET_URL" \
        --only-categories=pwa \
        --output=json \
        --output-path="$report_file" \
        --chrome-flags="--headless --no-sandbox" \
        --quiet 2>/dev/null; then

        # Parse Lighthouse results
        local pwa_score=$(jq -r '.categories.pwa.score * 100' "$report_file" 2>/dev/null || echo "0")

        if [ "$pwa_score" != "null" ] && [ -n "$pwa_score" ]; then
            if (( $(echo "$pwa_score >= 90" | bc -l) )); then
                print_pass "Lighthouse PWA Score: ${pwa_score}% (Excellent!)"
            elif (( $(echo "$pwa_score >= 70" | bc -l) )); then
                print_warning "Lighthouse PWA Score: ${pwa_score}% (Good, but can be improved)"
            else
                print_fail "Lighthouse PWA Score: ${pwa_score}% (Needs improvement)"
            fi

            # Show failed audits
            print_detail "Key findings:"
            jq -r '.categories.pwa.auditRefs[] |
                select(.weight > 0) |
                .id as $id |
                ($ENV.LIGHTHOUSE_AUDITS | fromjson | .[$id]) |
                select(.score != null and .score < 1) |
                "   ✗ \(.title)"' \
                --arg LIGHTHOUSE_AUDITS "$(jq '.audits' "$report_file")" \
                "$report_file" 2>/dev/null || true
        else
            print_warning "Could not parse Lighthouse PWA score"
        fi
    else
        print_warning "Lighthouse audit failed or timed out"
    fi

    # Clean up
    rm -rf "$temp_dir"
}

################################################################################
# Test 8: Performance Basics
################################################################################

check_performance() {
    print_section "8. Basic Performance Check"

    # Measure page load time
    local start_time=$(date +%s%N)
    curl -sL --max-time "$TIMEOUT" -o /dev/null "$TARGET_URL" 2>/dev/null
    local end_time=$(date +%s%N)

    local load_time_ms=$(( (end_time - start_time) / 1000000 ))

    if [ "$load_time_ms" -lt 1000 ]; then
        print_pass "Page load time: ${load_time_ms}ms (Excellent!)"
    elif [ "$load_time_ms" -lt 3000 ]; then
        print_warning "Page load time: ${load_time_ms}ms (Acceptable, but could be faster)"
    else
        print_fail "Page load time: ${load_time_ms}ms (Too slow for good UX)"
    fi

    # Check for HTTP/2
    local http_version=$(curl -sI --http2 "$TARGET_URL" 2>/dev/null | grep -i "^http/" | head -1)
    if echo "$http_version" | grep -q "HTTP/2"; then
        print_pass "HTTP/2 supported (better performance)"
    else
        print_warning "HTTP/2 not detected (recommended for PWA performance)"
    fi

    # Check for compression
    local content_encoding=$(curl -sI -H "Accept-Encoding: gzip, deflate, br" "$TARGET_URL" 2>/dev/null | grep -i "^content-encoding:" | cut -d: -f2 | tr -d ' \r')
    if [ -n "$content_encoding" ]; then
        print_pass "Compression enabled: $content_encoding"
    else
        print_warning "No compression detected (recommended: gzip or brotli)"
    fi
}

################################################################################
# Final Report
################################################################################

print_summary() {
    print_header "PWA Health Check Summary"

    local total=$TESTS_TOTAL
    local passed=$TESTS_PASSED
    local failed=$TESTS_FAILED
    local warnings=$TESTS_WARNING

    echo -e "Target URL: ${CYAN}$TARGET_URL${NC}\n"

    echo -e "Results:"
    echo -e "  ${GREEN}Passed:${NC}   $passed"
    echo -e "  ${RED}Failed:${NC}   $failed"
    echo -e "  ${YELLOW}Warnings:${NC} $warnings"
    echo -e "  ${BLUE}Total:${NC}    $total"
    echo ""

    # Calculate percentage
    if [ "$total" -gt 0 ]; then
        local pass_percentage=$(( (passed * 100) / total ))

        if [ "$failed" -eq 0 ] && [ "$pass_percentage" -ge 90 ]; then
            echo -e "${GREEN}${CHECK_MARK} PWA Status: EXCELLENT${NC}"
            echo -e "Your site meets all critical PWA requirements and is ready for installation!"
        elif [ "$failed" -eq 0 ]; then
            echo -e "${GREEN}${CHECK_MARK} PWA Status: GOOD${NC}"
            echo -e "Your site meets PWA requirements. Consider addressing warnings for optimal experience."
        elif [ "$failed" -le 2 ]; then
            echo -e "${YELLOW}${WARNING} PWA Status: NEEDS IMPROVEMENT${NC}"
            echo -e "Your site has some PWA features but needs fixes to be fully installable."
        else
            echo -e "${RED}${CROSS_MARK} PWA Status: NOT A PWA${NC}"
            echo -e "Your site does not meet PWA requirements. Address the failed checks above."
        fi
    fi

    echo ""
    echo -e "${BLUE}${INFO} Next Steps:${NC}"
    echo "1. Fix any failed checks (marked with ${RED}✗${NC})"
    echo "2. Address warnings for better user experience (marked with ${YELLOW}⚠${NC})"
    echo "3. Test installation on mobile devices"
    echo "4. Test offline functionality"
    echo "5. Consider running full Lighthouse audit for detailed analysis"
    echo ""
}

################################################################################
# Main Execution
################################################################################

main() {
    # Print banner
    clear
    echo -e "${MAGENTA}"
    cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║              PWA Health Check Tool v1.0                       ║
║              Progressive Web App Validator                    ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"

    # Check dependencies
    check_dependencies

    # Get URL from argument or prompt user
    if [ $# -eq 1 ]; then
        TARGET_URL=$(normalize_url "$1")
    else
        echo -e "${BLUE}Enter the URL to check:${NC}"
        read -p "> " user_url

        if [ -z "$user_url" ]; then
            echo -e "${RED}Error: URL cannot be empty${NC}"
            exit 1
        fi

        TARGET_URL=$(normalize_url "$user_url")
    fi

    echo -e "\n${CYAN}Checking:${NC} $TARGET_URL\n"

    # Run all checks
    set +e  # Don't exit on error, we want to run all tests

    check_https
    fetch_html
    check_manifest
    check_service_worker
    check_meta_tags
    run_lighthouse
    check_performance

    set -e

    # Print summary
    print_summary
}

# Run main function with all arguments
main "$@"
