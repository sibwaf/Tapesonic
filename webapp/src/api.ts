export interface FullSourceRs {
    Id: string;

    Url: string;
    Title: string;
    Uploader: string;

    AlbumArtist: string;
    AlbumTitle: string;
    AlbumIndex: number;
    TrackArtist: string;
    TrackTitle: string;
    DurationMs: number;

    ReleaseDate: string | null;

    ThumbnailId: string | null;
}

export interface ListSourceRs {
    Id: string;

    Url: string;
    Title: string;
    Uploader: string;

    DurationMs: number;

    ThumbnailId: string | null;
}

export interface ListSourceHierarchyRs {
    Id: string;
    ParentId: string | null;

    Url: string;
    Title: string;
    Uploader: string;

    ListIndex: number;

    ThumbnailId: string | null;
}

export interface SourceFileRs {
    Codec: string;
}

export interface TrackRs {
    Id: string;
    SourceId: string;

    Artist: string;
    Title: string;

    StartOffsetMs: number;
    EndOffsetMs: number;
}

export enum TapeType {
    Album = "album",
    Playlist = "playlist",
}

export interface Tape {
    Id: string;

    Name: string;
    Type: TapeType;

    ThumbnailId: string | null;

    Artist: string;
    ReleasedAt: string | null;

    Tracks: TrackRs[];
}

export interface ListTape {
    Id: string;

    Name: string;
    Type: TapeType;

    ThumbnailId: string | null;

    Artist: string;
    ReleasedAt: string | null;

    CreatedAt: string;
}

export interface GetListSourceRs {
    Source: ListSourceRs;
    File: SourceFileRs | null;
}

export interface ListThumbnailRs {
    Id: string;
}

export interface LastFmSessionRs {
    Username: string;
    UpdatedAt: string;
}

export interface LastFmAuthLinkRs {
    Url: string;
    Token: string;
}

export interface CreateLastFmSessionRq {
    Token: string;
}

export default {
    async getCurrentLastFmSession(): Promise<LastFmSessionRs | null> {
        const response = await fetch(`/api/settings/lastfm/auth`, { method: "GET" });
        const body = await response.text();
        if (body == "") {
            return null;
        } else {
            return JSON.parse(body);
        }
    },
    async createLastFmAuthLink(): Promise<LastFmAuthLinkRs> {
        const response = await fetch(`/api/settings/lastfm/create-auth-link`, { method: "POST" });
        return await response.json();
    },
    async createLastFmSession(token: string): Promise<LastFmSessionRs> {
        const request: CreateLastFmSessionRq = {
            Token: token,
        };
        const response = await fetch(`/api/settings/lastfm/auth`, { method: "POST", body: JSON.stringify(request) });
        return await response.json();
    },

    async addSource(url: string): Promise<FullSourceRs> {
        const params = new URLSearchParams({ "url": url });
        const response = await fetch(`/api/sources?${params}`, { method: "POST" });
        return await response.json();
    },
    async listSources(): Promise<GetListSourceRs[]> {
        const response = await fetch(`/api/sources`, { method: "GET" });
        return await response.json();
    },
    async getSource(id: string): Promise<FullSourceRs> {
        const response = await fetch(`/api/sources/${id}`, { method: "GET" });
        return await response.json();
    },
    async getSourceHierarchy(id: string): Promise<ListSourceHierarchyRs[]> {
        const response = await fetch(`/api/sources/${id}/hierarchy`, { method: "GET" });
        return await response.json();
    },
    async getSourceTracks(id: string, recursive: boolean): Promise<TrackRs[]> {
        const params = new URLSearchParams({ "recursive": `${recursive}` });
        const response = await fetch(`/api/sources/${id}/tracks?${params}`, { method: "GET" });
        return await response.json();
    },
    async replaceSourceTracks(id: string, tracks: TrackRs[]): Promise<TrackRs[]> {
        const response = await fetch(`/api/sources/${id}/tracks`, { method: "PUT", body: JSON.stringify(tracks) });
        return await response.json();
    },
    async getSourceFile(sourceId: string): Promise<SourceFileRs | null> {
        const response = await fetch(`/api/sources/${sourceId}/file`, { method: "GET" });
        const body = await response.text();
        if (body == "") {
            return null;
        } else {
            return JSON.parse(body);
        }
    },
    async deleteSourceFile(sourceId: string): Promise<void> {
        const response = await fetch(`/api/sources/${sourceId}/file`, { method: "DELETE" });
        return await response.json();
    },

    async createTape(tape: Tape): Promise<Tape> {
        const response = await fetch(`/api/tapes`, { method: "POST", body: JSON.stringify(tape) });
        return await response.json();
    },
    async updateTape(id: string, tape: Tape): Promise<Tape> {
        const response = await fetch(`/api/tapes/${id}`, { method: "PUT", body: JSON.stringify(tape) });
        return await response.json();
    },
    async deleteTape(id: string): Promise<void> {
        await fetch(`/api/tapes/${id}`, { method: "DELETE" });
    },
    async listTapes(): Promise<ListTape[]> {
        const response = await fetch(`/api/tapes`, { method: "GET" });
        return await response.json();
    },
    async getTape(id: string): Promise<Tape> {
        const response = await fetch(`/api/tapes/${id}`, { method: "GET" });
        return await response.json();
    },
    async guessTapeMetadata(trackIds: string[]): Promise<Tape> {
        const response = await fetch(`/api/tapes/guess-metadata`, { method: "POST", body: JSON.stringify({ trackIds }) });
        return await response.json();
    },

    async searchTracks(query: string): Promise<TrackRs[]> {
        const params = new URLSearchParams({ "q": query });
        const response = await fetch(`/api/tracks?${params}`, { method: "GET" });
        return await response.json();
    },

    async searchThumbnails(sourceIds: string[]): Promise<ListThumbnailRs[]> {
        const params = new URLSearchParams();
        for (const sourceId of sourceIds) {
            params.append("sourceId", sourceId);
        }
        const response = await fetch(`/api/thumbnails?${params}`, { method: "GET" });
        return await response.json();
    },
}
