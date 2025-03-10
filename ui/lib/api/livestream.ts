import { FetchError } from "../../types/fetch-error";
import { LivestreamingPreview } from "../../types/livestreaming";
import { User } from "../../types/user";
import { fetchClient } from "../../utils/fetchClient";

/**
 * Fetches the list of online users.
 */
export async function GetLivestreamings(page: number = 0): Promise<{
    livestreamings: LivestreamingPreview[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<LivestreamingPreview[]>(`/livestreamings?page=${page}`);
        return { livestreamings: data };
    } catch (error) {
        return { livestreamings: [], fetchError: error as FetchError };
    }
}