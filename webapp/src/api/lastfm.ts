import api from '@/api/api';

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
    async getCurrentLastFmSession(): Promise<LastFmSessionRs | null> {
        const response = await fetch(`/api/lastfm/auth`, { method: "GET" });
        return await api.handleNullableResponse(response);
    },
    async createLastFmAuthLink(): Promise<LastFmAuthLinkRs> {
        const response = await fetch(`/api/lastfm/create-auth-link`, { method: "POST" });
        return await api.handleRequiredResponse(response);
    },
    async createLastFmSession(token: string): Promise<LastFmSessionRs> {
        const request: CreateLastFmSessionRq = {
            Token: token,
        };
        const response = await fetch(`/api/lastfm/auth`, { method: "POST", body: JSON.stringify(request) });
        return await api.handleRequiredResponse(response);
    },
    async deleteLastFmSession(): Promise<void> {
        const response = await fetch(`/api/lastfm/auth`, { method: "DELETE" });
        await api.handleRequiredResponse(response);
    },
    async updateLastFmSessionSettings(rq: LastFmSessionSettingsRq): Promise<LastFmSessionRs> {
        const response = await fetch(`/api/lastfm/settings`, { method: "PUT", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
}
