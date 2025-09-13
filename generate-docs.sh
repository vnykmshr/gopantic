#!/bin/bash

# Generate comprehensive Go documentation for gopantic
# This script helps developers access detailed API documentation

echo "🚀 gopantic - Go Documentation Generator"
echo "========================================="

echo ""
echo "📖 Package Overview:"
echo "-------------------"
go doc ./pkg/model

echo ""
echo "🔧 Core Functions:"
echo "-----------------"
echo "ParseInto - Main parsing function:"
go doc ./pkg/model ParseInto

echo ""
echo "ParseIntoWithFormat - Format-specific parsing:"
go doc ./pkg/model ParseIntoWithFormat

echo ""
echo "🚀 Caching Functions:"
echo "--------------------"
echo "NewCachedParser - Create cached parser:"
go doc ./pkg/model NewCachedParser

echo ""
echo "📝 Validation Types:"
echo "-------------------"
echo "ValidationError - Validation error details:"
go doc ./pkg/model ValidationError

echo ""
echo "🔍 Format Detection:"
echo "-------------------"
echo "DetectFormat - Automatic format detection:"
go doc ./pkg/model DetectFormat

echo ""
echo "💡 To view complete documentation:"
echo "  go doc -all ./pkg/model              # Full package docs"
echo "  go doc ./pkg/model <FunctionName>    # Specific function"
echo "  go doc -http :6060                   # Start HTML docs server"
echo ""
echo "📚 For usage examples, see:"
echo "  examples/          # Practical examples"
echo "  docs/api.md        # Complete API reference"
echo "  docs/architecture.md  # Implementation details"
echo ""

# Optional: Start HTTP documentation server
read -p "🌐 Start HTML documentation server on http://localhost:6060? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Starting documentation server..."
    echo "Visit: http://localhost:6060/pkg/github.com/vnykmshr/gopantic/pkg/model/"
    go doc -http :6060
fi