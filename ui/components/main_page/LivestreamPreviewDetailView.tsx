import Image from "next/image";
import { LuMoreVertical } from "react-icons/lu";
import { cn } from "../../utils/cn";
import { User } from "../../types/user";
import { Button } from "../ui/button";
import { LivestreamingPreview } from "../../types/livestreaming";

const LivestreamPreviewDetailView = ({
    livestream
}: {
    livestream: LivestreamingPreview;
}) => {
    return (
        <div className="flex flex-row gap-2">
            <Image
                width={50}
                height={50}
                className="h-12 w-12 rounded-full"
                src={livestream.userProfilePicture ?? "https://github.com/shadcn.png"}
                alt="user avatar"
            />
            <div className="w-full flex flex-row items-start justify-between">
                <div className="w-full flex flex-col items-start justify-between">
                    <p className="text-lg hover:text-primary cursor-pointer font-semibold">
                        {livestream.title}
                    </p>
                    <p className="text-xs">
                        {livestream.displayName ?? livestream.username}
                    </p>
                    {/* <p className="text-sm text-secondaryWord hover:text-primary cursor-pointer">
                        Dep trai // FOR CATEGORY
                    </p> */}
                </div>
                {/* <Button>
                    <LuMoreVertical className="w-4 h-4" />
                </Button> */}
            </div>
        </div>
    );
};

export default LivestreamPreviewDetailView;
