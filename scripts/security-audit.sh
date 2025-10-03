#!/bin/bash

# Security Audit and Vulnerability Testing Script
# This script performs comprehensive security checks on the NFT Platform codebase

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔒 Starting Security Audit and Vulnerability Testing${NC}"
echo "=================================================="

# Create results directory
mkdir -p security-results
timestamp=$(date +%Y%m%d_%H%M%S)
report_file="security-results/security_audit_${timestamp}.md"

# Initialize security report
cat > "$report_file" << EOF
# NFT Platform Security Audit Report

**Date**: $(date '+%Y-%m-%d %H:%M:%S')
**Environment**: $(go version)
**Audit Tool**: Custom Security Audit Script

## Executive Summary

EOF

# Function to record security finding
record_finding() {
    local category="$1"
    local severity="$2"
    local issue="$3"
    local location="$4"
    local recommendation="$5"

    local severity_icon="🟡"
    case $severity in
        "HIGH") severity_icon="🔴" ;;
        "MEDIUM") severity_icon="🟡" ;;
        "LOW") severity_icon="🟢" ;;
    esac

    echo "Auditing: $category - $severity - $issue"

    echo "## $severity_icon $severity: $issue" >> "$report_file"
    echo "**Location**: \`$location\`" >> "$report_file"
    echo "**Issue**: $issue" >> "$report_file"
    "**Recommendation**: $recommendation >> "$report_file"
    echo "" >> "$report_file"
}

# Function to check file permissions
check_file_permissions() {
    echo -e "${YELLOW}🔐 Checking File Permissions...${NC}"

    # Check for executable files with excessive permissions
    find . -name "*.sh" -perm +0077 -type f 2>/dev/null | while read file; do
        record_finding "File Permissions" "HIGH" "World-writable executable script" "$file" "Remove world-write permissions: chmod 755 $file"
    done

    # Check for configuration files with sensitive data
    find . -name "*.env*" -o -name "*.config*" -type f 2>/dev/null | while read file; do
        if [ -r "$file" ]; then
            # Check for common secrets
            if grep -q -i "password\|secret\|key\|token" "$file"; then
                record_finding "Sensitive Data Exposure" "HIGH" "Sensitive data in config file" "$file" "Review file contents and move sensitive data to environment variables"
            fi
        fi
    done
    echo ""
}

# Function to check for hardcoded secrets
check_hardcoded_secrets() {
    echo -e "${YELLOW}🔑 Checking for Hardcoded Secrets...${NC}"

    # Check for hardcoded private keys
    grep -r "0x[a-fA-F0-9]\{64\}" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -v "//.*mock\|//.*test" "$file" > /dev/null; then
            record_finding "Hardcoded Secrets" "HIGH" "Hardcoded private key detected" "$file" "Use environment variables or secure key management"
        fi
    done

    # Check for hardcoded passwords
    grep -r "password.*=" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -v "//.*mock\|//.*test" "$file" > /dev/null; then
            record_finding "Hardcoded Secrets" "HIGH" "Hardcoded password detected" "$file" "Use environment variables or secure authentication"
        fi
    done

    # Check for hardcoded tokens
    grep -r "token.*=" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -v "//.*mock\|//.*test" "$file" > /dev/null; then
            record_finding "Hardcoded Secrets" "MEDIUM" "Hardcoded token detected" "$file" "Use environment variables or token management"
        fi
    done
    echo ""
}

# Function to check for SQL injection vulnerabilities
check_sql_injection() {
    echo -e "${YELLOW}🗄️ Checking for SQL Injection Vulnerabilities...${NC}"

    # Check for string concatenation in SQL queries
    grep -r "fmt\.Sprintf.*SELECT\|fmt\.Sprintf.*INSERT\|fmt\.Sprintf.*UPDATE\|fmt\.Sprintf.*DELETE" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -v "//.*safe\|//.*parametrized" "$file" > /dev/null; then
            record_finding "SQL Injection" "HIGH" "Unparameterized SQL query" "$file" "Use parameterized queries with prepared statements"
        fi
    done

    # Check for dynamic SQL construction
    grep -r "SELECT.*\+.*FROM" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d -f1)
        if grep -v "//.*safe\|//.*prepared\|//.*parameterized" "$file" > /dev/null; then
            record_finding "SQL Injection" "HIGH" "Dynamic SQL construction" "$file" "Use prepared statements or ORM methods"
        fi
    done
    echo ""
}

