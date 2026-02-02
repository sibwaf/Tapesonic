export interface UserRq {
    Name: string | null;
    Password: string | null;
    Role: string | null;
}

export interface UserRs {
    Id: string;
    Name: string;
    Role: string;
    ApiKey: string;
}

export interface RemoteRq {
    Name: string;
    Type: string;
    Url: string;

    IsScrobbleReplicationEnabled: boolean;
    IsExternalScrobblingEnabled: boolean;
}

export interface RemoteListRs {
    Id: string;

    Name: string;
    Type: string;
    Url: string;

    IsScrobbleReplicationEnabled: boolean;
    IsExternalScrobblingEnabled: boolean;
}

export interface RemoteFullRs {
    Id: string;

    Type: string;
    Name: string;
    Url: string;

    IsScrobbleReplicationEnabled: boolean;
    IsExternalScrobblingEnabled: boolean;

    Username: string;

    Description: string;
    Status: string;
}

export interface SourceFullRs {
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
    File: SourceFileRs | null;
}

export interface SourceListRs {
    Id: string;

    Url: string;
    Title: string;
    Uploader: string;
    DurationMs: number;
    ThumbnailId: string | null;
    File: SourceFileRs | null;
}

export interface SourceHierarchyListRs {
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

export interface SourceTrackRs {
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

