export interface ImportResult {
    Ok: boolean;
    Error: string | null;
}

export interface Tape {
    Id: string;
    Name: string;
    AuthorName: string;
    Tracks: TapeTrack[];
}

export interface TapeTrack {
    Id: string;

    RawStartOffsetMs: number;
    StartOffsetMs: number;
    RawEndOffsetMs: number;
    EndOffsetMs: number;

    Artist: string;
    Title: string;
}

export default {
    async import(url: string, format: string): Promise<ImportResult> {
        const response = await fetch(
            "/api/import?" + new URLSearchParams({ url, format }),
            { method: "POST" },
        );

        return {
            Ok: response.ok,
            Error: response.ok ? null : `${response.status} ${response.statusText}`,
        };
    },
    async getAllTapes(): Promise<Tape[]> {
        const response = await fetch(`/api/tapes`, { method: "GET" });
        return await response.json();
    },
    async getTape(id: string): Promise<Tape> {
        const response = await fetch(`/api/tapes/${id}`, { method: "GET" });
        return await response.json();
    },
    async saveTape(id: string, tape: Tape) {
        const response = await fetch(`/api/tapes/${id}`, { method: "PUT", body: JSON.stringify(tape) });
        if (!response.ok) {
            throw await response.json();
        }
    },
}
