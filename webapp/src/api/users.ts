import api from '@/api/api';

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

export default {
    async getUsers(): Promise<UserRs[]> {
        const response = await fetch(`/api/users`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
    async createUser(rq: UserRq): Promise<UserRs> {
        const response = await fetch(`/api/users`, { method: "POST", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
    async getCurrentUser(): Promise<UserRs | null> {
        const response = await fetch(`/api/users/me`, { method: "GET" });
        return await api.handleNullableResponse(response);
    },
    async putRootUser(rq: UserRq): Promise<UserRs> {
        const response = await fetch(`/api/users/root`, { method: "PUT", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
    async patchUser(id: string, rq: UserRq): Promise<UserRs> {
        const response = await fetch(`/api/users/${id}`, { method: "PATCH", body: JSON.stringify(rq) });
        return await api.handleRequiredResponse(response);
    },
    async postUserApiKeys(id: string): Promise<UserRs> {
        const response = await fetch(`/api/users/${id}/api-keys`, { method: "POST" });
        return await api.handleRequiredResponse(response);
    },
}
