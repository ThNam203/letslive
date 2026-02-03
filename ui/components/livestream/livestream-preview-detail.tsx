import { Livestream } from "../../types/livestream";
import { PublicUser } from "../../types/user";
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
    user: PublicUser | null;
}) => {
    const { t } = useT("common");

    return (
        <div className="flex items-start gap-3">
            <div className="h-10 w-10 flex-shrink-0 overflow-hidden rounded-full bg-muted">
                <Avatar className="border border-border">
                    <AvatarImage
                        src={user?.profilePicture}
                        alt={`${user?.username} avatar`}
                    />
                    <AvatarFallback>
                        {user?.username.charAt(0).toUpperCase()}
                    </AvatarFallback>
                </Avatar>
            </div>
            <div className="min-w-0 flex-1">
                <h3 className="truncate text-base font-semibold">
                    {livestream.title}
                </h3>
                <p className="text-muted-foreground truncate text-sm">
                    {user ? (user.displayName ?? user.username) : "Unknown"}
                </p>
                <div className="text-muted-foreground mt-1 flex items-center gap-3 text-xs">
                    <div className="flex items-center gap-1">
                        <IconEye className="h-3 w-3" />
                        <span>
                            {livestream.viewCount}{" "}
                            {livestream.viewCount < 2 ? "view" : "views"}
                        </span>
                    </div>
                    <div className="flex items-center gap-1">
                        <IconClock className="h-3 w-3" />
                        <span>
                            {t("started_at", {
                                time: dateDiffFromNow(livestream.startedAt, t),
                            })}
                        </span>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default LivestreamPreviewDetailView;
