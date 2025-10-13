#!/bin/bash

# Go-Next-ERP 测试运行脚本
# 用于快速运行项目的各种测试

set -e

COLOR_GREEN='\033[0;32m'
COLOR_YELLOW='\033[1;33m'
COLOR_RED='\033[0;31m'
COLOR_BLUE='\033[0;34m'
COLOR_RESET='\033[0m'

function print_header() {
    echo -e "${COLOR_BLUE}========================================${COLOR_RESET}"
    echo -e "${COLOR_BLUE}$1${COLOR_RESET}"
    echo -e "${COLOR_BLUE}========================================${COLOR_RESET}"
}

function print_success() {
    echo -e "${COLOR_GREEN}✅ $1${COLOR_RESET}"
}

function print_warning() {
    echo -e "${COLOR_YELLOW}⚠️  $1${COLOR_RESET}"
}

function print_error() {
    echo -e "${COLOR_RED}❌ $1${COLOR_RESET}"
}

function run_adapter_tests() {
    print_header "Running Adapter Tests"
    if go test -v ./internal/adapter/... -cover; then
        print_success "Adapter tests passed"
    else
        print_error "Adapter tests failed"
        return 1
    fi
}

function run_auth_tests() {
    print_header "Running Auth Module Tests"

    echo -e "\n${COLOR_YELLOW}Auth Model Tests:${COLOR_RESET}"
    go test -v ./internal/auth/model/... -cover

    echo -e "\n${COLOR_YELLOW}Auth Repository Tests:${COLOR_RESET}"
    go test -v ./internal/auth/repository/... -cover

    echo -e "\n${COLOR_YELLOW}Auth Authentication Tests:${COLOR_RESET}"
    go test -v ./internal/auth/authentication/... -cover

    echo -e "\n${COLOR_YELLOW}Auth Authorization Tests:${COLOR_RESET}"
    go test -v ./internal/auth/authorization/... -cover

    print_success "Auth module tests completed"
}

function run_handler_tests() {
    print_header "Running Handler Tests"

    echo -e "\n${COLOR_YELLOW}Organization Handler Tests:${COLOR_RESET}"
    if [ -d "./internal/organization/handler" ]; then
        go test -v ./internal/organization/handler/... -cover || true
    else
        print_warning "Organization handler not found"
    fi

    echo -e "\n${COLOR_YELLOW}Approval Handler Tests:${COLOR_RESET}"
    if [ -d "./internal/approval/handler" ]; then
        go test -v ./internal/approval/handler/... -cover || true
    else
        print_warning "Approval handler not found"
    fi

    echo -e "\n${COLOR_YELLOW}Form Handler Tests:${COLOR_RESET}"
    if [ -d "./internal/form/handler" ]; then
        go test -v ./internal/form/handler/... -cover || true
    else
        print_warning "Form handler not found"
    fi

    echo -e "\n${COLOR_YELLOW}Notification Handler Tests:${COLOR_RESET}"
    if [ -d "./internal/notification/handler" ]; then
        go test -v ./internal/notification/handler/... -cover || true
    else
        print_warning "Notification handler not found"
    fi

    print_success "Handler tests completed"
}

function run_workflow_tests() {
    print_header "Running Workflow Tests"
    if [ -d "./pkg/workflow" ]; then
        go test -v ./pkg/workflow/... -cover
        print_success "Workflow tests passed"
    else
        print_warning "Workflow package not found"
    fi
}

function generate_coverage_report() {
    print_header "Generating Coverage Report"

    # Generate coverage for all packages
    go test ./... -coverprofile=/tmp/go-next-erp-coverage.out 2>&1 || true

    # Show overall coverage
    echo -e "\n${COLOR_YELLOW}Overall Coverage:${COLOR_RESET}"
    go tool cover -func=/tmp/go-next-erp-coverage.out | grep total || true

    # Generate HTML report
    go tool cover -html=/tmp/go-next-erp-coverage.out -o /tmp/go-next-erp-coverage.html
    print_success "Coverage report generated: /tmp/go-next-erp-coverage.html"
}

function run_all_tests() {
    print_header "Running All Tests"

    run_adapter_tests
    echo ""

    run_auth_tests
    echo ""

    run_handler_tests
    echo ""

    run_workflow_tests
    echo ""

    generate_coverage_report

    print_header "Test Summary"
    print_success "All tests completed!"
}

# Main script
case "${1:-all}" in
    adapter)
        run_adapter_tests
        ;;
    auth)
        run_auth_tests
        ;;
    handler)
        run_handler_tests
        ;;
    workflow)
        run_workflow_tests
        ;;
    coverage)
        generate_coverage_report
        ;;
    all)
        run_all_tests
        ;;
    *)
        echo "Usage: $0 {adapter|auth|handler|workflow|coverage|all}"
        echo ""
        echo "Commands:"
        echo "  adapter   - Run Adapter layer tests"
        echo "  auth      - Run Auth module tests"
        echo "  handler   - Run Handler layer tests"
        echo "  workflow  - Run Workflow tests"
        echo "  coverage  - Generate coverage report"
        echo "  all       - Run all tests (default)"
        exit 1
        ;;
esac
