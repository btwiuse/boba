## boba

Wrap a local CLI command and serve it through a browser terminal

### Synopsis

boba wraps any local CLI program and serves it in the browser
through the same embedded Ghostty terminal stack that the boba library
uses. Everything after -- is treated as the wrapped command and its
arguments.

Examples:
  boba --listen 127.0.0.1:8080 -- htop
  boba --listen 127.0.0.1:8080 -- bash
  boba --listen 127.0.0.1:8080 --origin https://app.example.com -- vim README.md

```
boba [flags] -- <command> [args...]
```

### Options

```
      --cert-file string        TLS certificate file path for HTTPS/WSS/WebTransport
      --debug                   verbose logging
  -h, --help                    help for boba
      --http3-port int          HTTP/3 WebTransport port (default: same as --listen, -1 to disable)
      --idle-timeout duration   close idle HTTP/WebSocket sessions after this duration (0 disables)
      --key-file string         TLS key file path for HTTPS/WSS/WebTransport
      --listen string           start the web server on this address (e.g. 127.0.0.1:8080) (default "127.0.0.1:8080")
      --origin string           comma-separated additional allowed browser origins (path.Match shell globs, e.g. '*.example.com' or 'https://app.example.com'; NOT regex)
      --password string         Basic Auth password (prefer --password-file or $BOBA_PASSWORD to keep secrets off argv)
      --password-file string    path to a file containing the Basic Auth password (trailing whitespace is trimmed)
      --read-only               disable client input
      --username string         Basic Auth username
```

### SEE ALSO

* [boba completion](boba_completion.md)	 - Generate the autocompletion script for the specified shell

