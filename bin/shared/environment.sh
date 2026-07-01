#!/bin/bash

# Centralized environment configuration
export ENVIRONMENT=${ENVIRONMENT:-"development"}
export EVENT_RELAY_DATABASE_PATH=${EVENT_RELAY_DATABASE_PATH:-"../mdk-event-relay/events.db"}
export SERVER_ADDRESS=${SERVER_ADDRESS:-":8080"}

function show_config() {
    echo "─────────────────────────────────────"
    echo "⚙️  Configuration Profile:"
    echo "   🔹 ENVIRONMENT: $ENVIRONMENT"
    echo "   🔹 SERVER_ADDRESS: $SERVER_ADDRESS"
    echo "   🔹 EVENT_RELAY_DATABASE_PATH: $EVENT_RELAY_DATABASE_PATH"
    echo "─────────────────────────────────────"
}

function show_banner() {
    echo "============================================="
    echo " __  __  ____  _  __"
    echo "|  \/  |  _ \| |/ /"
    echo "| \  / | | | | ' / "
    echo "| |\/| | | | |  <  "
    echo "| |  | | |_| | . \ "
    echo "|_|  |_|____/|_|\\_\\"
    echo ""
    echo "Creator: Marco Antonio - markitos"
    echo "============================================="
    echo " > (mArKit0sDevSecOpsKit)"
    echo " > Markitos DevSecOps Kulture"
    echo ""
}