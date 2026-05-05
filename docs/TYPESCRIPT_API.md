# `BobaTerminal` TypeScript API

The `BobaTerminal` class wraps ghostty-web's Terminal and provides a high-level API for embedding BubbleTea programs in a web page.

## Quick Start

Install the package:

```sh
npm install @btwiuse/boba
```

```javascript
import { BobaTerminal } from '@btwiuse/boba';
// or, if using the files directly from the server's embedded assets:
// import { BobaTerminal } from './boba/boba.js';

const boba = new BobaTerminal('terminal-container', {
    cols: 80,
    rows: 24,
    fontSize: 14,
    scrollback: 1000,
    allowOSC52: true, // Enable OSC 52 clipboard access
    cursorBlink: true,
    cursorStyle: 'block',
    theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4',
    },
});

await boba.init();
boba.connectWebSocket('ws://localhost:8080/ws');
boba.focus();
```

## Terminal Options

All [ghostty-web ITerminalOptions](https://github.com/coder/ghostty-web) are supported: `fontSize`, `fontFamily`, `cols`, `rows`, `cursorBlink`, `cursorStyle`, `scrollback`, `allowOSC52`, `allowTransparency`, `convertEol`, `disableStdin`, `smoothScrollDuration`, and a full `theme` with 16-color palette and cursor/selection colors.

## Selection & Clipboard

```javascript
boba.getSelection()        // Get selected text
boba.hasSelection()        // Check if text is selected
boba.copySelection()       // Copy to clipboard
boba.selectAll()           // Select all text
boba.clearSelection()      // Clear selection
boba.select(col, row, len) // Select at position
boba.selectLines(start, end)
boba.getSelectionPosition() // Get selection range
```

## Scrollback & Viewport

```javascript
boba.scrollLines(amount)  // Scroll by lines
boba.scrollPages(amount)  // Scroll by pages
boba.scrollToTop()        // Scroll to top of history
boba.scrollToBottom()     // Scroll to current output
boba.scrollToLine(line)   // Scroll to specific line
boba.getViewportY()       // Get current scroll position
```

## Terminal Control

```javascript
boba.paste(data)          // Paste with bracketed paste support
boba.input(data)          // Input as if typed
boba.focus()              // Focus terminal
boba.blur()               // Remove focus
boba.clear()              // Clear screen
boba.reset()              // Reset terminal state
boba.write(data)          // Write to display
boba.writeln(data)        // Write with newline
```

## Terminal Mode Queries

```javascript
boba.hasMouseTracking()    // Is mouse tracking enabled?
boba.hasBracketedPaste()   // Is bracketed paste enabled?
boba.hasFocusEvents()      // Are focus events enabled?
boba.getMode(mode, isAnsi) // Query arbitrary terminal mode
```

## Events

```javascript
boba.onStatusChange = (state, message) => { /* connection state */ };
boba.onTitleChange = (title) => { /* program set window title */ };
boba.onBell = () => { /* bell/beep fired */ };
boba.onSelectionChange = () => { /* selection changed */ };
boba.onKey = (event) => { /* key pressed */ };
boba.onScroll = (viewportY) => { /* viewport scrolled */ };
boba.onRender = ({ start, end }) => { /* rows rendered */ };
boba.onCursorMove = () => { /* cursor moved */ };
```

## Link Detection

```javascript
const disposable = boba.registerLinkProvider({
    provideLinks(y, callback) {
        // Detect links on row y, call callback with results
        callback(links);
    }
});
disposable.dispose(); // Unregister when done
```

## Custom Event Handlers

```javascript
boba.attachCustomKeyEventHandler((event) => {
    // Return true to prevent default handling
    return false;
});

boba.attachCustomWheelEventHandler((event) => {
    // Return true to prevent default scroll handling
    return false;
});
```

## Lifecycle

```javascript
boba.dispose()   // Clean up all resources
boba.terminal    // Access underlying ghostty-web Terminal
boba.cols        // Current column count
boba.rows        // Current row count
```

## Types

All types are exported for TypeScript consumers:

```typescript
import type {
    BobaTerminalOptions,
    BobaTheme,
    BoobaBufferRange,
    BoobaKeyEvent,
    BoobaRenderEvent,
    BoobaLinkProvider,
    BoobaLink,
    BobaAdapter,
    BobaConnectionState,
} from './boba/boba.js';
```

For adapter usage (WebSocket, WASM, custom), see [ADAPTER_USAGE.md](../ADAPTER_USAGE.md).
