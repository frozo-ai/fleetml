# FleetML Demo — VHS Recording Script
# =====================================
# Records an animated GIF using https://github.com/charmbracelet/vhs
#
# Install: brew install vhs (or go install github.com/charmbracelet/vhs@latest)
# Record:  vhs scripts/demo.tape
# Output:  demo.gif (suitable for README)

Output demo.gif
Set FontSize 14
Set Width 1200
Set Height 800
Set Theme "Dracula"
Set Padding 20

# Startup
Type "# FleetML — Edge MLOps Platform"
Enter
Sleep 1s

# Version check
Type "fleetml version"
Enter
Sleep 2s

# Health check
Type "curl -s http://localhost:8080/api/v1/health | jq ."
Enter
Sleep 2s

# Deploy a model
Type "fleetml deploy --model face-detector --version v1.0 --target fleet --target-id production --policy rolling"
Enter
Sleep 3s

# Check status
Type "fleetml status"
Enter
Sleep 3s

# Canary deployment
Type "# Now let's do a canary deployment of v2.0..."
Enter
Sleep 1s
Type "fleetml deploy --model face-detector --version v2.0 --target fleet --target-id production --policy canary"
Enter
Sleep 3s

# Monitor
Type "fleetml status"
Enter
Sleep 3s

# Rollback
Type "# Something went wrong — rollback!"
Enter
Sleep 1s
Type "fleetml rollback --deployment-id <deployment-id>"
Enter
Sleep 3s

# Final status
Type "fleetml status"
Enter
Sleep 3s

Type "# Done! Zero-downtime deployment with automatic rollback."
Enter
Sleep 2s
