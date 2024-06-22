export interface ImportQueueItem {
    Id: string;
    Url: string;
}

export interface Tape {
    Id: string;
    Name: string;
    AuthorName: string;
    ThumbnailPath: string;
    ReleaseDate: string | null;
    Files: TapeFile[];
}

export interface TapeFile {
    Id: string;
    Name: string;
    AuthorName: string;
    ThumbnailPath: string;
    ReleaseDate: string | null;
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

export interface Playlist {
    Id: string;
    Name: string;
    ThumbnailPath: string;
    Tracks: PlaylistTrack[];
}

export interface PlaylistTrack {
    Id: string;

    TapeTrackId: string;
    TapeTrack: TapeTrack;
}

export interface Album {
    Id: string;

    Name: string;
    Artist: string;
    ReleaseDate: string | null;

    ThumbnailPath: string;

    Tracks: AlbumTrack[];
}

export interface AlbumTrack {
    Id: string;

    TapeTrackId: string;
    TapeTrack: TapeTrack;
}

export interface RelatedItems {
    Tapes: Tape[];
    Playlists: Playlist[];
    Albums: Album[];
}

export default {
    async getImportQueue(): Promise<ImportQueueItem[]> {
        const response = await fetch(`/api/import-queue`, { method: "GET" });
        return await response.json();
    },
    async addToImportQueue(url: string): Promise<ImportQueueItem> {
        const response = await fetch(
            "/api/import-queue?" + new URLSearchParams({ url }),
            { method: "POST" },
        );
        return await response.json();
    },
    async deleteFromImportQueue(id: string) {
        const response = await fetch(`/api/import-queue/${id}`, { method: "DELETE" });
        if (!response.ok) {
            throw await response.json();
        }
    },

    async getAllTapes(): Promise<Tape[]> {
        const response = await fetch(`/api/tapes`, { method: "GET" });
        return await response.json();
    },
    async getTape(id: string): Promise<Tape> {
        const response = await fetch(`/api/tapes/${id}`, { method: "GET" });
        return await response.json();
    },
    async getTapeRelationships(id: string): Promise<RelatedItems> {
        const response = await fetch(`/api/tapes/${id}/related`, { method: "GET" });
        return await response.json();
    },
    async saveTape(id: string, tape: Tape) {
        const response = await fetch(`/api/tapes/${id}`, { method: "PUT", body: JSON.stringify(tape) });
        if (!response.ok) {
            throw await response.json();
        }
    },

    async createPlaylist(playlist: Playlist): Promise<Playlist> {
        const response = await fetch(`/api/playlists`, { method: "POST", body: JSON.stringify(playlist) });
        if (!response.ok) {
            throw await response.json();
        } else {
            return await response.json();
        }
    },
    async deletePlaylist(id: string) {
        const response = await fetch(`/api/playlists/${id}`, { method: "DELETE" });
        if (!response.ok) {
            throw await response.json();
        }
    },
    async getAllPlaylists(): Promise<Playlist[]> {
        const response = await fetch(`/api/playlists`, { method: "GET" });
        return await response.json();
    },
    async getPlaylist(id: string): Promise<Playlist> {
        const response = await fetch(`/api/playlists/${id}`, { method: "GET" });
        return await response.json();
    },
    async getPlaylistRelationships(id: string): Promise<RelatedItems> {
        const response = await fetch(`/api/playlists/${id}/related`, { method: "GET" });
        return await response.json();
    },

    async createAlbum(album: Album): Promise<Album> {
        const response = await fetch(`/api/albums`, { method: "POST", body: JSON.stringify(album) });
        if (!response.ok) {
            throw await response.json();
        } else {
            return await response.json();
        }
    },
    async updateAlbum(id: string, album: Album): Promise<Album> {
        const response = await fetch(`/api/albums/${id}`, { method: "PUT", body: JSON.stringify(album) });
        if (!response.ok) {
            throw await response.json();
        } else {
            return await response.json();
        }
    },
    async deleteAlbum(id: string) {
        const response = await fetch(`/api/albums/${id}`, { method: "DELETE" });
        if (!response.ok) {
            throw await response.json();
        }
    },
    async getAllAlbums(): Promise<Album[]> {
        const response = await fetch(`/api/albums`, { method: "GET" });
        return await response.json();
    },
    async getAlbum(id: string): Promise<Album> {
        const response = await fetch(`/api/albums/${id}`, { method: "GET" });
        return await response.json();
    },
    async getAlbumRelationships(id: string): Promise<RelatedItems> {
        const response = await fetch(`/api/albums/${id}/related`, { method: "GET" });
        return await response.json();
    },
}
