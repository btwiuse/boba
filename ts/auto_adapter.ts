/**
 * Auto-detecting adapter that tries WebTransport first, then falls back to WebSocket.
 */
import { BobaAdapter, BobaConnectionState } from './adapter.js';
import { BobaProtocolAdapter, type WebSocketAdapterCallbacks } from './websocket_adapter.js';
import { BobaWebTransportAdapter } from './webtransport_adapter.js';

export class BobaAutoAdapter implements BobaAdapter {
    private adapter: BobaAdapter | null = null;
    private onDataCallback: ((data: string | Uint8Array) => void) | null = null;
    private onStateChangeCallback: ((state: BobaConnectionState, message: string) => void) | null = null;

    constructor(
        private wsUrl: string,
        private wtUrl: string | null,
        private certHashUrl: string | null,
        private callbacks: WebSocketAdapterCallbacks = {},
    ) {}

    bobaRead(): string | Uint8Array | null {
        return this.adapter?.bobaRead() ?? null;
    }

    bobaWrite(data: string | Uint8Array): void {
        this.adapter?.bobaWrite(data);
    }

    bobaResize(cols: number, rows: number, widthPx?: number, heightPx?: number): void {
        this.adapter?.bobaResize(cols, rows, widthPx, heightPx);
    }

    connect(
        onData: (data: string | Uint8Array) => void,
        onStateChange: (state: BobaConnectionState, message: string) => void
    ): void {
        this.onDataCallback = onData;
        this.onStateChangeCallback = onStateChange;
        this._tryConnect();
    }

    private async _tryConnect(): Promise<void> {
        // Try WebTransport first if URL and cert hash endpoint are available
        if (this.wtUrl && this.certHashUrl && typeof WebTransport !== 'undefined') {
            try {
                const resp = await fetch(this.certHashUrl);
                if (resp.ok) {
                    const { hash } = await resp.json();
                    const wt = new BobaWebTransportAdapter(this.wtUrl, hash, this.callbacks);
                    this.adapter = wt;
                    await wt.connect(this.onDataCallback!, this.onStateChangeCallback!);
                    return; // WebTransport connected successfully
                }
            } catch {
                // WebTransport failed — fall through to WebSocket
            }
        }

        // Fall back to WebSocket
        const ws = new BobaProtocolAdapter(this.wsUrl, this.callbacks);
        this.adapter = ws;
        ws.connect(this.onDataCallback!, this.onStateChangeCallback!);
    }

    disconnect(): void {
        this.adapter?.disconnect();
        this.adapter = null;
    }
}
