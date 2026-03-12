import api from '@/api/api';

export interface ListThumbnailRs {
    Id: string;
}

export default {
    async searchThumbnails(sourceIds: string[]): Promise<ListThumbnailRs[]> {
        const params = new URLSearchParams();
        for (const sourceId of sourceIds) {
            params.append("sourceId", sourceId);
        }
        const response = await fetch(`/api/thumbnails?${params}`, { method: "GET" });
        return await api.handleRequiredResponse(response);
    },
}