# Function to check for XSS vulnerabilities
check_xss() {
    echo -e "${YELLOW}🌐 Checking for XSS Vulnerabilities...${NC}"

    # Check for direct HTML rendering from user input
    grep -r "fmt\.Fprint\|fmt\.Fprintf.*<.*>" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d -f1)
        if grep -v "//.*sanitized\|//.*escaped" "$file" > /dev/null; then
            record_finding "XSS" "HIGH" "Unescaped HTML rendering" "$file" "Use HTML templating with proper escaping"
        fi
    done

    # Check for JavaScript in user content
    grep -r "javascript:" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cut -d: -f1)
        if grep -v "//.*sanitized\|//.*safe" "$file" > /dev/null; then
            record_finding "XSS" "MEDIUM" "JavaScript injection risk" "$file" "Use proper content security policies"
        fi
    done
    echo ""
}

# Function to check for insecure dependencies
check_dependencies() {
    echo -e "${YELLOW}📦 Checking for Insecure Dependencies...${NC}"

    # Check for known vulnerable packages
    if [ -f "go.mod" ]; then
        while read line; do
            if [[ "$line" =~ (github\.com/.*gopkg\.insecure|github\.com/known-vuln) ]]; then
                pkg=$(echo "$line" | awk '{print $2}')
                record_finding "Insecure Dependency" "HIGH" "Known vulnerable dependency: $pkg" "Update to secure version or remove dependency"
            fi
        done < <(go mod list -m all)
    fi

    # Check for outdated Go versions in dependencies
    if [ -f "go.mod" ]; then
        go list -m -u all 2>/dev/null | grep -E "\[(v0\.[0-9]+)\]" | while read line; do
            version=$(echo "$line" | grep -o "v[0-9]\+\.[0-9]+")
            if [[ "$version" < "v1.19.0" ]]; then
                pkg=$(echo "$line" | awk '{print $1}')
                record_finding "Outdated Dependency" "MEDIUM" "Outdated Go version: $version" "Update to latest stable version"
            fi
        done
    fi
    echo ""
}

# Function to check for cryptographic issues
check_cryptography() {
    echo -e "${YELLOW}🔐 Checking Cryptographic Implementations...${NC}"

    # Check for weak cryptographic algorithms
    grep -r "MD5\|sha1\|SHA1" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d: -f1)
        if grep -v "//.*hash\|//.*checksum\|//.*legacy" "$file" > /dev/null; then
            record_finding "Weak Cryptography" "HIGH" "Weak cryptographic algorithm" "$file" "Use stronger algorithms like SHA-256 or SHA-3"
        fi
    done

    # Check for hardcoded cryptographic keys
    grep -r "\b(0x[a-fA-F0-9]{64})" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line" | cut -d -f1)
        if grep -v "//.*mock\|//.*test" "$file" > /dev/null; then
            record_finding "Weak Cryptography" "HIGH" "Hardcoded cryptographic key" "$file" "Use secure key management"
        fi
    done

    # Check for predictable randomness
    grep -r "time\.Now\|time\.Unix\|math/rand" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cutd: -f1)
        if grep -v "//.*seed\|//.*crypto\|//\.crypt" "$file" > /dev/null; then
            record_finding "Weak Cryptography" "MEDIUM" "Predictable randomness" "$file" "Use cryptographically secure random number generators"
        fi
    done
    echo ""
}

# Function to check authentication and authorization
check_authentication() {
    echo -e "${YELLOW}🔐 Checking Authentication and Authorization...${NC}"

    # Check for session management issues
    grep -r "SessionID.*=.*" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cut -d -f1)
        if grep -v "//.*secure\|//.*https" "$file" > /dev/null; then
            record_finding "Authentication" "MEDIUM" "Insecure session management" "$file" "Use secure session management with HTTPS"
        fi
    done

    # Check for JWT security issues
    grep -r "jwt\.Parse.*\|jwt\.ParseWithClaims" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cut -d -f1)
        if ! grep -q "jwt\.VerifySigningString\|jwt\.ParseWithClaims.*error" "$file" > /dev/null; then
            record_finding "Authentication" "MEDIUM" "JWT without verification" "$file" "Always verify JWT signatures"
        fi
    done

    # Check for authorization bypass
    grep -r "//.*TODO.*auth\|//FIXME.*auth" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut -d -f1)
        record_finding "Authentication" "LOW" "Authentication TODO/FIXME" "$file" "Complete authentication implementation"
    done
    echo ""
}

