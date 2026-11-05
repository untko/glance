#!/bin/bash
PLIST="$HOME/Library/LaunchAgents/com.glance.dock.plist"
LOG="$HOME/Projects/glance/glance.log"

case "$1" in
  start|enable)
    launchctl load "$PLIST" 2>/dev/null || launchctl bootstrap "gui/$(id -u)" "$PLIST"
    echo "Glance service started."
    ;;
  stop|disable)
    launchctl unload "$PLIST" 2>/dev/null || launchctl bootout "gui/$(id -u)/com.glance.dock" 2>/dev/null
    echo "Glance service stopped."
    ;;
  restart)
    launchctl unload "$PLIST" 2>/dev/null
    sleep 1
    launchctl load "$PLIST" 2>/dev/null
    echo "Glance service restarted."
    ;;
  status)
    if launchctl list | grep -q "com.glance.dock"; then
      echo "Status: Running"
      launchctl list | grep "com.glance.dock"
    else
      echo "Status: Stopped"
    fi
    ;;
  logs)
    tail -n 30 -f "$LOG"
    ;;
  uninstall|delete)
    launchctl unload "$PLIST" 2>/dev/null
    rm -f "$PLIST"
    echo "Glance LaunchAgent service uninstalled and removed."
    ;;
  *)
    echo "Usage: $0 {start|stop|restart|status|logs|uninstall}"
    exit 1
    ;;
esac
