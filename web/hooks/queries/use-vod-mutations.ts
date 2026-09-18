import { useMutation, useQueryClient } from "@tanstack/react-query";
import { DeleteVOD, UpdateVOD } from "@/lib/api/vod";
import { UploadFile } from "@/lib/api/utils";
import { ApiError, unwrapResponse } from "@/lib/api/api-error";
import { AUTHOR_VODS_QUERY_KEY, vodQueryKey } from "./use-vods";

export type UpdateVodInput = {
    vodId: string;
    title: string;
    description: string;
    isPublic: boolean;
    /** New thumbnail to upload before saving; omitted to keep the current one. */
    image?: File;
};

/**
 * Saves a VOD's metadata, uploading a new thumbnail first when one was picked.
 * The upload belongs inside the mutation because a save that uploads an image
 * and then fails must not count as a success.
 */
export function useUpdateVod() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (input: UpdateVodInput) => {
            let thumbnailPath: string | undefined;

            if (input.image) {
                const res = await UploadFile(input.image);
                thumbnailPath = unwrapResponse(res)?.newPath;
                if (!thumbnailPath) {
                    throw new ApiError({
                        ...res,
                        key: "res_err_invalid_payload",
                    });
                }
            }

            return unwrapResponse(
                await UpdateVOD(
                    input.vodId,
                    input.title,
                    input.description,
                    input.isPublic ? "public" : "private",
                    thumbnailPath,
                ),
            );
        },
        onSuccess: (_data, input) => {
            queryClient.invalidateQueries({ queryKey: AUTHOR_VODS_QUERY_KEY });
            queryClient.invalidateQueries({
                queryKey: vodQueryKey(input.vodId),
            });
        },
    });
}

export function useDeleteVod() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async (vodId: string) =>
            unwrapResponse(await DeleteVOD(vodId)),
        onSuccess: (_data, vodId) => {
            queryClient.invalidateQueries({ queryKey: AUTHOR_VODS_QUERY_KEY });
            queryClient.removeQueries({ queryKey: vodQueryKey(vodId) });
        },
    });
}
