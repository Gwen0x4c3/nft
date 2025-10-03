#!/bin/bash

# Comprehensive Test Execution Script
# This script runs all test scenarios and generates a comprehensive report

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 Starting Comprehensive Test Execution${NC}"
echo "=================================================="

# Create results directory
mkdir -p test-results
timestamp=$(date +%Y%m%d_%H%M%S)
report_file="test-results/test_report_${timestamp}.md"

# Initialize test report
cat > "$report_file" << EOF
# Comprehensive Test Report

**Date**: $(date '+%Y-%m-%d %H:%M:%S')
**Environment**: $(go version)
**Test Suite**: NFT Platform Integration Tests

## Test Results Summary

EOF

# Function to record test result
record_test_result() {
    local test_name="$1"
    local status="$2"
    local duration="$3"
    local details="$4"

    local status_icon="❌"
    if [ "$status" = "PASS" ]; then
        status_icon="✅"
    fi

    echo "Recording: $test_name - $status ($duration)"

    echo "- **$test_name**: $status_icon $status ($duration)" >> "$report_file"
    if [ -n "$details" ]; then
        echo "  - $details" >> "$report_file"
    fi
    echo "" >> "$report_file"
}

# Test 1: Unit Tests
echo -e "${YELLOW}📋 Running Unit Tests...${NC}"
unit_start=$(date +%s)
make test > test-results/unit_tests.log 2>&1
unit_exit_code=$?
unit_end=$(date +%s)
unit_duration=$((unit_end - unit_start))

if [ $unit_exit_code -eq 0 ]; then
    record_test_result "Unit Tests" "PASS" "${unit_duration}s" "All unit tests passed"
    echo -e "${GREEN}✅ Unit Tests passed (${unit_duration}s)${NC}"
else
    record_test_result "Unit Tests" "FAIL" "${unit_duration}s" "Unit tests failed - check test-results/unit_tests.log"
    echo -e "${RED}❌ Unit Tests failed (${unit_duration}s)${NC}"
fi

# Test 2: Contract Tests
echo -e "${YELLOW}📄 Running Contract Tests...${NC}"
contract_start=$(date +%s)
go test -v ./tests/contract/... > test-results/contract_tests.log 2>&1
contract_exit_code=$?
contract_end=$(date +%s)
contract_duration=$((contract_end - contract_start))

if [ $contract_exit_code -eq 0 ]; then
    record_test_result "Contract Tests" "PASS" "${contract_duration}s" "All contract tests passed"
    echo -e "${GREEN}✅ Contract Tests passed (${contract_duration}s)${NC}"
else
    record_test_result "Contract Tests" "FAIL" "${contract_duration}s" "Contract tests failed - check test-results/contract_tests.log"
    echo -e "${RED}❌ Contract Tests failed (${contract_duration}s)${NC}"
fi

# Test 3: Performance Tests
echo -e "${YELLOW}⚡ Running Performance Tests...${NC}"
perf_start=$(date +%s)
go test -v ./tests/performance/... > test-results/performance_tests.log 2>&1
perf_exit_code=$?
perf_end=$(date +%s)
perf_duration=$((perf_end - perf_start))

if [ $perf_exit_code -eq 0 ]; then
    record_test_result "Performance Tests" "PASS" "${perf_duration}s" "Performance tests passed"
    echo -e "${GREEN}✅ Performance Tests passed (${perf_duration}s)${NC}"
else
    record_test_result "Performance Tests" "FAIL" "${perf_duration}s" "Performance tests failed - check test-results/performance_tests.log"
    echo -e "${RED}❌ Performance Tests failed (${perf_duration}s)${NC}"
fi

# Test 4: Integration Tests
echo -e "${YELLOW}🔗 Running Integration Tests...${NC}"
integration_start=$(date +%s)
go test -v ./tests/integration/... -timeout=30m > test-results/integration_tests.log 2>&1
integration_exit_code=$?
integration_end=$(date +%s)
integration_duration=$((integration_end - integration_start))

if [ $integration_exit_code -eq 0 ]; then
    record_test_result "Integration Tests" "PASS" "${integration_duration}s" "All integration tests passed"
    echo -e "${GREEN}✅ Integration Tests passed (${integration_duration}s)${NC}"
