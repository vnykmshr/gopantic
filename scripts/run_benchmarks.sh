#!/bin/bash

# Comprehensive benchmark runner for gopantic performance analysis
# This script runs various benchmark scenarios and generates comparison reports

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
RESULTS_DIR="${PROJECT_ROOT}/benchmark_results"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================================${NC}"
}

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create results directory
mkdir -p "${RESULTS_DIR}"
cd "${PROJECT_ROOT}"

# Get system info
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
HOSTNAME=$(hostname)
GO_VERSION=$(go version | cut -d' ' -f3)
OS_INFO=$(uname -srm)

REPORT_FILE="${RESULTS_DIR}/benchmark_report_${TIMESTAMP}.txt"
COMPARISON_FILE="${RESULTS_DIR}/comparison_${TIMESTAMP}.txt"

print_header "Starting Gopantic Performance Benchmarks"

print_info "System Information:"
echo "Timestamp: ${TIMESTAMP}" | tee "${REPORT_FILE}"
echo "Hostname: ${HOSTNAME}" | tee -a "${REPORT_FILE}"
echo "Go Version: ${GO_VERSION}" | tee -a "${REPORT_FILE}"
echo "OS: ${OS_INFO}" | tee -a "${REPORT_FILE}"
echo "CPU Cores: $(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 'Unknown')" | tee -a "${REPORT_FILE}"
echo "" | tee -a "${REPORT_FILE}"

# Run basic parsing benchmarks
print_header "Basic Parsing Benchmarks"
print_info "Running simple vs complex parsing comparisons..."

echo "=== Basic Parsing Benchmarks ===" | tee -a "${REPORT_FILE}"
go test -bench="Benchmark(StdJSON|Gopantic)_(Simple|Complex)User$" \
    -benchmem -count=3 -benchtime=5s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run cache performance benchmarks  
print_header "Cache Performance Benchmarks"
print_info "Running cached vs non-cached parsing comparisons..."

echo "=== Cache Performance Benchmarks ===" | tee -a "${REPORT_FILE}"
go test -bench="Benchmark.*_Cached$|BenchmarkCache_" \
    -benchmem -count=3 -benchtime=5s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run memory allocation benchmarks
print_header "Memory Allocation Analysis"
print_info "Running memory allocation comparison..."

echo "=== Memory Allocation Analysis ===" | tee -a "${REPORT_FILE}"
go test -bench="BenchmarkMemory" \
    -benchmem -count=3 -benchtime=5s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run real-world scenario benchmarks
print_header "Real-World Scenario Benchmarks"
print_info "Running user profile and configuration parsing benchmarks..."

echo "=== Real-World Scenarios ===" | tee -a "${REPORT_FILE}"
go test -bench="BenchmarkRealWorld|BenchmarkConfig" \
    -benchmem -count=3 -benchtime=5s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run large batch processing benchmarks
print_header "Large Batch Processing Benchmarks"
print_info "Running large array/batch processing benchmarks..."

echo "=== Large Batch Processing ===" | tee -a "${REPORT_FILE}"
go test -bench="BenchmarkLarge" \
    -benchmem -count=3 -benchtime=3s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run concurrent processing benchmarks
print_header "Concurrent Processing Benchmarks"  
print_info "Running concurrent/parallel processing benchmarks..."

echo "=== Concurrent Processing ===" | tee -a "${REPORT_FILE}"
go test -bench="BenchmarkConcurrent" \
    -benchmem -count=3 -benchtime=5s -cpu=1,2,4,8 \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run validation-specific benchmarks
print_header "Validation Performance Benchmarks"
print_info "Running validation and error handling benchmarks..."

echo "=== Validation Performance ===" | tee -a "${REPORT_FILE}"
go test -bench="BenchmarkValidation|BenchmarkErrorHandling|BenchmarkCrossField" \
    -benchmem -count=3 -benchtime=5s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Run type coercion benchmarks
print_header "Type Coercion Benchmarks"
print_info "Running type coercion performance tests..."

echo "=== Type Coercion Performance ===" | tee -a "${REPORT_FILE}"
go test -bench="BenchmarkTypeCoercion" \
    -benchmem -count=3 -benchtime=5s \
    ./benchmarks 2>&1 | tee -a "${REPORT_FILE}"

echo "" | tee -a "${REPORT_FILE}"

# Generate comparison summary
print_header "Generating Performance Comparison Report"

cat > "${COMPARISON_FILE}" << EOF
# Gopantic vs Standard Library JSON - Performance Comparison Report
Generated: $(date)
System: ${OS_INFO}
Go Version: ${GO_VERSION}

## Summary

This report compares the performance characteristics of gopantic against the standard library JSON package.

### Key Metrics Analyzed:
1. **Parsing Speed (ns/op)**: Time taken per operation
2. **Memory Allocations (allocs/op)**: Number of memory allocations
3. **Memory Usage (B/op)**: Bytes allocated per operation
4. **Cache Hit Performance**: Speedup from caching
5. **Validation Overhead**: Cost of validation features
6. **Concurrent Performance**: Scalability under load

### Expected Performance Characteristics:

**Gopantic Advantages:**
- Comprehensive validation built-in (no manual validation code needed)
- Type coercion reduces parsing errors
- Caching provides significant speedup for repeated parsing
- YAML support without external dependencies
- Structured error reporting

**Standard JSON Advantages:**
- Lower baseline parsing overhead (no validation)
- Fewer memory allocations for simple cases
- Part of standard library (no external dependencies)

### Raw Benchmark Results:

EOF

# Extract key comparisons from the full report
grep -E "Benchmark(StdJSON|Gopantic).*-" "${REPORT_FILE}" | \
    sort | tee -a "${COMPARISON_FILE}"

print_header "Benchmark Results Summary"

print_info "Full benchmark results: ${REPORT_FILE}"
print_info "Comparison summary: ${COMPARISON_FILE}"

# Calculate and display some quick stats
GOPANTIC_SIMPLE=$(grep "BenchmarkGopantic_SimpleUser-" "${REPORT_FILE}" | head -1 | awk '{print $3}' | sed 's/ns\/op//')
STDJSON_SIMPLE=$(grep "BenchmarkStdJSON_SimpleUser-" "${REPORT_FILE}" | head -1 | awk '{print $3}' | sed 's/ns\/op//')

if [[ -n "${GOPANTIC_SIMPLE}" && -n "${STDJSON_SIMPLE}" ]]; then
    RATIO=$(echo "scale=2; ${GOPANTIC_SIMPLE} / ${STDJSON_SIMPLE}" | bc -l 2>/dev/null || echo "N/A")
    echo ""
    print_info "Quick Comparison (Simple User Parsing):"
    echo "  Standard JSON: ${STDJSON_SIMPLE} ns/op"
    echo "  Gopantic:      ${GOPANTIC_SIMPLE} ns/op"
    echo "  Ratio:         ${RATIO}x (lower is better)"
fi

print_header "Benchmark Analysis Complete"

print_info "Next steps:"
echo "  1. Review detailed results in: ${REPORT_FILE}"
echo "  2. Check comparison summary: ${COMPARISON_FILE}"  
echo "  3. Run specific benchmarks: go test -bench=BenchmarkName ./benchmarks"
echo "  4. Profile memory: go test -bench=. -memprofile=mem.prof ./benchmarks"
echo "  5. Profile CPU: go test -bench=. -cpuprofile=cpu.prof ./benchmarks"

# Check for any concerning results
if grep -q "FAIL" "${REPORT_FILE}"; then
    print_warning "Some benchmarks failed - check the results file for details"
    exit 1
fi

print_info "All benchmarks completed successfully!"