import { FetchError } from "../../types/fetch-error";
import { UserVOD } from "../../types/user";
import { fetchClient } from "../../utils/fetchClient";


/**
 * Fetches the list of online users.
 */
export async function GetVODInformation(vodId: string): Promise<{
    vodInfo?: UserVOD;
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<UserVOD>(`/livestream/${vodId}`);
        return { vodInfo: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}