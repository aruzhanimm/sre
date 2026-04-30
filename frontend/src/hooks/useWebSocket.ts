import { useEffect, useRef, useCallback } from 'react';

interface OddsUpdate {
    type: string;
    eventId: string;
    data: {
        marketId: string;
        outcome: string;
        oldOdds: number;
        newOdds: number;
        change: number;
        direction: string;
        timestamp: number;
    };
}

export function useWebSocket(eventId: string | null, onMessage: (data: OddsUpdate) => void) {
    const ws = useRef<WebSocket | null>(null);
    const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

    const connect = useCallback(() => {
        if (!eventId) return;

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${protocol}//${window.location.host}/ws?eventId=${eventId}`;

        ws.current = new WebSocket(url);

        ws.current.onopen = () => {
            console.log('WebSocket connected for event:', eventId);
        };

        ws.current.onmessage = (e) => {
            try {
                const data = JSON.parse(e.data) as OddsUpdate;
                onMessage(data);
            } catch (err) {
                console.error('WebSocket parse error:', err);
            }
        };

        ws.current.onclose = () => {
            reconnectTimer.current = setTimeout(connect, 3000);
        };

        ws.current.onerror = () => {
            ws.current?.close();
        };
    }, [eventId, onMessage]);

    useEffect(() => {
        connect();
        return () => {
            if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
            ws.current?.close();
        };
    }, [connect]);

    return ws;
}
