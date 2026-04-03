import api from '@/api/api';

export enum TapeType {
    Album = "album",
    Playlist = "playlist",
}

export interface TapeRsArtist {
    Id: string;
    Name: string;
}

export interface TapeRsTrack {
    Id: string;

    SourceId: string | null;
    RemoteId: string | null;

    Artist: TapeRsArtist | null;
    Title: string;

    ThumbnailId: string | null;
}

export interface TapeListRs {
    Id: string;
    Name: string;
    Type: TapeType;

    ThumbnailId: string | null;

    Artist: TapeRsArtist | null;
    ReleasedAt: string | null;

    CreatedAt: string;
}

export interface TapeFullRs {
    Id: string;
    Name: string;
    Type: TapeType;

    ThumbnailId: string | null;

    Artist: TapeRsArtist | null;
    ReleasedAt: string | null;

    Tracks: TapeRsTrack[];
}

export interface TapeRq {
    Name: string;
    Type: string;

    ThumbnailId: string | null;

    ArtistId: string | null;
    ReleasedAt: string | null;

    TrackIds: string[];
}

export interface GuessTapeMetadataRq {
    TrackIds: string[];
}

export interface GuessTapeMetadataRs {
    Name: string;
    Type: TapeType;

    Artist: TapeRsArtist | null;
    ReleasedAt: string | null;
    ThumbnailId: string | null;
}

export default {
    async createTape(tape: TapeRq): Promise<TapeFullRs> {
        const response = await fetch(`/api/tapes`, { method: "POST", body: JSON.stringify(tape) });
        return await api.handleRequiredResponse(response);
    },
    async updateTape(id: string, tape: TapeRq): Promise<TapeFullRs> {
        const response = await fetch(`/api/tapes/${id}`, { method: "PUT", body: JSON.stringify(tape) });
        return await api.handleRequiredResponse(response);
    },
    async deleteTape(id: string): Promise<void> {
        const response = await fetch(`/api/tapes/${id}`, { method: "DELETE" });
        await api.handleRequiredResponse(response);
    },
    async listTapes(): Promise<TapeListRs[]> {
        const response = await fetch(`/api/tapes`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async getTape(id: string): Promise<TapeFullRs> {
        const response = await fetch(`/api/tapes/${id}`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async guessTapeMetadata(rq: GuessTapeMetadataRq): Promise<GuessTapeMetadataRs> {
        const response = await fetch(`/api/tapes/guess-metadata`, { method: "POST", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
}
