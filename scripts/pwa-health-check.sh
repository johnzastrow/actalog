#!/bin/bash

################################################################################
# PWA Health Check Script
# Version: 1.1.1
################################################################################
# This script checks whether a website meets Progressive Web App requirements.
#
# Usage:
#   ./pwa-health-check.sh [URL]
#
# Requirements:
#   - curl, jq, openssl
################################################################################

set -e
set -u

# Script version
SCRIPT_VERSION="1.1.1"

# Configuration
TIMEOUT=10

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Symbols
CHECK_MARK="✓"
CROSS_MARK="✗"
WARNING="⚠"
INFO="ℹ"

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_WARNING=0

# Global variables
TARGET_URL=""
HTML_CONTENT=""
MANIFEST_CONTENT=""

################################################################################
# Helper Functions
################################################################################

print_header() {
    echo -e "\n${MAGENTA}═══════════════════════════════════════${NC}"
    echo -e "${MAGENTA}$1${NC}"
    echo -e "${MAGENTA}═══════════════════════════════════════${NC}\n"
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
    echo -e "${CYAN}${INFO}${NC} $1"
}

print_detail() {
    echo -e "   ${NC}$1${NC}"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

################################################################################
# Dependency Checks
################################################################################

check_dependencies() {
    local missing_deps=()

    for cmd in curl jq openssl; do
        if ! command_exists "$cmd"; then
            missing_deps+=("$cmd")
        fi
    done

    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "${RED}Error: Missing required dependencies: ${missing_deps[*]}${NC}"
        echo "Install with: sudo apt-get install ${missing_deps[*]}"
        exit 1
    fi
}

normalize_url() {
    local url="$1"

    # Add https:// if no protocol specified
    if [[ ! "$url" =~ ^https?:// ]]; then
        url="https://$url"
    fi

    # Remove trailing slash
    url="${url%/}"

    echo "$url"
}

################################################################################
# Test 1: HTTPS Check
################################################################################

check_https() {
    print_section "1. HTTPS Security"

    if [[ "$TARGET_URL" =~ ^https:// ]]; then
        if curl -sI --max-time "$TIMEOUT" "$TARGET_URL" >/dev/null 2>&1; then
            print_pass "Site is served over HTTPS"
        else
            print_fail "Cannot connect to site via HTTPS"
            return 1
        fi
    else
        print_fail "Site is NOT served over HTTPS - PWA requires HTTPS"
        print_detail "HTTP is only allowed for localhost during development"
        return 1
    fi
}

################################################################################
# Test 2: Fetch HTML
################################################################################

fetch_html() {
    print_section "2. Fetching Site Content"

    HTML_CONTENT=$(curl -sL --max-time "$TIMEOUT" "$TARGET_URL" 2>/dev/null)

    if [ -n "$HTML_CONTENT" ]; then
        print_pass "Successfully fetched HTML content"
    else
        print_fail "Failed to fetch HTML content"
        return 1
    fi
}

################################################################################
# Test 3: Manifest Check
################################################################################

check_manifest() {
    print_section "3. Web App Manifest"

    # Find manifest link in HTML
    local manifest_link=$(echo "$HTML_CONTENT" | grep -oP '<link[^>]*rel="manifest"[^>]*href="\K[^"]+' | head -1)

    if [ -z "$manifest_link" ]; then
        print_fail "No manifest link found in HTML"
        return 1
    fi

    print_pass "Manifest link found: $manifest_link"

    # Construct full manifest URL
    local manifest_url
    if [[ "$manifest_link" =~ ^https?:// ]]; then
        manifest_url="$manifest_link"
    elif [[ "$manifest_link" =~ ^/ ]]; then
        local base_url=$(echo "$TARGET_URL" | grep -oP '^https?://[^/]+')
        manifest_url="$base_url$manifest_link"
    else
        manifest_url="$TARGET_URL/$manifest_link"
    fi

    # Fetch manifest
    MANIFEST_CONTENT=$(curl -sL --max-time "$TIMEOUT" "$manifest_url" 2>/dev/null)

    if [ -z "$MANIFEST_CONTENT" ]; then
        print_fail "Failed to fetch manifest file"
        return 1
    fi

    # Validate JSON
    if echo "$MANIFEST_CONTENT" | jq empty 2>/dev/null; then
        print_pass "Manifest is valid JSON"
    else
        print_fail "Manifest is not valid JSON"
        return 1
    fi

    check_manifest_fields
}

check_manifest_fields() {
    local manifest="$MANIFEST_CONTENT"

    # Check name
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
    local icons_count=$(echo "$manifest" | jq '.icons | length' 2>/dev/null || echo "0")

    if [ "$icons_count" -gt 0 ]; then
        print_pass "Icons defined: $icons_count icons"

        # Check for required sizes
        local has_192=$(echo "$manifest" | jq '.icons[] | select(.sizes | contains("192x192"))' 2>/dev/null)
        local has_512=$(echo "$manifest" | jq '.icons[] | select(.sizes | contains("512x512"))' 2>/dev/null)

        if [ -n "$has_192" ]; then
            print_pass "Has 192x192 icon - required for installability"
        else
            print_fail "Missing 192x192 icon - required for PWA"
        fi

        if [ -n "$has_512" ]; then
            print_pass "Has 512x512 icon - required for splash screen"
        else
            print_warning "Missing 512x512 icon - recommended for splash screen"
        fi
    else
        print_fail "No icons defined in manifest"
    fi

    # Check start_url
    local start_url=$(echo "$manifest" | jq -r '.start_url // empty' 2>/dev/null)
    if [ -n "$start_url" ]; then
        print_pass "Start URL: \"$start_url\""
    else
        print_warning "Missing 'start_url' - recommended"
    fi

    # Check display mode
    local display=$(echo "$manifest" | jq -r '.display // empty' 2>/dev/null)
    if [ -n "$display" ]; then
        print_pass "Display mode: \"$display\""
        if [[ "$display" =~ ^(standalone|fullscreen|minimal-ui)$ ]]; then
            print_detail "Good! Display mode is PWA-friendly"
        fi
    else
        print_warning "Missing 'display' property - recommended: standalone"
    fi

    # Check theme_color
    local theme_color=$(echo "$manifest" | jq -r '.theme_color // empty' 2>/dev/null)
    if [ -n "$theme_color" ]; then
        print_pass "Theme color: $theme_color"
    else
        print_warning "Missing 'theme_color' - recommended"
    fi

    # Check background_color
    local background_color=$(echo "$manifest" | jq -r '.background_color // empty' 2>/dev/null)
    if [ -n "$background_color" ]; then
        print_pass "Background color: $background_color"
    else
        print_warning "Missing 'background_color' - recommended for splash screen"
    fi

    # Check description
    local description=$(echo "$manifest" | jq -r '.description // empty' 2>/dev/null)
    if [ -n "$description" ]; then
        print_pass "Description present"
    else
        print_warning "Missing 'description' - recommended"
    fi
}

################################################################################
# Test 4: Service Worker
################################################################################

check_service_worker() {
    print_section "4. Service Worker"

    # Common service worker filenames
    local sw_files=("service-worker.js" "sw.js" "serviceworker.js" "firebase-messaging-sw.js")
    local found_sw=false
    local base_url=$(echo "$TARGET_URL" | grep -oP '^https?://[^/]+')

    # Check if service worker is registered in HTML
    if echo "$HTML_CONTENT" | grep -q "serviceWorker.register"; then
        print_pass "Service worker registration found in HTML"
    else
        print_warning "No service worker registration found in HTML"
    fi

    # Try to fetch common service worker files
    for sw_file in "${sw_files[@]}"; do
        local sw_url="$base_url/$sw_file"
        local status_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "$TIMEOUT" "$sw_url" 2>/dev/null || echo "000")

        if [[ "$status_code" =~ ^2 ]]; then
            print_pass "Service worker file found: $sw_file - HTTP $status_code"
            found_sw=true

            # Check service worker content
            local sw_content=$(curl -sL --max-time "$TIMEOUT" "$sw_url" 2>/dev/null || true)

            if echo "$sw_content" | grep -q "addEventListener.*fetch"; then
                print_pass "Service worker has fetch event handler - offline support"
            else
                print_warning "Service worker may not handle fetch events - limited offline"
            fi

            if echo "$sw_content" | grep -q "addEventListener.*install"; then
                print_pass "Service worker has install event handler"
            fi

            if echo "$sw_content" | grep -q "addEventListener.*activate"; then
                print_pass "Service worker has activate event handler"
            fi

            break
        fi
    done

    if [ "$found_sw" = false ]; then
        print_warning "Could not locate service worker file"
    fi
}

################################################################################
# Test 5: HTML Meta Tags
################################################################################

check_meta_tags() {
    print_section "5. HTML Meta Tags"

    # Viewport meta tag
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name="viewport"'; then
        print_pass "Viewport meta tag present"
    else
        print_fail "Missing viewport meta tag - required for responsive PWA"
    fi

    # Theme color meta tag
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name="theme-color"'; then
        local theme_color=$(echo "$HTML_CONTENT" | grep -oP '<meta[^>]*name="theme-color"[^>]*content="\K[^"]+' | head -1)
        print_pass "Theme color meta tag: $theme_color"
    else
        print_warning "Missing theme-color meta tag - recommended"
    fi

    # Description meta tag
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name="description"'; then
        print_pass "Description meta tag present"
    else
        print_warning "Missing description meta tag - recommended for SEO"
    fi

    # Apple mobile web app capable
    if echo "$HTML_CONTENT" | grep -q '<meta[^>]*name="apple-mobile-web-app-capable"'; then
        print_pass "Apple mobile web app capable tag present"
    else
        print_info "Missing apple-mobile-web-app-capable - optional for iOS"
    fi

    # Apple touch icon
    if echo "$HTML_CONTENT" | grep -q '<link[^>]*rel="apple-touch-icon"'; then
        print_pass "Apple touch icon present - good for iOS"
    else
        print_info "Missing apple-touch-icon - optional for iOS home screen"
    fi
}

################################################################################
# Test 6: Basic Performance
################################################################################

check_performance() {
    print_section "6. Basic Performance Check"

    # Measure page load time
    local start_time=$(date +%s%N)
    curl -sL --max-time "$TIMEOUT" "$TARGET_URL" >/dev/null 2>&1
    local end_time=$(date +%s%N)

    local load_time_ms=$(( (end_time - start_time) / 1000000 ))

    if [ "$load_time_ms" -lt 1000 ]; then
        print_pass "Page load time: ${load_time_ms}ms - Excellent!"
    elif [ "$load_time_ms" -lt 3000 ]; then
        print_warning "Page load time: ${load_time_ms}ms - Acceptable, could be faster"
    else
        print_fail "Page load time: ${load_time_ms}ms - Too slow for good UX"
    fi

    # Check for HTTP/2
    local http_version=$(curl -sI --http2 "$TARGET_URL" 2>/dev/null | grep -i "^http/" | head -1)
    if echo "$http_version" | grep -q "HTTP/2"; then
        print_pass "HTTP/2 supported - better performance"
    else
        print_warning "HTTP/2 not detected - recommended for PWA performance"
    fi

    # Check for compression
    local content_encoding=$(curl -sI -H "Accept-Encoding: gzip, deflate, br" "$TARGET_URL" 2>/dev/null | grep -i "^content-encoding:" | cut -d: -f2 | tr -d ' \r')
    if [ -n "$content_encoding" ]; then
        print_pass "Compression enabled: $content_encoding"
    else
        print_warning "No compression detected - recommended: gzip or brotli"
    fi
}

################################################################################
# Test 7: Lighthouse PWA Audit (Optional)
################################################################################

run_lighthouse() {
    print_section "7. Lighthouse PWA Audit - Optional"

    if ! command_exists lighthouse; then
        print_info "Lighthouse not installed - skipping detailed PWA audit"
        print_detail "Install with: npm install -g lighthouse"
        return 0
    fi

    print_info "Running Lighthouse PWA audit - this may take 30-60 seconds..."

    # Create temp directory for report
    local temp_dir=$(mktemp -d)
    local report_file="$temp_dir/lighthouse-report.json"

    # Run Lighthouse in headless mode (v13+ removed pwa category, now part of best-practices)
    if lighthouse "$TARGET_URL" \
        --output=json \
        --output-path="$report_file" \
        --chrome-flags="--headless --no-sandbox --disable-gpu" \
        --quiet 2>/dev/null; then

        # Parse Lighthouse results
        # In Lighthouse v13+, check best-practices category for PWA-related audits
        local best_practices_score=$(jq -r '.categories["best-practices"].score * 100' "$report_file" 2>/dev/null || echo "0")

        if [ "$best_practices_score" != "null" ] && [ -n "$best_practices_score" ]; then
            # Convert to integer for comparison
            local score_int=${best_practices_score%.*}

            if [ "$score_int" -ge 90 ]; then
                print_pass "Lighthouse Best Practices Score: ${best_practices_score}% - Excellent!"
            elif [ "$score_int" -ge 70 ]; then
                print_warning "Lighthouse Best Practices Score: ${best_practices_score}% - Good, but can be improved"
            else
                print_fail "Lighthouse Best Practices Score: ${best_practices_score}% - Needs improvement"
            fi

            # Show PWA-related audits from best-practices
            print_detail "PWA-related findings:"

            # Check for specific PWA audits
            local pwa_audit_ids=("viewport" "service-worker" "installable-manifest" "splash-screen" "themed-omnibox" "maskable-icon")
            local found_issues=false

            for audit_id in "${pwa_audit_ids[@]}"; do
                local audit_score=$(jq -r ".audits.\"$audit_id\".score" "$report_file" 2>/dev/null)
                local audit_title=$(jq -r ".audits.\"$audit_id\".title" "$report_file" 2>/dev/null)

                if [ "$audit_score" != "null" ] && [ "$audit_title" != "null" ]; then
                    if [ "$audit_score" != "1" ]; then
                        found_issues=true
                        if [ "$audit_score" == "0" ]; then
                            echo -e "   ${RED}✗${NC} $audit_title"
                        else
                            echo -e "   ${YELLOW}⚠${NC} $audit_title - partial pass"
                        fi
                    fi
                fi
            done

            if [ "$found_issues" = false ]; then
                print_detail "All PWA audits passed!"
            fi

            # Show performance metrics
            local performance_score=$(jq -r '.categories.performance.score * 100' "$report_file" 2>/dev/null || echo "N/A")
            if [ "$performance_score" != "null" ] && [ "$performance_score" != "N/A" ]; then
                print_detail "Performance Score: ${performance_score}%"
            fi

            local accessibility_score=$(jq -r '.categories.accessibility.score * 100' "$report_file" 2>/dev/null || echo "N/A")
            if [ "$accessibility_score" != "null" ] && [ "$accessibility_score" != "N/A" ]; then
                print_detail "Accessibility Score: ${accessibility_score}%"
            fi

            local best_practices_score=$(jq -r '.categories["best-practices"].score * 100' "$report_file" 2>/dev/null || echo "N/A")
            if [ "$best_practices_score" != "null" ] && [ "$best_practices_score" != "N/A" ]; then
                print_detail "Best Practices Score: ${best_practices_score}%"
            fi

            local seo_score=$(jq -r '.categories.seo.score * 100' "$report_file" 2>/dev/null || echo "N/A")
            if [ "$seo_score" != "null" ] && [ "$seo_score" != "N/A" ]; then
                print_detail "SEO Score: ${seo_score}%"
            fi

            # Offer to save full report
            print_detail "Full report saved to: $report_file"
            print_info "View HTML report with: lighthouse $TARGET_URL --view"
        else
            print_warning "Could not parse Lighthouse PWA score"
        fi
    else
        print_warning "Lighthouse audit failed or timed out"
        print_detail "This may happen if Chrome/Chromium is not installed"
    fi

    # Clean up (optional - keep for debugging)
    # rm -rf "$temp_dir"
}

################################################################################
# Summary Report
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

    # Calculate status
    if [ "$total" -gt 0 ]; then
        local pass_percentage=$(( (passed * 100) / total ))

        if [ "$failed" -eq 0 ] && [ "$pass_percentage" -ge 90 ]; then
            echo -e "${GREEN}${CHECK_MARK} PWA Status: EXCELLENT${NC}"
            echo -e "Your site meets all critical PWA requirements and is ready for installation!"
        elif [ "$failed" -eq 0 ]; then
            echo -e "${GREEN}${CHECK_MARK} PWA Status: GOOD${NC}"
            echo -e "Your site meets PWA requirements. Consider addressing warnings."
        elif [ "$failed" -le 2 ]; then
            echo -e "${YELLOW}${WARNING} PWA Status: NEEDS IMPROVEMENT${NC}"
            echo -e "Your site has some PWA features but needs fixes to be installable."
        else
            echo -e "${RED}${CROSS_MARK} PWA Status: NOT A PWA${NC}"
            echo -e "Your site does not meet PWA requirements. Address the failed checks."
        fi
    fi

    echo ""
    echo -e "${BLUE}${INFO} Next Steps:${NC}"
    echo -e "1. Fix any failed checks marked with ${RED}✗${NC}"
    echo -e "2. Address warnings for better user experience marked with ${YELLOW}⚠${NC}"
    echo -e "3. Test installation on mobile devices"
    echo -e "4. Test offline functionality"
    echo -e "5. Review Lighthouse audit results if available above"
    echo -e "6. Generate full HTML Lighthouse report: lighthouse $TARGET_URL --view"
    echo ""
}

################################################################################
# Main Execution
################################################################################

main() {
    # Print banner
    clear
    echo -e "${MAGENTA}"
    cat << 'BANNER'
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║       PWA Health Check Tool v1.1.1                            ║
║       Progressive Web App Validator                           ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
BANNER
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

    # Run all checks (don't exit on error)
    set +e

    check_https
    fetch_html
    check_manifest
    check_service_worker
    check_meta_tags
    check_performance
    run_lighthouse

    set -e

    # Print summary
    print_summary
}

# Run main function with all arguments
main "$@"