# Function to check input validation
check_input_validation() {
    echo -e "${YELLOW}✅ Checking Input Validation...${NC}"

    # Check for missing input validation
    grep -r "c\.BindJSON\|c\.Bind\|c\.ShouldBind" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cut -d -f1)
        if ! grep -q "validate\|validator\|Validate" "$file" > /dev/null; then
            record_finding "Input Validation" "HIGH" "Missing input validation" "$file" "Add validation middleware or validator functions"
        fi
    done

    # Check for SQL injection prevention
    grep -r "c\.Query\|db\.Query\|db\.Raw" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        if ! grep -q "parameter\|parametrized\|prepared\|safe" "$file" > /dev/null; then
            record_finding "Input Validation" "HIGH" "Direct database query without validation" "$file" "Use ORM with parameter binding"
        fi
    done

    # Check for file upload security
    grep -r "os\.Create\|multipart\.FormFile" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cut -d -f1)
        if ! grep -q "os\.RemoveAll\|path\.Clean\|file\.Ext\|content-type" "$file" > /dev/null; then
            record_finding "Input Validation" "MEDIUM" "File upload without security checks" "$file" "Add file type validation and size limits"
        fi
    done
    echo ""
}

# Function to check API security
check_api_security() {
    echo -e "${YELLOW}🌐 Checking API Security...${NC}"

    # Check for CORS configuration
    grep -r "Access-Control" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; do
        file=$(echo "$line | cut -d: -f1)
        if grep -q "Access-Control-Allow-Origin.*\*" "$file" > /dev/null && ! grep -q "Access-Control-Allow-Credentials" "$file" > /dev/null; then
            record_finding "API Security" "MEDIUM" "Permissive CORS policy" "$file" "Be more specific with allowed origins"
        fi
    done

    # Check for security headers
    grep -r "Content-Security-Policy\|X-Frame-Options\|X-Content-Type-Options" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        if grep -v "security" "$file" > /dev/null; then
            record_finding "API Security" "MEDIUM" "Missing security headers" "$file" "Add security headers for enhanced protection"
        fi
    done

    # Check for rate limiting
    grep -r "rate.*limit\|throttle" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        record_finding "API Security" "LOW" "Rate limiting implemented" "$file" "Ensure rate limiting is properly configured"
    done
    echo ""
}

# Function to check logging security
check_logging_security() {
    echo -e "${YELLOW}📝 Checking Logging Security...${NC}"

    # Check for sensitive data in logs
    grep -r "password\|secret\|token\|key" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        if grep -v "//.*masked\|//.*sanitized\|//.*filtered" "$file" > /dev/null; then
            record_finding "Logging Security" "MEDIUM" "Sensitive data in logs" "$file" "Mask sensitive information in logs"
        fi
    done

    # Check for debug information in production
    grep -r "DEBUG\|TRACE\|println\|fmt\.Print" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut -d -f1)
        record_finding "Logging Security" "LOW" "Debug logging present" "$file" "Ensure debug logging is disabled in production"
    done

    # Check for structured logging
    grep -r "log\|logger" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        if ! grep -v "zap\|structured\|json" "$file" > /dev/null; then
            record_finding "Logging Security" "LOW" "Unstructured logging" "$file" "Use structured logging like Zap"
        fi
    done
    echo ""
}

# Function to check Docker security
check_docker_security() {
    echo -e "${YELLOW}🐳 Checking Docker Security...${NC}"

    if [ -f "Dockerfile" ]; then
        # Check for running as root
        if grep -q "USER root" Dockerfile > /dev/null; then
            record_finding "Docker Security" "HIGH" "Container running as root" "Use non-root user"
        fi

        # Check for sensitive data in image
        if grep -q "ADD .env\|COPY.*\.env" Dockerfile > /dev/null; then
            record_finding "Docker Security" "HIGH" "Sensitive files in image" "$file" "Use Docker secrets or build-time injection"
        fi

        # Check for health check
        if ! grep -q "HEALTHCHECK\|HEALTH.*CHECK" Dockerfile > /dev/null; then
            record_finding "Docker Security" "LOW" "Missing health check" "Add health check endpoint"
        fi
    fi

    # Check docker-compose files
    if [ -f "docker-compose.yml" ]; then
        # Check for exposed ports
        if grep -q "ports:.*-\s\"8080:8080\"" docker-compose.yml > /dev/null; then
            record_finding "Docker Security" "LOW" "Port exposed to public" "Use internal networking"
        fi

        # Check for default credentials
        if grep -q "MYSQL_ROOT_PASSWORD:\|POSTGRES_PASSWORD:" docker-compose.yml > /dev/null; then
            record_finding "Docker Security" "HIGH" "Default credentials in compose file" "Use secrets management"
        fi
    fi
    echo ""
}

