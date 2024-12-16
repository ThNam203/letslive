import { CustomLink } from "@/components/Hover3DBox";
import LivestreamsPreviewView from "@/components/LivesteamsPreviewView";
import ShowToast from "@/components/ShowToast";
import { GetOnlineUsers } from "@/lib/user";

export default async function HomePage() {
    const { users, fetchError } = await GetOnlineUsers();

    return (
        <>
            {fetchError != undefined && (
                <ShowToast id={fetchError.id} err={fetchError.message} />
            )}
            <div className="flex flex-col w-full max-h-full p-8 overflow-y-scroll overflow-x-hidden">
                <h1 className="font-semibold text-lg">
                    <CustomLink content="Live channels" href="" /> we think
                    you&#39;ll like
                </h1>

                <div className="w-full flex flex-row items-center justify-between gap-4">
                    <LivestreamsPreviewView users={users} />
                </div>
            </div>
        </>
    );
}
