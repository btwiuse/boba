import { BobaProtocolAdapter } from './websocket_adapter.js';
import { BobaWebTransportAdapter } from './webtransport_adapter.js';
export class BobaAutoAdapter {
    constructor(wsUrl, wtUrl, certHashUrl, callbacks = {}) {
        this.wsUrl = wsUrl;
        this.wtUrl = wtUrl;
        this.certHashUrl = certHashUrl;
        this.callbacks = callbacks;
        this.adapter = null;
        this.onDataCallback = null;
        this.onStateChangeCallback = null;
    }
    bobaRead() {
        return this.adapter?.bobaRead() ?? null;
    }
    bobaWrite(data) {
        this.adapter?.bobaWrite(data);
    }
    bobaResize(cols, rows, widthPx, heightPx) {
        this.adapter?.bobaResize(cols, rows, widthPx, heightPx);
    }
    connect(onData, onStateChange) {
        this.onDataCallback = onData;
        this.onStateChangeCallback = onStateChange;
        this._tryConnect();
    }
    async _tryConnect() {
        // Try WebTransport first if URL and cert hash endpoint are available
        if (this.wtUrl && this.certHashUrl && typeof WebTransport !== 'undefined') {
            try {
                const resp = await fetch(this.certHashUrl);
                if (resp.ok) {
                    const { hash } = await resp.json();
                    const wt = new BobaWebTransportAdapter(this.wtUrl, hash, this.callbacks);
                    this.adapter = wt;
                    await wt.connect(this.onDataCallback, this.onStateChangeCallback);
                    return; // WebTransport connected successfully
                }
            }
            catch {
                // WebTransport failed — fall through to WebSocket
            }
        }
        // Fall back to WebSocket
        const ws = new BobaProtocolAdapter(this.wsUrl, this.callbacks);
        this.adapter = ws;
        ws.connect(this.onDataCallback, this.onStateChangeCallback);
    }
    disconnect() {
        this.adapter?.disconnect();
        this.adapter = null;
    }
}
//# sourceMappingURL=auto_adapter.js.map