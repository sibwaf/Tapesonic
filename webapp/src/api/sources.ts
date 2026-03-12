import api from '@/api/api';

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

export interface SourceTrackRsArtist {
    Id: string;
    Name: string;
}

export interface SourceTrackRs {
    Id: string;
    SourceId: string;

    Artist: SourceTrackRsArtist | null;
    Title: string;

    StartOffsetMs: number;
    EndOffsetMs: number;
}

export interface SourceTrackRq {
    Id: string | null;
    ArtistId: string | null;
    Title: string;
    StartOffsetMs: number;
    EndOffsetMs: number;
}

export default {
    async listSources(): Promise<SourceListRs[]> {
        const params = new URLSearchParams();
        params.append("managementPolicy", "MANUAL");
        const response = await fetch(`/api/sources?${params}`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async getSource(id: string): Promise<SourceFullRs> {
        const response = await fetch(`/api/sources/${id}`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async getSourceHierarchy(id: string): Promise<SourceHierarchyListRs[]> {
        const response = await fetch(`/api/sources/${id}/hierarchy`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async getSourceTracks(id: string, recursive: boolean): Promise<SourceTrackRs[]> {
        const params = new URLSearchParams({ "recursive": `${recursive}` });
        const response = await fetch(`/api/sources/${id}/tracks?${params}`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async replaceSourceTracks(id: string, tracks: SourceTrackRq[]): Promise<SourceTrackRs[]> {
        const response = await fetch(`/api/sources/${id}/tracks`, { method: "PUT", body: JSON.stringify(tracks) });
        return await api.handleRequiredResponse(response);
    },
    async deleteSourceFile(sourceId: string): Promise<void> {
        const response = await fetch(`/api/sources/${sourceId}/file`, { method: "DELETE" });
        await api.handleRequiredResponse(response);
    },
}
