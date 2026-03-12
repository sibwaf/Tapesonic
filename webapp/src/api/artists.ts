import api from "@/api/api";

export interface ArtistListRs {
    Id: string;
    Name: string;
}

export interface ArtistFullRs {
    Id: string;
    Name: string;
    Aliases: string[];
    MusicBrainzId: string | null;
}

export interface ArtistRq {
    Name: string;
    Aliases: string[];
    MusicBrainzId: string | null;
}

export interface MergeArtistsRq {
    Ids: string[];
}

export default {
    async mergeArtists(rq: MergeArtistsRq): Promise<ArtistFullRs> {
        const response = await fetch(`/api/artists/merge`, { method: "POST", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
    async listArtists(query: string, count: number, offset: number): Promise<ArtistListRs[]> {
        const params = new URLSearchParams();
        params.append("q", query);
        params.append("count", count.toString());
        params.append("offset", offset.toString());

        const response = await fetch(`/api/artists?${params}`, { method: "GET" });
        return await response.json();
    },
    async createArtist(rq: ArtistRq): Promise<ArtistFullRs> {
        const response = await fetch(`/api/artists`, { method: "POST", body: JSON.stringify(rq) });
        return await response.json();
    },
    async getArtistById(id: string): Promise<ArtistFullRs> {
        const response = await fetch(`/api/artists/${id}`, { method: "GET" });
        return await response.json();
    },
    async putArtistById(id: string, rq: ArtistRq): Promise<ArtistFullRs> {
        const response = await fetch(`/api/artists/${id}`, { method: "PUT", body: JSON.stringify(rq) });
        return await response.json();
    }
}
