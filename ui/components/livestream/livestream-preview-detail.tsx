import { Livestream } from "../../types/livestream";
import { User } from "../../types/user";
import { dateDiffFromNow } from "@/utils/timeFormats";
import IconClock from "../icons/clock";
import IconEye from "../icons/eye";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import useT from "@/hooks/use-translation";

const LivestreamPreviewDetailView = ({
    livestream,
    user,
}: {
    livestream: Livestream;
    user: User | null;
}) => {
    const { t } = useT("common");

    return (
        <div className="flex items-start gap-3">
            <div className="h-10 w-10 rounded-full overflow-hidden bg-muted flex-shrink-0">
                <Avatar className="border border-border">
                    <AvatarImage
                        src={
                            user?.profilePicture
                        }
                        alt={`${user?.username} avatar`}
                    />
                    <AvatarFallback>
                        {user?.username.charAt(0).toUpperCase()}
                    </AvatarFallback>
                </Avatar>
            </div>
            <div className="flex-1 min-w-0">
                <h3 className="font-semibold text-base truncate">
                    {livestream.title}
                </h3>
                <p className="text-sm text-muted-foreground truncate">
                    {user ? user.displayName ?? user.username : "Unknown"}
                </p>
                <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                    <div className="flex items-center gap-1">
                        <IconEye className="h-3 w-3" />
                        <span>{livestream.viewCount} {livestream.viewCount < 2 ? "view" : "views"}</span>
                    </div>
                    <div className="flex items-center gap-1">
                        <IconClock className="h-3 w-3" />
                        <span>
                            {t('started_at', { time: dateDiffFromNow(livestream.startedAt, t) })}
                        </span>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default LivestreamPreviewDetailView;
