import Image from "next/image";
import { LuMoreVertical } from "react-icons/lu";
import { cn } from "../../utils/cn";
import { User } from "../../types/user";
import { Button } from "../ui/button";

const LivestreamPreviewDetailView = ({
    title,
    category,
    tags,
    user
}: {
    title: string;
    category: string | undefined;
    tags: string[];
    user: User;
}) => {
    return (
        <div className="flex flex-row gap-2">
            <Image
                width={500}
                height={500}
                className={cn(
                    "h-8 w-8 rounded-full overflow-hidden cursor-pointer"
                )}
                src={user.profilePicture ?? "https://github.com/shadcn.png"}
                alt="user avatar"
            />
            <div className="flex-1 flex-col space-y-1">
                <div className="w-full flex flex-row items-center justify-between font-semibold">
                    <span className="text-sm hover:text-primary cursor-pointer">
                        {title}
                    </span>

                    <Button><LuMoreVertical className="w-4 h-4" /></Button>
                </div>
                <div className="text-sm text-secondaryWord cursor-pointer">
                    {user.displayName ?? user.username}
                </div>
                <div className="text-sm text-secondaryWord hover:text-primary cursor-pointer">
                    {category ? category : null}
                </div>
                <div className="flex flex-row gap-2 justify-self-end">
                    {tags.map((tag, idx) => {
                        return <Button key={idx} content={tag} />;
                    })}
                </div>
            </div>
        </div>
    );
};

export default LivestreamPreviewDetailView;