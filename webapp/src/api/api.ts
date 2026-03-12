export default {
    async handleNullableResponse<T>(response: Response): Promise<T | null> {
        if (response.status == 404) {
            return null;
        }

        await this.throwOnHttpError(response);
        return await response.json();
    },
    async handleRequiredResponse<T>(response: Response): Promise<T> {
        await this.throwOnHttpError(response);
        return await response.json();
    },

    async throwOnHttpError(response: Response) {
        if (response.status != 200) {
            const body = await response.text();
            let err = `HTTP ${response.status}`;
            if (body.length > 0) {
                err += `: ${body}`;
            }
            throw err;
        }
    }
}
