import api from '@/api/api';

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

export default {
    async getRemotes(): Promise<RemoteListRs[]> {
        const response = await fetch(`/api/remotes`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async getRemote(id: string): Promise<RemoteFullRs | null> {
        const response = await fetch(`/api/remotes/${id}`, { method: "GET" });
        return await api.handleNullableResponse(response);
    },
    async createRemote(rq: RemoteRq): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes`, { method: "POST", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
    async updateRemote(id: string, rq: RemoteRq): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes/${id}`, { method: "PUT", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
    async deleteRemote(id: string): Promise<void> {
        const response = await fetch(`/api/remotes/${id}`, { method: "DELETE" });
        await api.handleRequiredResponse(response);
    },
    async authenticateRemote(id: string, credentials: any): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes/${id}/auth`, { method: "PUT", body: JSON.stringify(credentials) });
        return await api.handleRequiredResponse(response);
    },
    async deauthenticateRemote(id: string): Promise<RemoteFullRs> {
        const response = await fetch(`/api/remotes/${id}/auth`, { method: "DELETE" });
        return await api.handleRequiredResponse(response);
    },
}