else
    record_test_result "Integration Tests" "FAIL" "${integration_duration}s" "Integration tests failed - check test-results/integration_tests.log"
    echo -e "${RED}❌ Integration Tests failed (${integration_duration}s)${NC}"
fi

# Test 5: Comprehensive Scenarios
echo -e "${YELLOW}🎯 Running Comprehensive Integration Scenarios...${NC}"
comprehensive_start=$(date +%s)
go test -v ./tests/integration/comprehensive_scenarios_test.go -timeout=30m > test-results/comprehensive_tests.log 2>&1
comprehensive_exit_code=$?
comprehensive_end=$(date +%s)
comprehensive_duration=$((comprehensive_end - comprehensive_start))

if [ $comprehensive_exit_code -eq 0 ]; then
    record_test_result "Comprehensive Scenarios" "PASS" "${comprehensive_duration}s" "All comprehensive scenarios passed"
    echo -e "${GREEN}✅ Comprehensive Scenarios passed (${comprehensive_duration}s)${NC}"
else
    record_test_result "Comprehensive Scenarios" "FAIL" "${comprehensive_duration}s" "Comprehensive scenarios failed - check test-results/comprehensive_tests.log"
    echo -e "${RED}❌ Comprehensive Scenarios failed (${comprehensive_duration}s)${NC}"
fi

# Calculate overall statistics
total_tests=$((unit_tests_passed + contract_tests_passed + performance_tests_passed + integration_tests_passed + comprehensive_tests_passed))
total_failed=$((unit_tests_failed + contract_tests_failed + performance_tests_failed + integration_tests_failed + comprehensive_tests_failed))
total_duration=$((unit_duration + contract_duration + perf_duration + integration_duration + comprehensive_duration))

# Generate final summary
cat >> "$report_file" << EOF
## Overall Statistics

- **Total Test Duration**: ${total_duration}s
- **Total Tests Executed**: $((total_tests_passed + total_failed))
- **Tests Passed**: $total_tests_passed
- **Tests Failed**: $total_failed
- **Success Rate**: $(echo "scale=2; $total_tests_passed * 100 / ($total_tests_passed + $total_failed)" | bc)%

## Test Artifacts

All test logs and artifacts are available in the \`test-results/\` directory:
- \`unit_tests.log\` - Unit test execution log
- \`contract_tests.log\` - Contract test execution log
- \`performance_tests.log\ - Performance test execution log
- \`integration_tests.log\ - Integration test execution log
- \`comprehensive_tests.log\` - Comprehensive scenarios test execution log

## Recommendations

EOF

if [ $total_failed -eq 0 ]; then
    cat >> "$report_file" << EOF
🎉 **All tests passed!** The system is ready for production deployment.

Next Steps:
1. Run security audit and vulnerability testing
2. Perform load testing in staging environment
3. Review test coverage reports
4. Deploy to staging environment for final validation

EOF
    echo -e "${GREEN}🎉 All tests passed! System is ready for production.${NC}"
else
    cat >> "$report_file" << EOF
⚠️ **Some tests failed.** Please review the failed tests and fix issues before proceeding.

Next Steps:
1. Review failed test logs in test-results/
2. Fix identified issues
3. Re-run failed tests
4. Address any security vulnerabilities found
5. Re-run comprehensive testing suite

EOF
    echo -e "${RED}⚠️ Some tests failed. Please review the issues before proceeding.${NC}"
fi

cat >> "$report_file" << EOF
---
**Report Generated**: $(date '+%Y-%m-%d %H:%M:%S')
**Test Report File**: $report_file
EOF

# Display summary
echo ""
echo "=================================================="
echo -e "${BLUE}📊 Test Execution Summary${NC}"
echo "=================================================="
echo "Total Duration: ${total_duration}s"
echo "Tests Passed: $total_tests_passed"
echo "Tests Failed: $total_failed"
echo "Success Rate: $(echo "scale=2; $total_tests_passed * 100 / ($total_tests_passed + total_failed)" | bc)%"
echo ""

if [ $total_failed -eq 0 ]; then
    echo -e "${GREEN}🎯 All test suites passed successfully!${NC}"
else
    echo -e "${RED}❌ Some test suites failed. Check test-results/ directory for details.${NC}"
fi

echo "Full report available at: $report_file"
echo "Test logs available in: test-results/"

# Exit with appropriate code
if [ $total_failed -eq 0 ]; then
    exit 0
else
    exit 1
fi