    Tracks: SourceTrackRs[];
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

export interface GuessTapeMetadataRq {
    TrackIds: string[];
}

export interface GuessTapeMetadataRs {
    Name: string;
    Type: TapeType;
    Artist: string;
    ReleasedAt: string | null;
    ThumbnailId: string | null;
}

export interface ListThumbnailRs {
    Id: string;
}

export interface ListenBrainzSessionRs {
    Username: string;
    IsScrobblingEnabled: boolean;
    UpdatedAt: string;
}

export interface CreateListenBrainzSessionRq {
    Token: string;
}

export interface ListenBrainzSessionSettingsRq {
    IsScrobblingEnabled: boolean;
}

export interface LastFmSessionRs {
    Username: string;
    IsScrobblingEnabled: boolean;
    UpdatedAt: string;
}

export interface LastFmAuthLinkRs {
    Url: string;
    Token: string;
}

export interface CreateLastFmSessionRq {
    Token: string;
}

export interface LastFmSessionSettingsRq {
    IsScrobblingEnabled: boolean;
}

export default {
    async getUsers(): Promise<UserRs[]> {
        const response = await fetch(`/api/users`, { method: "GET" });
        return await response.json();
    },
    async createUser(rq: UserRq): Promise<UserRs> {
        const response = await fetch(`/api/users`, { method: "POST", body: JSON.stringify(rq) });
        return await response.json();
    },
    async getCurrentUser(): Promise<UserRs | null> {
        const response = await fetch(`/api/users/me`, { method: "GET" });
        if (response.status == 401) {
            return null;
        }
        return await response.json();
    },
    async putRootUser(rq: UserRq): Promise<UserRs> {
        const response = await fetch(`/api/users/root`, { method: "PUT", body: JSON.stringify(rq) });
        return await response.json();
    },
    async patchUser(id: string, rq: UserRq): Promise<UserRs | null> {
        const response = await fetch(`/api/users/${id}`, { method: "PATCH", body: JSON.stringify(rq) });
        if (response.status == 404) {
            return null;
        }
        return await response.json();
    },
    async postUserApiKeys(id: string): Promise<UserRs> {
        const response = await fetch(`/api/users/${id}/api-keys`, { method: "POST" });
        return await response.json();
    },

    async getRemotes(): Promise<RemoteListRs[]> {
        const response = await fetch(`/api/remotes`, { method: "GET" });
        return await response.json();
    },
    async getRemote(id: string): Promise<RemoteFullRs | null> {
        const response = await fetch(`/api/remotes/${id}`, { method: "GET" });
        if (response.status == 404) {
            return null;
        }
        return await response.json();
    },
    async createRemote(rq: RemoteRq): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes`, { method: "POST", body: JSON.stringify(rq) });
        return await response.json();
    },
    async updateRemote(id: string, rq: RemoteRq): Promise<RemoteFullRs | null> {
        const response = await fetch(`/api/remotes/${id}`, { method: "PUT", body: JSON.stringify(rq) });
        if (response.status == 404) {
            return null;
        }
        return await response.json();
    },
    async deleteRemote(id: string): Promise<void> {
        const response = await fetch(`/api/remotes/${id}`, { method: "DELETE" });
        if (response.status != 200) {
            const body = await response.text();
            let err = `HTTP ${response.status}`;
            if (body.length > 0) {
                err += `: ${body}`;
            }
            throw err;
        }
    },
    async authenticateRemote(id: string, credentials: any): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes/${id}/auth`, { method: "PUT", body: JSON.stringify(credentials) });
        return await response.json();
    },
    async deauthenticateRemote(id: string): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes/${id}/auth`, { method: "DELETE" });
        return await response.json();
    },

    async getCurrentListenBrainzSession(): Promise<ListenBrainzSessionRs | null> {
        const response = await fetch(`/api/listenbrainz/auth`, { method: "GET" });
        if (response.status == 404) {
            return null;
        }
        return await response.json();
    },
    async createListenBrainzSession(token: string): Promise<ListenBrainzSessionRs> {
        const request: CreateListenBrainzSessionRq = {
            Token: token,
        };
        const response = await fetch(`/api/listenbrainz/auth`, { method: "POST", body: JSON.stringify(request) });
        return await response.json();
    },
    async deleteListenBrainzSession(): Promise<void> {
        const response = await fetch(`/api/listenbrainz/auth`, { method: "DELETE" });
        return await response.json();
    },
    async updateListenBrainzSessionSettings(rq: ListenBrainzSessionSettingsRq): Promise<ListenBrainzSessionRs> {
        const response = await fetch(`/api/listenbrainz/settings`, { method: "PUT", body: JSON.stringify(rq) });
        return await response.json();
    },

    async getCurrentLastFmSession(): Promise<LastFmSessionRs | null> {
        const response = await fetch(`/api/lastfm/auth`, { method: "GET" });
        if (response.status == 404) {
            return null;
        }
        return await response.json();
    },
    async createLastFmAuthLink(): Promise<LastFmAuthLinkRs> {
        const response = await fetch(`/api/lastfm/create-auth-link`, { method: "POST" });
        return await response.json();
    },
    async createLastFmSession(token: string): Promise<LastFmSessionRs> {
        const request: CreateLastFmSessionRq = {
            Token: token,
        };
        const response = await fetch(`/api/lastfm/auth`, { method: "POST", body: JSON.stringify(request) });
        return await response.json();
    },
    async deleteLastFmSession(): Promise<void> {
        const response = await fetch(`/api/lastfm/auth`, { method: "DELETE" });
        return await response.json();
    },
    async updateLastFmSessionSettings(rq: LastFmSessionSettingsRq): Promise<LastFmSessionRs> {
        const response = await fetch(`/api/lastfm/settings`, { method: "PUT", body: JSON.stringify(rq) });
        return await response.json();
    },

    async listSources(): Promise<SourceListRs[]> {
        const params = new URLSearchParams();
        params.append("managementPolicy", "MANUAL");
        const response = await fetch(`/api/sources?${params}`, { method: "GET" });
        return await response.json();
    },
    async getSource(id: string): Promise<SourceFullRs> {
        const response = await fetch(`/api/sources/${id}`, { method: "GET" });
        return await response.json();
    },
    async getSourceHierarchy(id: string): Promise<SourceHierarchyListRs[]> {
        const response = await fetch(`/api/sources/${id}/hierarchy`, { method: "GET" });
        return await response.json();
    },
    async getSourceTracks(id: string, recursive: boolean): Promise<SourceTrackRs[]> {
        const params = new URLSearchParams({ "recursive": `${recursive}` });
        const response = await fetch(`/api/sources/${id}/tracks?${params}`, { method: "GET" });
        return await response.json();
    },
    async replaceSourceTracks(id: string, tracks: SourceTrackRs[]): Promise<SourceTrackRs[]> {
        const response = await fetch(`/api/sources/${id}/tracks`, { method: "PUT", body: JSON.stringify(tracks) });
        return await response.json();
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
    async guessTapeMetadata(rq: GuessTapeMetadataRq): Promise<GuessTapeMetadataRs> {
        const response = await fetch(`/api/tapes/guess-metadata`, { method: "POST", body: JSON.stringify(rq) });
        return await response.json();
    },

    async searchTracks(query: string): Promise<SourceTrackRs[]> {
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