# Function to check for race conditions
check_race_conditions() {
    echo -e "${YELLOW}🏃 Checking for Race Conditions...${NC}"

    # Check for shared mutable state without synchronization
    grep -r "var.*map\[.*\].*map" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut -d -f1)
        record_finding "Race Conditions" "HIGH" "Unprotected shared state" "$file" "Use proper synchronization primitives"
    done

    # Check for goroutine synchronization issues
    grep -r "go func.*()" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut -d -f1)
        if grep -v "sync\|mutex\|channel\|WaitGroup\|atomic" "$file" > /dev/null; then
            record_finding "Race Conditions" "MEDIUM" "Goroutine without synchronization" "$file" "Add proper synchronization"
        fi
    done

    # Check for concurrent database access
    grep -r "db\.Exec\|db\.Query\|db\.Begin" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        if ! grep -v "BEGIN\|COMMIT\|ROLLBACK" "$file" > /dev/null && grep -v "//.*tx\|//.*transaction" "$file" > /dev/null; then
            record_finding "Race Conditions" "HIGH" "Database access without transaction" "$file" "Use proper transaction management"
        fi
    done
    echo ""
}

# Function to check for dependency injection security
check_dependency_injection() {
    echo -e "${YELLOW}🔌 Checking Dependency Injection Security...${NC}"

    # Check for unsafe type assertions
    grep -r "\.Interface{}" --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        record_finding "Dependency Injection" "LOW" "Interface assertion without safety check" "$file" "Add type safety checks"
    done

    # Check for reflection usage in untrusted contexts
    grep -r "reflect\." --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        record_finding "Dependency Injection" "MEDIUM" "Reflection in untrusted context" "$file" "Validate input before reflection"
    done

    # Check for mock usage patterns
    grep -r "mock\." --include="*.go" --exclude-dir=./vendor 2>/dev/null | while read line; line
        file=$(echo "$line | cut-d: -f1)
        record_finding "Dependency Injection" "LOW" "Mock object injection" "$file" "Ensure mocks are properly implemented"
    done
    echo ""
}

# Function to generate security recommendations
generate_recommendations() {
    echo -e "${BLUE}💡 Generating Security Recommendations...${NC}"

    cat >> "$report_file" << 'EOF'
## Security Recommendations

### High Priority
1. **Implement Comprehensive Input Validation**
   - Add validation middleware for all endpoints
   - Use parameterized queries to prevent SQL injection
   - Implement file upload security checks

2. **Secure Authentication and Authorization**
   - Implement proper JWT validation and revocation
   - Add rate limiting to prevent brute force attacks
   - Implement proper session management

3. **Encrypt Sensitive Data**
   - Use environment variables for secrets
   - Implement database encryption at rest
   - Secure file storage and access

### Medium Priority
1. **Enhance Logging Security**
   - Implement structured logging with proper levels
   - Mask sensitive information in logs
   - Add security event logging

2. **Improve API Security**
   - Implement comprehensive CORS policies
   - Add security headers
   - Enhance rate limiting strategies

3. **Container Security**
   - Use non-root container execution
   - Implement secure base images
   - Add health checks and monitoring

### Low Priority
1. **Performance Optimization**
   - Implement caching strategies
   - Add database connection pooling
   - Optimize database queries

2. **Monitoring and Alerting**
   - Implement security monitoring
   - Add automated security scanning
   - Set up alerting for security events

### Automated Testing
- Integrate security testing in CI/CD pipeline
- Run automated vulnerability scans
- Monitor for new security issues

---

EOF

    echo ""
    echo -e "${GREEN}✅ Security audit completed!${NC}"
    echo "Full report available at: $report_file"
}

# Main execution
main() {
    # Run all security checks
    check_file_permissions
    check_hardcoded_secrets
    check_sql_injection
    check_xss
    check_dependencies
    check_cryptography
    check_authentication
    check_input_validation
    check_api_security
    check_logging_security
    check_docker_security
    check_race_conditions
    check_dependency_injection
    generate_recommendations

    # Generate summary statistics
    total_findings=$(grep -c "## " "$report_file" | wc -l)
    high_findings=$(grep -c "🔴" "$report_file" | wc -l)
    medium_findings=$(grep -c "🟡" "$report_file" | wc -l)
    low_findings=$(grep -c "🟢" "$report_file" | wc -l)

    echo ""
    echo -e "${BLUE}📊 Security Audit Summary${NC}"
    echo "=================================================="
    echo "Total Security Findings: $total_findings"
    echo "High Risk Issues: $high_findings"
    echo "Medium Risk Issues: $medium_findings"
    echo "Low Risk Issues: $low_findings"
    echo ""
    echo "Security report generated: $report_file"

    # Exit with appropriate code based on findings
    if [ $high_findings -gt 0 ]; then
        exit 1
    elif [ $medium_findings -gt 10 ]; then
        exit 1
    else
        exit 0
    fi
}

# Execute main function
main "$@"