import api from '@/api/api';

export interface TrackRsArtist {
    Id: string;
    Name: string;
}

export interface TrackRs {
    Id: string;

    SourceId: string | null;
    RemoteId: string | null;

    Artist: TrackRsArtist | null;
    Title: string;

    ThumbnailId: string | null;
}

export default {
    async searchTracks(query: string): Promise<TrackRs[]> {
        const params = new URLSearchParams({ "q": query });
        const response = await fetch(`/api/tracks?${params}`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
}
