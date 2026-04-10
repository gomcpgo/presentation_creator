#!/bin/bash
set -e

# Presentation Creator MCP Server Build/Test Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

command=$1
shift || true

case "$command" in
    build)
        echo "Building presentation_creator..."
        mkdir -p bin
        go build -o bin/presentation_creator cmd/main.go
        echo "Build complete: bin/presentation_creator"
        ;;

    test)
        echo "Running tests..."
        go test ./... -v
        ;;

    install)
        echo "Installing dependencies..."
        go mod download
        go mod tidy
        ;;

    create)
        if [ -z "$1" ] || [ -z "$2" ]; then
            echo "Usage: ./run.sh create <name> <slides_json>"
            echo "  slides_json: JSON array of HTML strings"
            echo "  Example: ./run.sh create 'My Deck' '[\"<h1>Slide 1</h1>\",\"<h1>Slide 2</h1>\"]'"
            exit 1
        fi
        bin/presentation_creator -create "$1" -slides "$2"
        ;;

    list)
        bin/presentation_creator -list
        ;;

    get)
        if [ -z "$1" ]; then
            echo "Usage: ./run.sh get <presentation_id>"
            exit 1
        fi
        bin/presentation_creator -get "$1"
        ;;

    update-slide)
        if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
            echo "Usage: ./run.sh update-slide <presentation_id> <slide_number> <html_content>"
            exit 1
        fi
        bin/presentation_creator -update "$1" -slide "$2" -html "$3"
        ;;

    add-slide)
        if [ -z "$1" ] || [ -z "$2" ]; then
            echo "Usage: ./run.sh add-slide <presentation_id> <html_content> [position]"
            exit 1
        fi
        position_flag=""
        if [ -n "$3" ]; then
            position_flag="-position $3"
        fi
        bin/presentation_creator -add-slide "$1" -html "$2" $position_flag
        ;;

    delete-slide)
        if [ -z "$1" ] || [ -z "$2" ]; then
            echo "Usage: ./run.sh delete-slide <presentation_id> <slide_number>"
            exit 1
        fi
        bin/presentation_creator -delete-slide "$1" -slide "$2"
        ;;

    export)
        if [ -z "$1" ] || [ -z "$2" ]; then
            echo "Usage: ./run.sh export <presentation_id> <output_path> [format]"
            echo "  format: pptx (default) or pdf"
            exit 1
        fi
        format="${3:-pptx}"
        bin/presentation_creator -export "$1" -output "$2" -format "$format"
        ;;

    add-media)
        if [ -z "$1" ] || [ -z "$2" ]; then
            echo "Usage: ./run.sh add-media <presentation_id> <media_path>"
            exit 1
        fi
        bin/presentation_creator -add-media "$1" -media-path "$2"
        ;;

    clean)
        echo "Cleaning build artifacts..."
        rm -rf bin
        echo "Clean complete"
        ;;

    *)
        echo "Presentation Creator MCP Server"
        echo ""
        echo "Usage: ./run.sh <command> [args]"
        echo ""
        echo "Commands:"
        echo "  build                                        Build the MCP server"
        echo "  test                                         Run tests"
        echo "  install                                      Install dependencies"
        echo "  create <name> <slides_json>                  Create a new presentation"
        echo "  list                                         List all presentations"
        echo "  get <id>                                     Get presentation by ID"
        echo "  update-slide <id> <slide_num> <html>         Update a slide"
        echo "  add-slide <id> <html> [position]             Add a new slide"
        echo "  delete-slide <id> <slide_num>                Delete a slide"
        echo "  export <id> <output_path> [format]           Export (pptx/pdf)"
        echo "  add-media <id> <path>                        Add media file"
        echo "  clean                                        Remove build artifacts"
        echo ""
        echo "Examples:"
        echo "  ./run.sh build"
        echo "  ./run.sh create 'Q4 Review' '[\"<h1>Q4 Review</h1>\",\"<h1>Revenue</h1>\"]'"
        echo "  ./run.sh list"
        echo "  ./run.sh update-slide q4-review-a3f9 1 '<h1>Updated Title</h1>'"
        echo "  ./run.sh export q4-review-a3f9 ~/Desktop/q4-review.pptx"
        ;;
esac
