import api from "@/api/api";

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

export default {
    async getCurrentListenBrainzSession(): Promise<ListenBrainzSessionRs | null> {
        const response = await fetch(`/api/listenbrainz/auth`, { method: "GET" });
        return await api.handleNullableResponse(response);
    },
    async createListenBrainzSession(token: string): Promise<ListenBrainzSessionRs> {
        const request: CreateListenBrainzSessionRq = {
            Token: token,
        };
        const response = await fetch(`/api/listenbrainz/auth`, { method: "POST", body: JSON.stringify(request) });
        return await api.handleRequiredResponse(response);
    },
    async deleteListenBrainzSession(): Promise<void> {
        const response = await fetch(`/api/listenbrainz/auth`, { method: "DELETE" });
        await api.handleRequiredResponse(response);
    },
    async updateListenBrainzSessionSettings(rq: ListenBrainzSessionSettingsRq): Promise<ListenBrainzSessionRs> {
        const response = await fetch(`/api/listenbrainz/settings`, { method: "PUT", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
}
