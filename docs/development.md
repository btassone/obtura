# Development Guide

## Hot Module Reloading

Obtura uses Air for hot module reloading during development. When you run `make dev` or `air`, the following happens:

1. **Air Proxy**: Air starts a proxy server on port 3000 that forwards requests to your application on port 8080
2. **File Watching**: Air watches for changes in `.go`, `.templ`, `.css`, `.js`, and `.html` files
3. **Auto Rebuild**: When files change, Air automatically:
   - Regenerates Templ files
   - Rebuilds the Go application
   - Restarts the server
4. **Browser Reload**: The browser automatically reloads when changes are detected

### Access URLs

- **Development (with hot reload)**: http://localhost:3000
- **Direct application**: http://localhost:8080

### Configuration

The hot reload behavior is configured in `.air.toml`:

```toml
[proxy]
  app_port = 8080      # Your application port
  enabled = true       # Enable proxy for hot reload
  proxy_port = 3000    # Port you access in browser
```

### Troubleshooting

If hot reload isn't working:

1. Make sure you're accessing http://localhost:3000 (not 8080)
2. Check that Air is running: `ps aux | grep air`
3. Check Air logs for errors
4. Ensure WebSocket connection is established (check browser console)

### Manual Reload

If automatic reload fails, you can:
- Manually refresh the browser
- Restart Air with `make dev`
- Check `tmp/build-errors.log` for compilation errors